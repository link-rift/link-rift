import { useAuthStore } from "@/stores/authStore"

export default function DashboardPage() {
  const { user } = useAuthStore()

  return (
    <div>
      <h2 className="text-2xl font-bold tracking-tight">
        Welcome{user?.name ? `, ${user.name}` : ""}
      </h2>
      <p className="text-muted-foreground mt-1">
        Here&apos;s an overview of your link performance.
      </p>
    </div>
  )
}
