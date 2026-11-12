import Link from "next/link"

export function Footer() {
  return (
    <footer className="border-t border-border">
      <div className="mx-auto flex max-w-4xl flex-wrap items-center justify-between gap-4 p-6 text-sm text-muted-foreground">
        <span>&copy; {new Date().getFullYear()} Jank Labs</span>
        <nav aria-label="Footer" className="flex gap-5">
          <Link
            href="/docs/security"
            className="transition-colors hover:text-foreground"
          >
            Security
          </Link>
          <a
            href="https://github.com/cranklabs/obscuro/releases"
            className="transition-colors hover:text-foreground"
          >
            Releases
          </a>
        </nav>
      </div>
    </footer>
  )
}
