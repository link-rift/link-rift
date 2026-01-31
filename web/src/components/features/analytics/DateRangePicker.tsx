import { useState } from "react"
import { format } from "date-fns"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import type { DateRangePreset, DateRange } from "@/types/analytics"

interface DateRangePickerProps {
  selectedPreset: DateRangePreset
  onPresetChange: (preset: DateRangePreset) => void
  onCustomRangeChange?: (range: DateRange) => void
}

const presets: { label: string; value: DateRangePreset }[] = [
  { label: "24h", value: "24h" },
  { label: "7d", value: "7d" },
  { label: "30d", value: "30d" },
  { label: "90d", value: "90d" },
]

export default function DateRangePicker({
  selectedPreset,
  onPresetChange,
  onCustomRangeChange,
}: DateRangePickerProps) {
  const [customStart, setCustomStart] = useState<Date>()
  const [customEnd, setCustomEnd] = useState<Date>()
  const [isCustom, setIsCustom] = useState(false)

  function handlePreset(preset: DateRangePreset) {
    setIsCustom(false)
    onPresetChange(preset)
  }

  function handleCustomApply() {
    if (customStart && customEnd && onCustomRangeChange) {
      onCustomRangeChange({
        start: customStart.toISOString(),
        end: customEnd.toISOString(),
      })
    }
  }

  return (
    <div className="flex items-center gap-2">
      {presets.map((p) => (
        <Button
          key={p.value}
          variant={selectedPreset === p.value && !isCustom ? "default" : "outline"}
          size="sm"
          onClick={() => handlePreset(p.value)}
        >
          {p.label}
        </Button>
      ))}

      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant={isCustom ? "default" : "outline"}
            size="sm"
            onClick={() => setIsCustom(true)}
          >
            {isCustom && customStart && customEnd
              ? `${format(customStart, "MMM d")} - ${format(customEnd, "MMM d")}`
              : "Custom"}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-4" align="end">
          <div className="space-y-3">
            <div>
              <p className="mb-1 text-xs font-medium text-muted-foreground">Start</p>
              <Calendar
                mode="single"
                selected={customStart}
                onSelect={(d) => d && setCustomStart(d)}
                disabled={(date) => date > new Date()}
              />
            </div>
            <div>
              <p className="mb-1 text-xs font-medium text-muted-foreground">End</p>
              <Calendar
                mode="single"
                selected={customEnd}
                onSelect={(d) => d && setCustomEnd(d)}
                disabled={(date) => date > new Date()}
              />
            </div>
            <Button size="sm" className="w-full" onClick={handleCustomApply}>
              Apply
            </Button>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
