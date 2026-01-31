import { useRevokeAPIKey } from "@/hooks/useAPIKeys"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import type { APIKey } from "@/types/apikey"

interface APIKeyListProps {
  apiKeys: APIKey[]
  isLoading: boolean
}

export default function APIKeyList({ apiKeys, isLoading }: APIKeyListProps) {
  const revokeKey = useRevokeAPIKey()

  if (isLoading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    )
  }

  if (apiKeys.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-center">
        <p className="text-sm text-muted-foreground">
          No API keys yet. Create one to get started with programmatic access.
        </p>
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Key Prefix</TableHead>
          <TableHead>Scopes</TableHead>
          <TableHead>Last Used</TableHead>
          <TableHead>Requests</TableHead>
          <TableHead>Created</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {apiKeys.map((key) => (
          <TableRow key={key.id}>
            <TableCell className="font-medium">{key.name}</TableCell>
            <TableCell>
              <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
                {key.key_prefix}...
              </code>
            </TableCell>
            <TableCell>
              <div className="flex flex-wrap gap-1">
                {key.scopes.slice(0, 3).map((scope) => (
                  <Badge key={scope} variant="secondary" className="text-xs">
                    {scope}
                  </Badge>
                ))}
                {key.scopes.length > 3 && (
                  <Badge variant="outline" className="text-xs">
                    +{key.scopes.length - 3}
                  </Badge>
                )}
              </div>
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {key.last_used_at
                ? new Date(key.last_used_at).toLocaleDateString()
                : "Never"}
            </TableCell>
            <TableCell className="text-sm">
              {key.request_count.toLocaleString()}
            </TableCell>
            <TableCell className="text-sm text-muted-foreground">
              {new Date(key.created_at).toLocaleDateString()}
            </TableCell>
            <TableCell className="text-right">
              <Button
                variant="destructive"
                size="sm"
                onClick={() => revokeKey.mutate(key.id)}
                disabled={revokeKey.isPending}
              >
                Revoke
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
