import { useState } from "react"
import { useInviteMember } from "@/hooks/useWorkspace"
import { useWorkspaceStore } from "@/stores/workspaceStore"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { WorkspaceRole } from "@/types/workspace"

interface InviteMemberModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function InviteMemberModal({ open, onOpenChange }: InviteMemberModalProps) {
  const { currentWorkspace } = useWorkspaceStore()
  const inviteMember = useInviteMember()
  const [email, setEmail] = useState("")
  const [role, setRole] = useState<WorkspaceRole>("viewer")

  function handleInvite() {
    if (!email.trim() || !currentWorkspace) return
    inviteMember.mutate(
      { workspaceId: currentWorkspace.id, data: { email: email.trim(), role } },
      {
        onSuccess: () => {
          onOpenChange(false)
          setEmail("")
          setRole("viewer")
        },
      }
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Invite Member</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="invite-email">Email Address</Label>
            <Input
              id="invite-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="user@example.com"
              onKeyDown={(e) => e.key === "Enter" && handleInvite()}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="invite-role">Role</Label>
            <Select value={role} onValueChange={(v) => setRole(v as WorkspaceRole)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="viewer">Viewer - Can view links and analytics</SelectItem>
                <SelectItem value="editor">Editor - Can create and edit links</SelectItem>
                <SelectItem value="admin">Admin - Can manage members and settings</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleInvite}
              disabled={!email.trim() || inviteMember.isPending}
            >
              {inviteMember.isPending ? "Inviting..." : "Invite"}
            </Button>
          </div>
          {inviteMember.isError && (
            <p className="text-sm text-destructive">
              {inviteMember.error instanceof Error
                ? inviteMember.error.message
                : "Failed to invite member"}
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
