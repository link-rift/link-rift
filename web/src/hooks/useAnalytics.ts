import { useQuery } from "@tanstack/react-query"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import * as analyticsService from "@/services/analytics"
import type { DateRangePreset, TimeSeriesInterval, DateRange } from "@/types/analytics"

export function useLinkAnalytics(
  linkId: string | undefined,
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "link-stats", wsId, linkId, range_, dateRange],
    queryFn: () => analyticsService.getLinkStats(linkId!, range_, dateRange),
    enabled: !!linkId && !!wsId,
    staleTime: 30 * 1000,
  })
}

export function useTimeSeries(
  linkId: string | undefined,
  interval: TimeSeriesInterval = "day",
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "timeseries", wsId, linkId, interval, range_, dateRange],
    queryFn: () => analyticsService.getTimeSeries(linkId!, interval, range_, dateRange),
    enabled: !!linkId && !!wsId,
    staleTime: 30 * 1000,
  })
}

export function useReferrers(
  linkId: string | undefined,
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "referrers", wsId, linkId, range_, dateRange],
    queryFn: () => analyticsService.getReferrers(linkId!, range_, dateRange),
    enabled: !!linkId && !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useCountries(
  linkId: string | undefined,
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "countries", wsId, linkId, range_, dateRange],
    queryFn: () => analyticsService.getCountries(linkId!, range_, dateRange),
    enabled: !!linkId && !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useDevices(
  linkId: string | undefined,
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "devices", wsId, linkId, range_, dateRange],
    queryFn: () => analyticsService.getDevices(linkId!, range_, dateRange),
    enabled: !!linkId && !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useBrowsers(
  linkId: string | undefined,
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "browsers", wsId, linkId, range_, dateRange],
    queryFn: () => analyticsService.getBrowsers(linkId!, range_, dateRange),
    enabled: !!linkId && !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useWorkspaceTimeSeries(
  interval: TimeSeriesInterval = "day",
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "workspace-timeseries", wsId, interval, range_, dateRange],
    queryFn: () => analyticsService.getWorkspaceTimeSeries(interval, range_, dateRange),
    enabled: !!wsId,
    staleTime: 30 * 1000,
  })
}

export function useWorkspaceReferrers(
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "workspace-referrers", wsId, range_, dateRange],
    queryFn: () => analyticsService.getWorkspaceReferrers(range_, dateRange),
    enabled: !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useWorkspaceCountries(
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "workspace-countries", wsId, range_, dateRange],
    queryFn: () => analyticsService.getWorkspaceCountries(range_, dateRange),
    enabled: !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useWorkspaceDevices(
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "workspace-devices", wsId, range_, dateRange],
    queryFn: () => analyticsService.getWorkspaceDevices(range_, dateRange),
    enabled: !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useWorkspaceBrowsers(
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "workspace-browsers", wsId, range_, dateRange],
    queryFn: () => analyticsService.getWorkspaceBrowsers(range_, dateRange),
    enabled: !!wsId,
    staleTime: 60 * 1000,
  })
}

export function useWorkspaceAnalytics(
  range_?: DateRangePreset,
  dateRange?: DateRange
) {
  const { currentWorkspace } = useWorkspaceStore()
  const wsId = currentWorkspace?.id

  return useQuery({
    queryKey: ["analytics", "workspace", wsId, range_, dateRange],
    queryFn: () => analyticsService.getWorkspaceStats(range_, dateRange),
    enabled: !!wsId,
    staleTime: 30 * 1000,
  })
}
