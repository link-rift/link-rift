import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import type { Tier } from "@/types/license"

interface UpgradePromptProps {
  feature?: string
  requiredTier: Tier
  variant?: "inline" | "card" | "banner"
  className?: string
}

const tierNames: Record<Tier, string> = {
  free: "Community",
  pro: "Pro",
  business: "Business",
  enterprise: "Enterprise",
}

export function UpgradePrompt({
  feature,
  requiredTier,
  variant = "card",
  className,
}: UpgradePromptProps) {
  const tierName = tierNames[requiredTier]

  if (variant === "inline") {
    return (
      <span className={cn("text-sm text-muted-foreground", className)}>
        {feature ? `${feature} requires` : "Requires"} the{" "}
        <strong>{tierName}</strong> plan.{" "}
        <Button variant="link" size="sm" className="h-auto p-0">
          Upgrade
        </Button>
      </span>
    )
  }

  if (variant === "banner") {
    return (
      <div
        className={cn(
          "flex items-center justify-between rounded-lg border border-primary/20 bg-primary/5 px-4 py-3",
          className
        )}
      >
        <p className="text-sm">
          {feature ? (
            <>
              <strong>{feature}</strong> requires the {tierName} plan.
            </>
          ) : (
            <>This feature requires the {tierName} plan.</>
          )}
        </p>
        <Button size="sm">Upgrade to {tierName}</Button>
      </div>
    )
  }

  return (
    <Card className={cn("", className)}>
      <CardHeader>
        <CardTitle>Upgrade Required</CardTitle>
        <CardDescription>
          {feature
            ? `${feature} is available on the ${tierName} plan and above.`
            : `This feature requires the ${tierName} plan or higher.`}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Button>Upgrade to {tierName}</Button>
      </CardContent>
    </Card>
  )
}
