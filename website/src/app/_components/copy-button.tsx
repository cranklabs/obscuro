"use client"

import { useState } from "react"
import { Check, Copy } from "lucide-react"

import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"

interface CopyButtonProps {
  text: string
  className?: string
}

export function CopyButton({ text, className }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Tooltip>
      <TooltipTrigger
        onClick={handleCopy}
        className={cn(
          "border-border text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:ring-ring inline-flex shrink-0 cursor-pointer items-center rounded-sm border p-1.5 transition-colors focus-visible:ring-2 focus-visible:outline-none",
          className,
        )}
        aria-label="Copy to clipboard"
      >
        {copied ? (
          <Check className="text-terminal size-3.5" />
        ) : (
          <Copy className="text-muted-foreground size-3.5" />
        )}
      </TooltipTrigger>
      <TooltipContent>{copied ? "Copied!" : "Copy"}</TooltipContent>
    </Tooltip>
  )
}
