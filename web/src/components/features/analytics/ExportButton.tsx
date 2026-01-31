import { useState } from "react"
import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { FeatureGate } from "@/components/ui/FeatureGate"
import { exportData } from "@/services/analytics"
import type { DateRangePreset, ExportFormat } from "@/types/analytics"

interface ExportButtonProps {
  linkId: string
  range: DateRangePreset
}

export default function ExportButton({ linkId, range }: ExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false)

  async function handleExport(format: ExportFormat) {
    setIsExporting(true)
    try {
      const blob = await exportData(linkId, format, range)
      const url = URL.createObjectURL(blob)
      const a = document.createElement("a")
      a.href = url
      a.download = `analytics-export.${format}`
      a.click()
      URL.revokeObjectURL(url)
    } catch {
      // Error handled silently in production
    } finally {
      setIsExporting(false)
    }
  }

  return (
    <FeatureGate feature="export_data" fallback={null}>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="outline" size="sm" disabled={isExporting}>
            {isExporting ? "Exporting..." : "Export"}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent>
          <DropdownMenuItem onClick={() => handleExport("csv")}>
            Export CSV
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => handleExport("json")}>
            Export JSON
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </FeatureGate>
  )
}
