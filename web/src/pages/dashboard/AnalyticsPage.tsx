import { useState } from "react"
import { useParams, Link } from "react-router-dom"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  useLinkAnalytics,
  useTimeSeries,
  useReferrers,
  useCountries,
  useDevices,
  useBrowsers,
  useWorkspaceAnalytics,
} from "@/hooks/useAnalytics"
import { useRealtimeAnalytics } from "@/hooks/useRealtimeAnalytics"
import DateRangePicker from "@/components/features/analytics/DateRangePicker"
import StatsCards from "@/components/features/analytics/StatsCards"
import ClicksChart from "@/components/features/analytics/ClicksChart"
import ReferrersTable from "@/components/features/analytics/ReferrersTable"
import CountriesChart from "@/components/features/analytics/CountriesChart"
import DevicesPieChart from "@/components/features/analytics/DevicesPieChart"
import BrowsersChart from "@/components/features/analytics/BrowsersChart"
import ExportButton from "@/components/features/analytics/ExportButton"
import RealtimeIndicator from "@/components/features/analytics/RealtimeIndicator"
import type { DateRangePreset, TimeSeriesInterval, DateRange } from "@/types/analytics"

export default function AnalyticsPage() {
  const { linkId } = useParams<{ linkId: string }>()
  const isLinkView = !!linkId

  const [range, setRange] = useState<DateRangePreset>("7d")
  const [customRange, setCustomRange] = useState<DateRange>()
  const [interval, setInterval] = useState<TimeSeriesInterval>("day")

  const activeRange = customRange ? undefined : range
  const activeDateRange = customRange

  // Workspace-level analytics
  const workspaceAnalytics = useWorkspaceAnalytics(activeRange, activeDateRange)

  // Link-level analytics
  const linkStats = useLinkAnalytics(linkId, activeRange, activeDateRange)
  const timeSeries = useTimeSeries(linkId, interval, activeRange, activeDateRange)
  const referrers = useReferrers(linkId, activeRange, activeDateRange)
  const countries = useCountries(linkId, activeRange, activeDateRange)
  const devices = useDevices(linkId, activeRange, activeDateRange)
  const browsers = useBrowsers(linkId, activeRange, activeDateRange)

  // Real-time WebSocket
  const { isConnected, recentClicks } = useRealtimeAnalytics(linkId)

  function handlePresetChange(preset: DateRangePreset) {
    setRange(preset)
    setCustomRange(undefined)
  }

  function handleCustomRangeChange(dr: DateRange) {
    setCustomRange(dr)
  }

  const stats = isLinkView ? linkStats.data : workspaceAnalytics.data
  const isLoadingStats = isLinkView
    ? linkStats.isLoading
    : workspaceAnalytics.isLoading

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold">
            {isLinkView ? "Link Analytics" : "Analytics"}
          </h1>
          {isLinkView && (
            <Link
              to="/analytics"
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Back to workspace analytics
            </Link>
          )}
        </div>
        <div className="flex items-center gap-3">
          <RealtimeIndicator
            isConnected={isConnected}
            recentClicks={recentClicks}
          />
          {isLinkView && <ExportButton linkId={linkId!} range={range} />}
        </div>
      </div>

      <DateRangePicker
        selectedPreset={range}
        onPresetChange={handlePresetChange}
        onCustomRangeChange={handleCustomRangeChange}
      />

      <StatsCards stats={stats} isLoading={isLoadingStats} />

      {isLinkView ? (
        <>
          <ClicksChart
            data={timeSeries.data}
            isLoading={timeSeries.isLoading}
            interval={interval}
            onIntervalChange={setInterval}
          />

          <Tabs defaultValue="referrers">
            <TabsList>
              <TabsTrigger value="referrers">Referrers</TabsTrigger>
              <TabsTrigger value="countries">Countries</TabsTrigger>
              <TabsTrigger value="devices">Devices</TabsTrigger>
              <TabsTrigger value="browsers">Browsers</TabsTrigger>
            </TabsList>
            <TabsContent value="referrers">
              <ReferrersTable data={referrers.data} isLoading={referrers.isLoading} />
            </TabsContent>
            <TabsContent value="countries">
              <CountriesChart data={countries.data} isLoading={countries.isLoading} />
            </TabsContent>
            <TabsContent value="devices">
              <DevicesPieChart data={devices.data} isLoading={devices.isLoading} />
            </TabsContent>
            <TabsContent value="browsers">
              <BrowsersChart data={browsers.data} isLoading={browsers.isLoading} />
            </TabsContent>
          </Tabs>
        </>
      ) : (
        workspaceAnalytics.data?.top_links && workspaceAnalytics.data.top_links.length > 0 && (
          <div className="rounded-lg border">
            <div className="border-b p-4">
              <h2 className="font-semibold">Top Links</h2>
            </div>
            <div className="divide-y">
              {workspaceAnalytics.data.top_links.map((link) => (
                <Link
                  key={link.link_id}
                  to={`/analytics/${link.link_id}`}
                  className="flex items-center justify-between p-4 transition-colors hover:bg-muted/50"
                >
                  <span className="text-sm font-medium text-primary">
                    /{link.short_code}
                  </span>
                  <span className="text-sm text-muted-foreground">
                    {link.total_clicks} clicks
                  </span>
                </Link>
              ))}
            </div>
          </div>
        )
      )}
    </div>
  )
}
