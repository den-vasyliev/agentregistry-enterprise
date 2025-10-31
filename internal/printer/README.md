# Table Printer Package

A kubectl-style table printer for CLI output that provides clean, minimal formatting.

## Features

- **Minimal Design**: Clean, tab-separated columns with no fancy borders (like kubectl)
- **Column Alignment**: Automatic column width adjustment using Go's `tabwriter`
- **Consistent Formatting**: Standardized status messages and timestamps
- **Age Formatting**: kubectl-style age strings (e.g., "5d", "3h", "45m")

## Usage

### Basic Table

```go
import "github.com/agentregistry-dev/agentregistry/internal/printer"

// Create a new table printer
t := printer.NewTablePrinter(os.Stdout)

// Set headers
t.SetHeaders("Name", "Status", "Age")

// Add rows
t.AddRow("server-1", "Running", "5d")
t.AddRow("server-2", "Stopped", "2h")

// Render the table
t.Render()
```

Output:
```
NAME       STATUS    AGE
server-1   Running   5d
server-2   Stopped   2h
```

### Helper Functions

#### TruncateString
Truncate long strings with ellipsis:
```go
printer.TruncateString("very-long-server-name-that-needs-truncation", 20)
// Returns: "very-long-server-..."
```

#### FormatStatus
Format installation status:
```go
printer.FormatStatus(true)   // Returns: "Installed"
printer.FormatStatus(false)  // Returns: "Available"
```

#### FormatAge
kubectl-style age formatting:
```go
printer.FormatAge(time.Now().Add(-5 * 24 * time.Hour))  // Returns: "5d"
printer.FormatAge(time.Now().Add(-3 * time.Hour))       // Returns: "3h"
printer.FormatAge(time.Now().Add(-45 * time.Minute))    // Returns: "45m"
```

#### EmptyValueOrDefault
Handle empty values:
```go
printer.EmptyValueOrDefault("", "<none>")     // Returns: "<none>"
printer.EmptyValueOrDefault("value", "<none>") // Returns: "value"
```

### Status Messages

```go
printer.PrintSuccess("Operation completed successfully")
printer.PrintError("Failed to connect")
printer.PrintWarning("Deprecated feature")
printer.PrintInfo("Processing...")
```

## Design Philosophy

This package follows kubectl's table output philosophy:
- **Simple and Clean**: No decorative borders or colors
- **Machine-Readable**: Tab-separated output that's easy to parse
- **Consistent**: Same formatting across all commands
- **Minimal**: Focus on content, not decoration

## Examples from arctl

### List MCP Servers
```
NAME                                      TITLE                           VERSION   TRANSPORT   STATUS
@modelcontextprotocol/server-everything  MCP Everything Server           0.6.2     stdio       Available
@modelcontextprotocol/server-filesystem  Filesystem MCP Server           0.6.0     stdio       Installed
```

### List Registries
```
NAME              URL                                   TYPE     AGE
mcp-registry      https://api.mcp.registry/servers      public   5d
private-registry  https://private.company.com/mcp       private  2h
```

### Show Server Details
```
PROPERTY   VALUE
Name       @modelcontextprotocol/server-filesystem
Title      Filesystem MCP Server
Version    0.6.0
Status     Installed
Registry   mcp-registry
```

## Comparison with go-pretty

**Before (go-pretty):**
```
┌──────────────┬──────────────────────┬─────────┬──────────┐
│ NAME         │ TITLE                │ VERSION │ STATUS   │
├──────────────┼──────────────────────┼─────────┼──────────┤
│ server-1     │ My Server            │ 1.0.0   │ Running  │
│ server-2     │ Another Server       │ 2.0.0   │ Stopped  │
└──────────────┴──────────────────────┴─────────┴──────────┘
```

**After (kubectl-style):**
```
NAME        TITLE           VERSION   STATUS
server-1    My Server       1.0.0     Running
server-2    Another Server  2.0.0     Stopped
```

The kubectl-style is cleaner, more standard, and easier to parse programmatically.

