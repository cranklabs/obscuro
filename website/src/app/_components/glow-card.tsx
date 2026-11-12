import { cn } from "@/lib/utils"

interface GlowCardProps extends React.ComponentProps<"div"> {
  children: React.ReactNode
}

export function GlowCard({ className, children, ...props }: GlowCardProps) {
  return (
    <div
      className={cn(
        "group rounded-lg border border-border bg-card p-6 transition-all duration-300",
        "hover:border-terminal hover:shadow-[0_0_20px_var(--color-terminal-dim),inset_0_0_20px_var(--color-terminal-dim)]",
        className,
      )}
      {...props}
    >
      {children}
    </div>
  )
}
