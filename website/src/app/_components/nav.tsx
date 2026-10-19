import Link from "next/link"

export function Nav() {
  return (
    <header className="border-border border-b">
      <div className="mx-auto flex h-16 max-w-4xl items-center justify-between px-6">
        <Link
          href="/"
          className="flex items-center gap-2 text-sm font-semibold"
        >
          <span
            className="bg-terminal size-2 rounded-full"
            aria-hidden="true"
          />
          Obscuro
        </Link>
        <nav
          aria-label="Main navigation"
          className="flex items-center gap-5 text-sm"
        >
          <Link
            href="/docs"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            Docs
          </Link>
          <a
            href="https://github.com/cranklabs/obscuro"
            className="text-muted-foreground hover:text-foreground transition-colors"
          >
            GitHub
          </a>
        </nav>
      </div>
    </header>
  )
}
