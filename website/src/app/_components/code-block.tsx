import { cn } from "@/lib/utils"
import { CopyButton } from "./copy-button"

interface CodeBlockProps {
  code: string
  /** Show a $ prompt before the code. Default true */
  prompt?: boolean
  className?: string
}

export function CodeBlock({ code, prompt = true, className }: CodeBlockProps) {
  return (
    <div
      className={cn(
        "border-border bg-card flex min-w-0 items-start gap-2 rounded-lg border p-3 font-mono text-xs/relaxed",
        className,
      )}
    >
      {prompt && <span className="text-terminal shrink-0 select-none">$</span>}
      <pre className="min-w-0 flex-1 overflow-x-auto font-mono">
        <code className="font-mono">{code}</code>
      </pre>
      <CopyButton text={code} />
    </div>
  )
}
