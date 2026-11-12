"use client"

import { Check, Copy } from "lucide-react"
import { useState } from "react"

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
          "inline-flex shrink-0 cursor-pointer items-center rounded-sm border border-border p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none",
          className,
        )}
        aria-label="Copy to clipboard"
      >
        {copied ? (
          <Check className="size-3.5 text-terminal" />
        ) : (
          <Copy className="size-3.5 text-muted-foreground" />
        )}
      </TooltipTrigger>
      <TooltipContent>{copied ? "Copied!" : "Copy"}</TooltipContent>
    </Tooltip>
  )
}
