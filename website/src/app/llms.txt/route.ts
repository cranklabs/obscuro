import { absoluteUrl } from "@/lib/site"
import { source } from "@/lib/source"

export const dynamic = "force-static"

export function GET() {
  const pages = source
    .getPages()
    .map(
      (page) =>
        `- [${page.data.title}](${absoluteUrl(page.url)}): ${page.data.description}`,
    )
    .join("\n")

  const body = [
    "# Obscuro",
    "> Encrypted secrets in git that AI coding agents can use without reading. Values are encrypted with Argon2id + AES-256-GCM in .obscuro/ at the repository root; agents list names, run commands with secrets in the environment, and render __KEY__ placeholders without printing values.",
    "## Agents",
    `- [Agent Guide (Markdown)](${absoluteUrl("/agents.md")}): Rules and workflows for AI coding agents. Read this first.`,
    "## Docs",
    pages,
  ].join("\n\n")

  return new Response(`${body}\n`, {
    headers: { "Content-Type": "text/plain; charset=utf-8" },
  })
}
