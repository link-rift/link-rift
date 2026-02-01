import { useState } from "react"
import { Link } from "react-router-dom"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"

export default function SSOLoginPage() {
  const [workspaceSlug, setWorkspaceSlug] = useState("")
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!workspaceSlug.trim()) return

    setIsLoading(true)
    // Redirect to the SSO login endpoint with the workspace slug
    // The backend will resolve the slug to a workspace ID and redirect to the IdP
    const baseUrl = window.location.origin
    window.location.href = `${baseUrl}/api/v1/auth/sso/login/${workspaceSlug.trim()}`
  }

  return (
    <div className="space-y-6">
      <div className="text-center">
        <h1 className="text-2xl font-bold">Sign in with SSO</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Enter your workspace identifier to sign in via your company's
          identity provider.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <Label htmlFor="workspace-slug">Workspace Identifier</Label>
          <Input
            id="workspace-slug"
            placeholder="your-company"
            value={workspaceSlug}
            onChange={(e) => setWorkspaceSlug(e.target.value)}
            required
          />
        </div>
        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? "Redirecting..." : "Continue with SSO"}
        </Button>
      </form>

      <div className="text-center text-sm">
        <Link to="/auth/login" className="text-primary hover:underline">
          Back to regular login
        </Link>
      </div>
    </div>
  )
}
