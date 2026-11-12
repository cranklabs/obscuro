import Link from "next/link"

export function Nav() {
  return (
    <header className="border-b border-border">
      <div className="mx-auto flex h-16 max-w-4xl items-center justify-between px-6">
        <Link
          href="/"
          className="flex items-center gap-2 text-sm font-semibold"
        >
          <span
            className="size-2 rounded-full bg-terminal"
            aria-hidden="true"
          />
          Obscuro
        </Link>
        <nav
          aria-label="Main navigation"
          className="flex items-center gap-5 text-sm"
        >
          <Link
            href="/docs/agents"
            className="text-muted-foreground transition-colors hover:text-foreground"
          >
            Agents
          </Link>
          <Link
            href="/docs"
            className="text-muted-foreground transition-colors hover:text-foreground"
          >
            Docs
          </Link>
          <a
            href="https://github.com/cranklabs/obscuro"
            className="text-muted-foreground transition-colors hover:text-foreground"
          >
            GitHub
          </a>
        </nav>
      </div>
    </header>
  )
}
