"use client"

import { useEffect, useState, useCallback } from "react"
import dynamic from "next/dynamic"
import { Button } from "@/components/ui/button"
import { ArrowLeft, RefreshCw } from "lucide-react"
import { adminApiClient, type DiscoveryMapConfig } from "@/lib/admin-api"

const DiscoveryMapGraph = dynamic(
  () =>
    import("@/components/discovery-map-graph").then(
      (mod) => mod.DiscoveryMapGraph,
    ),
  { ssr: false },
)

interface DiscoveryMapViewProps {
  onClose: () => void
  serverCount?: number
  agentCount?: number
  skillCount?: number
  modelCount?: number
}

export function DiscoveryMapView({ onClose, serverCount = 0, agentCount = 0, skillCount = 0, modelCount = 0 }: DiscoveryMapViewProps) {
  const [configs, setConfigs] = useState<DiscoveryMapConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchData = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await adminApiClient.getDiscoveryMap()
      setConfigs(data.configs || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load discovery map")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  return (
    <main className="min-h-screen bg-background flex flex-col">
      <div className="border-b bg-muted/30">
        <div className="container mx-auto px-6 py-4 flex items-center gap-4">
          <Button variant="ghost" size="icon" onClick={onClose}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          <div className="flex-1">
            <h1 className="text-lg font-semibold">Discovery Map</h1>
            <p className="text-sm text-muted-foreground">
              Topology of auto-discovery across clusters, environments, and resource types
            </p>
          </div>
          <Button variant="ghost" size="icon" onClick={fetchData} title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="flex-1 p-4">
        {loading && (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-3"></div>
              <p className="text-sm text-muted-foreground">Loading discovery topology...</p>
            </div>
          </div>
        )}

        {error && !loading && (
          <div className="flex items-center justify-center h-full">
            <div className="text-center">
              <p className="text-sm text-red-500 mb-2">{error}</p>
              <Button variant="outline" size="sm" onClick={fetchData}>
                Retry
              </Button>
            </div>
          </div>
        )}

        {!loading && !error && (
          <DiscoveryMapGraph
            configs={configs}
            resourceCounts={{ mcpServers: serverCount, agents: agentCount, skills: skillCount, models: modelCount }}
          />
        )}
      </div>
    </main>
  )
}
