// Transform functions to convert backend data to frontend types

import { ServerDetail } from "./api"
import { MCPServerWithStatus, Server } from "./types"

export function transformServerDetail(serverDetail: ServerDetail): MCPServerWithStatus {
  // Parse the JSON data blob
  let serverData: Server
  try {
    serverData = JSON.parse(serverDetail.data)
  } catch (err) {
    // Fallback if data is invalid
    serverData = {
      name: serverDetail.name,
      title: serverDetail.title,
      description: serverDetail.description,
      version: serverDetail.version,
      websiteUrl: serverDetail.website_url,
    }
  }

  return {
    installed: serverDetail.installed,
    installedAt: serverDetail.installed ? serverDetail.updated_at : undefined,
    _meta: {
      "io.modelcontextprotocol.registry/official": {
        isLatest: true, // We can enhance this later
        publishedAt: serverDetail.created_at,
        status: "active",
        updatedAt: serverDetail.updated_at,
      },
    },
    server: {
      ...serverData,
      name: serverDetail.name,
      title: serverDetail.title || serverData.title,
      description: serverDetail.description,
      version: serverDetail.version,
      websiteUrl: serverDetail.website_url || serverData.websiteUrl,
    },
    // Store the database ID for API calls
    _dbId: serverDetail.id,
  }
}

export function transformServerList(servers: ServerDetail[]): MCPServerWithStatus[] {
  return servers.map(transformServerDetail)
}

