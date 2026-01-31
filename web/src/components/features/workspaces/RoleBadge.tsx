import { Badge } from "@/components/ui/badge"
import type { WorkspaceRole } from "@/types/workspace"
import { ROLE_LABELS } from "@/types/workspace"

const roleVariants: Record<WorkspaceRole, string> = {
  owner: "bg-amber-100 text-amber-800 border-amber-200",
  admin: "bg-purple-100 text-purple-800 border-purple-200",
  editor: "bg-blue-100 text-blue-800 border-blue-200",
  viewer: "bg-gray-100 text-gray-800 border-gray-200",
}

interface RoleBadgeProps {
  role: WorkspaceRole
}

export default function RoleBadge({ role }: RoleBadgeProps) {
  return (
    <Badge variant="outline" className={roleVariants[role]}>
      {ROLE_LABELS[role]}
    </Badge>
  )
}
