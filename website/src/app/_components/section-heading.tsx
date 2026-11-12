import { cn } from "@/lib/utils"

interface SectionHeadingProps {
  children: React.ReactNode
  className?: string
  id?: string
}

export function SectionHeading({
  children,
  className,
  id,
}: SectionHeadingProps) {
  return (
    <h2
      id={id}
      className={cn(
        "mb-8 text-xs font-bold tracking-[0.3em] text-muted-foreground uppercase",
        className,
      )}
    >
      {children}
    </h2>
  )
}
