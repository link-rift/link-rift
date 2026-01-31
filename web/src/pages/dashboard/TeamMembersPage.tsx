import { useState } from "react"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { useWorkspaceMembers } from "@/hooks/useWorkspace"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import MemberRow from "@/components/features/workspaces/MemberRow"
import InviteMemberModal from "@/components/features/workspaces/InviteMemberModal"

export default function TeamMembersPage() {
  const { currentWorkspace, canManageMembers } = useWorkspaceStore()
  const { data: members, isLoading } = useWorkspaceMembers(currentWorkspace?.id ?? "")
  const [inviteOpen, setInviteOpen] = useState(false)

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Team Members</h2>
          <p className="text-sm text-muted-foreground">
            Manage who has access to this workspace.
          </p>
        </div>
        {canManageMembers() && (
          <Button onClick={() => setInviteOpen(true)}>Invite Member</Button>
        )}
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-16 w-full" />
          ))}
        </div>
      ) : members && members.length > 0 ? (
        <div className="space-y-2">
          {members.map((member) => (
            <MemberRow key={member.id} member={member} />
          ))}
        </div>
      ) : (
        <div className="rounded-md border border-dashed p-8 text-center">
          <p className="text-sm text-muted-foreground">No members found.</p>
        </div>
      )}

      <InviteMemberModal open={inviteOpen} onOpenChange={setInviteOpen} />
    </div>
  )
}
