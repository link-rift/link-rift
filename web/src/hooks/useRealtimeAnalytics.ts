import { useEffect, useRef, useState, useCallback } from "react"
import { useQueryClient } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import type { ClickNotification } from "@/types/analytics"

const MAX_RECENT_CLICKS = 20
const RECONNECT_DELAY = 3000

export function useRealtimeAnalytics(linkId?: string) {
  const { currentWorkspace } = useWorkspaceStore()
  const queryClient = useQueryClient()
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>(undefined)

  const [isConnected, setIsConnected] = useState(false)
  const [recentClicks, setRecentClicks] = useState<ClickNotification[]>([])

  const connect = useCallback(() => {
    if (!currentWorkspace?.id) return

    const token = localStorage.getItem("access_token")
    if (!token) return

    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:"
    const wsUrl = `${protocol}//${window.location.host}/ws/analytics?token=${token}&workspace_id=${currentWorkspace.id}`

    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onopen = () => {
      setIsConnected(true)
      if (linkId) {
        ws.send(JSON.stringify({ action: "subscribe", link_id: linkId }))
      }
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === "click") {
          const notification = msg.data as ClickNotification
          setRecentClicks((prev) => [notification, ...prev].slice(0, MAX_RECENT_CLICKS))

          // Invalidate analytics queries to refresh data
          queryClient.invalidateQueries({ queryKey: ["analytics"] })
        }
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      setIsConnected(false)
      wsRef.current = null
      // Auto-reconnect
      reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [currentWorkspace?.id, linkId, queryClient])

  useEffect(() => {
    connect()

    return () => {
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [connect])

  // Subscribe/unsubscribe to link when linkId changes
  useEffect(() => {
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) return

    if (linkId) {
      ws.send(JSON.stringify({ action: "subscribe", link_id: linkId }))
    }

    return () => {
      if (linkId && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ action: "unsubscribe", link_id: linkId }))
      }
    }
  }, [linkId])

  return { isConnected, recentClicks }
}
