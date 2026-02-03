package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	agentregistryv1alpha1 "github.com/agentregistry-dev/agentregistry/api/v1alpha1"
	"github.com/agentregistry-dev/agentregistry/internal/cluster"
	"github.com/agentregistry-dev/agentregistry/internal/controller"
	"github.com/agentregistry-dev/agentregistry/internal/httpapi"

	kagentv1alpha2 "github.com/kagent-dev/kagent/go/api/v1alpha2"
	kmcpv1alpha1 "github.com/kagent-dev/kmcp/api/v1alpha1"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = agentregistryv1alpha1.AddToScheme(scheme)
	_ = kagentv1alpha2.AddToScheme(scheme)
	_ = kmcpv1alpha1.AddToScheme(scheme)
}

func main() {
	var (
		httpAPIAddr string
		skipUI      bool
	)

	flag.StringVar(&httpAPIAddr, "http-api-address", ":8080", "The address the HTTP API server binds to.")
	flag.BoolVar(&skipUI, "skip-ui", false, "Skip starting the UI dev server.")
	flag.Parse()

	// Set up logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	logf.SetLogger(zerologr.New(&log.Logger))

	log.Info().Msg("=== Agent Registry Demo ===")

	// Disable auth for demo
	os.Setenv("AGENTREGISTRY_DISABLE_AUTH", "true")

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to find project root")
	}

	// Start envtest
	log.Info().Msg("starting envtest...")
	env := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join(projectRoot, "config", "crd"),
			filepath.Join(projectRoot, "config", "external-crds"),
		},
		ErrorIfCRDPathMissing: true,
	}

	config, err := env.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start envtest")
	}
	log.Info().Str("host", config.Host).Msg("envtest started")

	// Write kubeconfig
	kubeconfigPath, err := writeKubeconfig(env)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to write kubeconfig")
	}

	// Create manager
	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                server.Options{BindAddress: "0"},
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create manager")
	}

	// Setup indexes and reconcilers
	if err := controller.SetupIndexes(mgr); err != nil {
		log.Fatal().Err(err).Msg("failed to setup indexes")
	}
	setupReconcilers(mgr, log.Logger)

	// Initialize remote client factory
	clusterFactory := cluster.NewFactory(mgr.GetClient(), log.Logger)
	controller.RemoteClientFactory = clusterFactory.CreateClientFunc()

	// Setup HTTP API
	httpServer := httpapi.NewServer(mgr.GetClient(), mgr.GetCache(), log.Logger)
	if err := mgr.Add(httpServer.Runnable(httpAPIAddr)); err != nil {
		log.Fatal().Err(err).Msg("failed to add HTTP API server")
	}

	// Start manager
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := mgr.Start(ctx); err != nil {
			log.Error().Err(err).Msg("manager error")
		}
	}()

	// Wait for cache
	time.Sleep(500 * time.Millisecond)
	if !mgr.GetCache().WaitForCacheSync(ctx) {
		log.Fatal().Msg("cache sync timeout")
	}

	// Create sample resources
	log.Info().Msg("creating sample resources...")
	if err := createSampleResources(ctx, mgr.GetClient()); err != nil {
		log.Fatal().Err(err).Msg("failed to create sample resources")
	}

	// Start UI
	var uiCmd *exec.Cmd
	if !skipUI {
		log.Info().Msg("starting UI...")
		uiCmd = exec.Command("npm", "run", "dev")
		uiCmd.Dir = filepath.Join(projectRoot, "ui")
		uiCmd.Env = append(os.Environ(), "NEXT_PUBLIC_API_URL=http://localhost:8080")
		uiCmd.Stdout = os.Stdout
		uiCmd.Stderr = os.Stderr
		uiCmd.Start()
	}

	// Print info
	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Println("  Demo Running")
	fmt.Println("════════════════════════════════════════════════════════")
	fmt.Printf("  API:        http://localhost%s\n", httpAPIAddr)
	if !skipUI {
		fmt.Println("  UI:         http://localhost:3000")
	}
	fmt.Printf("  Kubeconfig: %s\n", kubeconfigPath)
	fmt.Println()
	fmt.Println("  kubectl --kubeconfig=" + kubeconfigPath + " get mcpservercatalog")
	fmt.Println("  curl http://localhost:8080/v0/servers")
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop")
	fmt.Println("════════════════════════════════════════════════════════")

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info().Msg("shutting down...")
	if uiCmd != nil && uiCmd.Process != nil {
		uiCmd.Process.Signal(syscall.SIGTERM)
	}
	cancel()
	env.Stop()
	os.Remove(kubeconfigPath)
}

func findProjectRoot() (string, error) {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("project root not found")
		}
		dir = parent
	}
}

func writeKubeconfig(env *envtest.Environment) (string, error) {
	if len(env.KubeConfig) == 0 {
		return "", fmt.Errorf("no kubeconfig from envtest")
	}
	// Use fixed path in current directory for easy access
	kubeconfigPath := "./demo-kubeconfig.yaml"
	if err := os.WriteFile(kubeconfigPath, env.KubeConfig, 0600); err != nil {
		return "", err
	}
	return kubeconfigPath, nil
}

func setupReconcilers(mgr ctrl.Manager, logger zerolog.Logger) {
	(&controller.MCPServerCatalogReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Logger: logger}).SetupWithManager(mgr)
	(&controller.AgentCatalogReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Logger: logger}).SetupWithManager(mgr)
	(&controller.SkillCatalogReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Logger: logger}).SetupWithManager(mgr)
	(&controller.RegistryDeploymentReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Logger: logger}).SetupWithManager(mgr)
	(&controller.DiscoveryConfigReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Logger: logger}).SetupWithManager(mgr)
}

func createSampleResources(ctx context.Context, c client.Client) error {
	now := metav1.Now()

	// Create all resources first, then update status
	// MCP Servers
	servers := []*agentregistryv1alpha1.MCPServerCatalog{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "filesystem-server-v1.0.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.MCPServerCatalogSpec{
				Name: "filesystem-server", Version: "1.0.0", Title: "Filesystem MCP Server",
				Description: "File system operations for AI agents",
				Packages:    []agentregistryv1alpha1.Package{{RegistryType: "npm", Identifier: "@modelcontextprotocol/server-filesystem", Transport: agentregistryv1alpha1.Transport{Type: "stdio"}}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "github-server-v2.1.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.MCPServerCatalogSpec{
				Name: "github-server", Version: "2.1.0", Title: "GitHub MCP Server",
				Description: "GitHub API integration",
				Remotes:     []agentregistryv1alpha1.Transport{{Type: "streamable-http", URL: "https://mcp.example.com/github"}},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "slack-server-v1.5.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.MCPServerCatalogSpec{
				Name: "slack-server", Version: "1.5.0", Title: "Slack MCP Server",
				Description: "Slack integration",
				Packages:    []agentregistryv1alpha1.Package{{RegistryType: "npm", Identifier: "@modelcontextprotocol/server-slack", Transport: agentregistryv1alpha1.Transport{Type: "stdio"}}},
			},
		},
	}
	for _, s := range servers {
		if err := c.Create(ctx, s); err != nil {
			return err
		}
	}
	time.Sleep(100 * time.Millisecond) // Let reconciler settle
	for _, s := range servers {
		if err := c.Get(ctx, client.ObjectKeyFromObject(s), s); err != nil {
			return err
		}
		s.Status.Published = true
		s.Status.IsLatest = true
		s.Status.PublishedAt = &now
		s.Status.Status = agentregistryv1alpha1.CatalogStatusActive
		if err := c.Status().Update(ctx, s); err != nil {
			return err
		}
	}

	// Agents
	agents := []*agentregistryv1alpha1.AgentCatalog{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "research-agent-v0.5.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.AgentCatalogSpec{
				Name: "research-agent", Version: "0.5.0", Title: "Research Agent",
				Description: "AI research assistant", Image: "ghcr.io/example/research-agent:0.5.0",
				Framework: "langgraph", ModelProvider: "anthropic",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "code-review-agent-v1.2.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.AgentCatalogSpec{
				Name: "code-review-agent", Version: "1.2.0", Title: "Code Review Agent",
				Description: "Automated code review", Image: "ghcr.io/example/code-review-agent:1.2.0",
				Framework: "autogen", ModelProvider: "openai",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "devops-agent-v2.0.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.AgentCatalogSpec{
				Name: "devops-agent", Version: "2.0.0", Title: "DevOps Agent",
				Description: "DevOps automation", Image: "ghcr.io/example/devops-agent:2.0.0",
				Framework: "custom", ModelProvider: "anthropic",
			},
		},
	}
	for _, a := range agents {
		if err := c.Create(ctx, a); err != nil {
			return err
		}
	}
	time.Sleep(100 * time.Millisecond)
	for _, a := range agents {
		if err := c.Get(ctx, client.ObjectKeyFromObject(a), a); err != nil {
			return err
		}
		a.Status.Published = true
		a.Status.IsLatest = true
		a.Status.PublishedAt = &now
		a.Status.Status = agentregistryv1alpha1.CatalogStatusActive
		if err := c.Status().Update(ctx, a); err != nil {
			return err
		}
	}

	// Skills
	skills := []*agentregistryv1alpha1.SkillCatalog{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "terraform-skill-v1.5.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.SkillCatalogSpec{
				Name: "terraform-skill", Version: "1.5.0", Title: "Terraform Skill",
				Category: "infrastructure", Description: "Infrastructure management",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "sql-query-skill-v0.8.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.SkillCatalogSpec{
				Name: "sql-query-skill", Version: "0.8.0", Title: "SQL Query Skill",
				Category: "data", Description: "SQL query generation",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "code-generation-skill-v2.0.0", Namespace: "default"},
			Spec: agentregistryv1alpha1.SkillCatalogSpec{
				Name: "code-generation-skill", Version: "2.0.0", Title: "Code Generation Skill",
				Category: "development", Description: "Multi-language code generation",
			},
		},
	}
	for _, s := range skills {
		if err := c.Create(ctx, s); err != nil {
			return err
		}
	}
	time.Sleep(100 * time.Millisecond)
	for _, s := range skills {
		if err := c.Get(ctx, client.ObjectKeyFromObject(s), s); err != nil {
			return err
		}
		s.Status.Published = true
		s.Status.IsLatest = true
		s.Status.PublishedAt = &now
		s.Status.Status = agentregistryv1alpha1.CatalogStatusActive
		if err := c.Status().Update(ctx, s); err != nil {
			return err
		}
	}

	// Models (cluster-scoped)
	models := []*agentregistryv1alpha1.ModelCatalog{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "claude-3-opus-prod"},
			Spec: agentregistryv1alpha1.ModelCatalogSpec{
				Name: "claude-3-opus-prod", Provider: "Anthropic", Model: "claude-3-opus-20240229",
				Description: "Claude 3 Opus",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "gpt-4-dev"},
			Spec: agentregistryv1alpha1.ModelCatalogSpec{
				Name: "gpt-4-dev", Provider: "OpenAI", Model: "gpt-4-turbo",
				Description: "GPT-4 Turbo",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "llama3-local"},
			Spec: agentregistryv1alpha1.ModelCatalogSpec{
				Name: "llama3-local", Provider: "Ollama", Model: "llama3:70b",
				BaseURL: "http://localhost:11434", Description: "Local Llama 3",
			},
		},
	}
	for _, m := range models {
		if err := c.Create(ctx, m); err != nil {
			return err
		}
	}
	time.Sleep(100 * time.Millisecond)
	for _, m := range models {
		if err := c.Get(ctx, client.ObjectKeyFromObject(m), m); err != nil {
			return err
		}
		m.Status.Published = true
		m.Status.PublishedAt = &now
		m.Status.Status = agentregistryv1alpha1.CatalogStatusActive
		m.Status.Ready = true
		if err := c.Status().Update(ctx, m); err != nil {
			return err
		}
	}

	// RegistryDeployments
	deployments := []*agentregistryv1alpha1.RegistryDeployment{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "filesystem-server-deploy", Namespace: "default"},
			Spec: agentregistryv1alpha1.RegistryDeploymentSpec{
				ResourceName: "filesystem-server",
				Version:      "1.0.0",
				ResourceType: agentregistryv1alpha1.ResourceTypeMCP,
				Runtime:      agentregistryv1alpha1.RuntimeTypeKubernetes,
				Namespace:    "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "research-agent-deploy", Namespace: "default"},
			Spec: agentregistryv1alpha1.RegistryDeploymentSpec{
				ResourceName: "research-agent",
				Version:      "0.5.0",
				ResourceType: agentregistryv1alpha1.ResourceTypeAgent,
				Runtime:      agentregistryv1alpha1.RuntimeTypeKubernetes,
				Namespace:    "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "github-server-deploy", Namespace: "default"},
			Spec: agentregistryv1alpha1.RegistryDeploymentSpec{
				ResourceName: "github-server",
				Version:      "2.1.0",
				ResourceType: agentregistryv1alpha1.ResourceTypeMCP,
				Runtime:      agentregistryv1alpha1.RuntimeTypeKubernetes,
				Namespace:    "default",
				PreferRemote: true,
			},
		},
	}
	for _, d := range deployments {
		if err := c.Create(ctx, d); err != nil {
			return err
		}
	}
	// Let the reconciler handle status updates

	log.Info().Int("servers", len(servers)).Int("agents", len(agents)).Int("skills", len(skills)).Int("models", len(models)).Int("deployments", len(deployments)).Msg("sample resources created")
	return nil
}
