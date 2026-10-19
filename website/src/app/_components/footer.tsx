import Link from "next/link"

export function Footer() {
  return (
    <footer className="border-border border-t">
      <div className="text-muted-foreground mx-auto flex max-w-4xl flex-wrap items-center justify-between gap-4 p-6 text-sm">
        <span>&copy; {new Date().getFullYear()} Jank Labs</span>
        <nav aria-label="Footer" className="flex gap-5">
          <Link
            href="/docs/security"
            className="hover:text-foreground transition-colors"
          >
            Security
          </Link>
          <a
            href="https://github.com/cranklabs/obscuro/releases"
            className="hover:text-foreground transition-colors"
          >
            Releases
          </a>
        </nav>
      </div>
    </footer>
  )
}
