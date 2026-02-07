export default function KubernetesIcon({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      {/* Kubernetes-style hexagon with helm wheel */}
      <path d="M12 2L3 7v10l9 5 9-5V7l-9-5z" />
      <circle cx="12" cy="12" r="2.5" />
      <line x1="12" y1="2" x2="12" y2="9.5" />
      <line x1="12" y1="14.5" x2="12" y2="22" />
      <line x1="3" y1="7" x2="9.8" y2="10.6" />
      <line x1="14.2" y1="13.4" x2="21" y2="17" />
      <line x1="21" y1="7" x2="14.2" y2="10.6" />
      <line x1="9.8" y1="13.4" x2="3" y2="17" />
    </svg>
  )
}
