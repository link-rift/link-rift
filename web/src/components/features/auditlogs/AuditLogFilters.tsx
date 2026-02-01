import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  AUDIT_ACTIONS,
  AUDIT_RESOURCE_TYPES,
  ACTION_LABELS,
  RESOURCE_TYPE_LABELS,
} from "@/types/auditlog"
import type { AuditLogFilter } from "@/types/auditlog"

interface AuditLogFiltersProps {
  filter: AuditLogFilter
  onFilterChange: (filter: AuditLogFilter) => void
  onExport: (format: "json" | "csv") => void
}

export default function AuditLogFilters({
  filter,
  onFilterChange,
  onExport,
}: AuditLogFiltersProps) {
  return (
    <div className="flex flex-wrap items-end gap-3">
      <div className="w-40">
        <label className="mb-1 block text-xs font-medium text-muted-foreground">
          Action
        </label>
        <Select
          value={filter.action ?? "all"}
          onValueChange={(v) =>
            onFilterChange({ ...filter, action: v === "all" ? undefined : v })
          }
        >
          <SelectTrigger>
            <SelectValue placeholder="All actions" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All actions</SelectItem>
            {AUDIT_ACTIONS.map((action) => (
              <SelectItem key={action} value={action}>
                {ACTION_LABELS[action] ?? action}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="w-40">
        <label className="mb-1 block text-xs font-medium text-muted-foreground">
          Resource
        </label>
        <Select
          value={filter.resource_type ?? "all"}
          onValueChange={(v) =>
            onFilterChange({
              ...filter,
              resource_type: v === "all" ? undefined : v,
            })
          }
        >
          <SelectTrigger>
            <SelectValue placeholder="All resources" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All resources</SelectItem>
            {AUDIT_RESOURCE_TYPES.map((rt) => (
              <SelectItem key={rt} value={rt}>
                {RESOURCE_TYPE_LABELS[rt] ?? rt}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      <div className="w-40">
        <label className="mb-1 block text-xs font-medium text-muted-foreground">
          From
        </label>
        <Input
          type="date"
          value={filter.start_date ?? ""}
          onChange={(e) =>
            onFilterChange({
              ...filter,
              start_date: e.target.value || undefined,
            })
          }
        />
      </div>
      <div className="w-40">
        <label className="mb-1 block text-xs font-medium text-muted-foreground">
          To
        </label>
        <Input
          type="date"
          value={filter.end_date ?? ""}
          onChange={(e) =>
            onFilterChange({
              ...filter,
              end_date: e.target.value || undefined,
            })
          }
        />
      </div>
      <div className="flex gap-2">
        <Button variant="outline" size="sm" onClick={() => onExport("csv")}>
          Export CSV
        </Button>
        <Button variant="outline" size="sm" onClick={() => onExport("json")}>
          Export JSON
        </Button>
      </div>
    </div>
  )
}
