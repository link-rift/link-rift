import { useState } from "react"
import { useSCIMTokens, useCreateSCIMToken, useRevokeSCIMToken } from "@/hooks/useSCIM"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import type { SCIMToken } from "@/types/scim"

export default function SCIMTokenList() {
  const { data: tokens, isLoading } = useSCIMTokens()
  const createMutation = useCreateSCIMToken()
  const revokeMutation = useRevokeSCIMToken()

  const [name, setName] = useState("")
  const [newToken, setNewToken] = useState<string | null>(null)

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return

    createMutation.mutate(
      { name: name.trim() },
      {
        onSuccess: (data) => {
          setNewToken(data.token)
          setName("")
        },
      }
    )
  }

  const handleRevoke = (token: SCIMToken) => {
    if (confirm(`Revoke token "${token.name}"?`)) {
      revokeMutation.mutate(token.id)
    }
  }

  if (isLoading) {
    return <div className="py-4 text-sm text-muted-foreground">Loading tokens...</div>
  }

  return (
    <div className="space-y-4">
      <form onSubmit={handleCreate} className="flex items-end gap-3">
        <div className="flex-1">
          <Label htmlFor="token-name">Token Name</Label>
          <Input
            id="token-name"
            placeholder="e.g., Okta SCIM"
            value={name}
            onChange={(e) => setName(e.target.value)}
            required
          />
        </div>
        <Button type="submit" disabled={createMutation.isPending}>
          {createMutation.isPending ? "Creating..." : "Create Token"}
        </Button>
      </form>

      {newToken && (
        <div className="rounded-lg border border-yellow-500/50 bg-yellow-50 p-4 dark:bg-yellow-900/20">
          <p className="mb-2 text-sm font-medium">
            Copy this token now. It won't be shown again.
          </p>
          <code className="block break-all rounded bg-background p-2 font-mono text-xs">
            {newToken}
          </code>
          <Button
            variant="outline"
            size="sm"
            className="mt-2"
            onClick={() => {
              navigator.clipboard.writeText(newToken)
              setNewToken(null)
            }}
          >
            Copy & Dismiss
          </Button>
        </div>
      )}

      <div className="space-y-2">
        {tokens && tokens.length > 0 ? (
          tokens.map((token) => (
            <div
              key={token.id}
              className="flex items-center justify-between rounded-lg border p-3"
            >
              <div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">{token.name}</span>
                  {!token.is_active && (
                    <span className="rounded bg-red-100 px-1.5 py-0.5 text-xs text-red-700 dark:bg-red-900/30 dark:text-red-400">
                      Revoked
                    </span>
                  )}
                </div>
                <div className="mt-1 text-xs text-muted-foreground">
                  <span className="font-mono">{token.token_prefix}...</span>
                  {" · Created "}
                  {new Date(token.created_at).toLocaleDateString()}
                  {token.last_used_at && (
                    <>
                      {" · Last used "}
                      {new Date(token.last_used_at).toLocaleDateString()}
                    </>
                  )}
                  {token.expires_at && (
                    <>
                      {" · Expires "}
                      {new Date(token.expires_at).toLocaleDateString()}
                    </>
                  )}
                </div>
              </div>
              {token.is_active && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleRevoke(token)}
                  disabled={revokeMutation.isPending}
                >
                  Revoke
                </Button>
              )}
            </div>
          ))
        ) : (
          <p className="py-4 text-center text-sm text-muted-foreground">
            No SCIM tokens yet. Create one to enable SCIM provisioning.
          </p>
        )}
      </div>
    </div>
  )
}
