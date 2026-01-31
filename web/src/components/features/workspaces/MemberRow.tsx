import { useWorkspaceStore } from "@/stores/workspaceStore"
import { useUpdateMemberRole, useRemoveMember } from "@/hooks/useWorkspace"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import RoleBadge from "./RoleBadge"
import type { WorkspaceMember, WorkspaceRole } from "@/types/workspace"

interface MemberRowProps {
  member: WorkspaceMember
}

export default function MemberRow({ member }: MemberRowProps) {
  const { currentWorkspace, canManageMembers, isOwner } = useWorkspaceStore()
  const updateRole = useUpdateMemberRole()
  const removeMember = useRemoveMember()

  const canManage = canManageMembers()
  const isCurrentOwner = isOwner()
  const isOwnerMember = member.role === "owner"
  const canChangeRole = canManage && !isOwnerMember
  const canRemove = canManage && !isOwnerMember && (isCurrentOwner || member.role !== "admin")

  function handleRoleChange(newRole: WorkspaceRole) {
    if (!currentWorkspace) return
    updateRole.mutate({
      workspaceId: currentWorkspace.id,
      userId: member.user_id,
      data: { role: newRole },
    })
  }

  function handleRemove() {
    if (!currentWorkspace) return
    if (!confirm(`Remove ${member.name || member.email} from the workspace?`)) return
    removeMember.mutate({
      workspaceId: currentWorkspace.id,
      userId: member.user_id,
    })
  }

  return (
    <div className="flex items-center justify-between gap-4 rounded-md border px-4 py-3">
      <div className="flex items-center gap-3 min-w-0">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted text-sm font-medium">
          {(member.name || member.email).charAt(0).toUpperCase()}
        </div>
        <div className="min-w-0">
          <p className="truncate text-sm font-medium">{member.name || "Unnamed"}</p>
          <p className="truncate text-xs text-muted-foreground">{member.email}</p>
        </div>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        {canChangeRole ? (
          <Select
            value={member.role}
            onValueChange={(v) => handleRoleChange(v as WorkspaceRole)}
            disabled={updateRole.isPending}
          >
            <SelectTrigger className="w-28 h-8 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="viewer">Viewer</SelectItem>
              <SelectItem value="editor">Editor</SelectItem>
              <SelectItem value="admin">Admin</SelectItem>
            </SelectContent>
          </Select>
        ) : (
          <RoleBadge role={member.role} />
        )}
        {canRemove && (
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive h-8 px-2"
            onClick={handleRemove}
            disabled={removeMember.isPending}
          >
            Remove
          </Button>
        )}
      </div>
    </div>
  )
}
