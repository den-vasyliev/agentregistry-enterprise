"use client"

import { useCallback, useEffect, useRef } from "react"
import { driver, type DriveStep } from "driver.js"
import "driver.js/dist/driver.css"

const TOUR_STORAGE_KEY = "agentregistry-tour-completed"

const tourSteps: DriveStep[] = [
  {
    element: "#stats-cards",
    popover: {
      title: "Overview Dashboard",
      description:
        "These cards show a quick summary of all registered resources — MCP servers, skills, agents, and models. Click any card to jump to that category.",
      side: "bottom",
      align: "center",
    },
  },
  {
    element: "#search-input",
    popover: {
      title: "Search",
      description:
        "Search across all resources by name, title, or description. Results update as you type with a small debounce delay.",
      side: "bottom",
      align: "start",
    },
  },
  {
    element: "#deployment-status-filter",
    popover: {
      title: "Deployment Status Filter",
      description:
        "Filter servers and agents by their deployment status — External, Running, Not Deployed, or Failed.",
      side: "bottom",
      align: "center",
    },
  },
  {
    element: "#verified-filters",
    popover: {
      title: "Verification Filters",
      description:
        "Filter by verified organization (blue shield) or verified publisher (green badge) to find trusted resources.",
      side: "bottom",
      align: "center",
    },
  },
  {
    element: "#refresh-button",
    popover: {
      title: "Refresh Data",
      description: "Reload all catalog data from the registry.",
      side: "bottom",
      align: "end",
    },
  },
  {
    element: "#tab-content",
    popover: {
      title: "Resource Catalog",
      description:
        "Browse resources organized by type — MCP Servers, Skills, Agents, and Models. Switch tabs to explore each category.",
      side: "top",
      align: "center",
    },
  },
  {
    element: "#resource-card",
    popover: {
      title: "Resource Card",
      description:
        "Each card shows resource badges — type, deployment status, verified organization (blue shield), and verified publisher (green check). Click to view full details.",
      side: "bottom",
      align: "start",
    },
  },
  {
    element: "#submit-button",
    popover: {
      title: "Contribute a Resource",
      description:
        "Submit your own MCP server, agent, skill, or model to the registry. Fill in resource details and get a Kubernetes manifest ready to apply.",
      side: "bottom",
      align: "end",
    },
  },
]

interface ProductTourProps {
  /** If true, automatically start tour for first-time visitors */
  autoStart?: boolean
}

export function ProductTour({ autoStart = true }: ProductTourProps) {
  const hasStarted = useRef(false)

  const startTour = useCallback(() => {
    const driverObj = driver({
      showProgress: true,
      animate: true,
      smoothScroll: true,
      allowClose: true,
      overlayColor: "black",
      overlayOpacity: 0.6,
      stagePadding: 8,
      stageRadius: 8,
      popoverClass: "agentregistry-tour-popover",
      nextBtnText: "Next →",
      prevBtnText: "← Back",
      doneBtnText: "Done",
      steps: tourSteps,
      onDestroyed: () => {
        localStorage.setItem(TOUR_STORAGE_KEY, "true")
      },
    })
    driverObj.drive()
  }, [])

  useEffect(() => {
    if (!autoStart || hasStarted.current) return
    hasStarted.current = true

    const completed = localStorage.getItem(TOUR_STORAGE_KEY)
    if (completed) return

    // Delay to let the page fully render
    const timer = setTimeout(() => {
      startTour()
    }, 800)

    return () => clearTimeout(timer)
  }, [autoStart, startTour])

  return null
}

export function useProductTour() {
  const startTour = useCallback(() => {
    const driverObj = driver({
      showProgress: true,
      animate: true,
      smoothScroll: true,
      allowClose: true,
      overlayColor: "black",
      overlayOpacity: 0.6,
      stagePadding: 8,
      stageRadius: 8,
      popoverClass: "agentregistry-tour-popover",
      nextBtnText: "Next →",
      prevBtnText: "← Back",
      doneBtnText: "Done",
      steps: tourSteps,
      onDestroyed: () => {
        localStorage.setItem(TOUR_STORAGE_KEY, "true")
      },
    })
    driverObj.drive()
  }, [])

  const resetTour = useCallback(() => {
    localStorage.removeItem(TOUR_STORAGE_KEY)
  }, [])

  return { startTour, resetTour }
}
