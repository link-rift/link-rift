import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"

interface LinkFiltersProps {
  search: string
  onSearchChange: (search: string) => void
  showActive: boolean | undefined
  onActiveChange: (active: boolean | undefined) => void
}

export default function LinkFilters({
  search,
  onSearchChange,
  showActive,
  onActiveChange,
}: LinkFiltersProps) {
  return (
    <div className="flex items-center gap-3">
      <div className="relative flex-1">
        <svg
          className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
        <Input
          placeholder="Search links..."
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-9"
        />
      </div>
      <div className="flex gap-1">
        <Button
          variant={showActive === undefined ? "default" : "ghost"}
          size="sm"
          onClick={() => onActiveChange(undefined)}
        >
          All
        </Button>
        <Button
          variant={showActive === true ? "default" : "ghost"}
          size="sm"
          onClick={() => onActiveChange(true)}
        >
          Active
        </Button>
        <Button
          variant={showActive === false ? "default" : "ghost"}
          size="sm"
          onClick={() => onActiveChange(false)}
        >
          Inactive
        </Button>
      </div>
    </div>
  )
}
