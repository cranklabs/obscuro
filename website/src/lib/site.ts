import { source } from "./source"

const SITE_URL = (
  process.env.SITE_URL || "https://cranklabs.github.io/obscuro"
).replace(/\/+$/, "")

export function absoluteUrl(path: string) {
  return `${SITE_URL}${path.startsWith("/") ? path : `/${path}`}`
}

export async function getPageMarkdown(slug: string[]) {
  const page = source.getPage(slug)
  if (!page) throw new Error(`Missing docs page: /docs/${slug.join("/")}`)

  const raw = await page.data.getText("raw")
  const body = raw
    .replace(/^---\r?\n[\s\S]*?\r?\n---\r?\n/, "")
    .replace(/\]\((\/docs[^)]*)\)/g, (_, href: string) => {
      return `](${absoluteUrl(href)})`
    })
    .trim()

  return [
    `# ${page.data.title}`,
    `> ${page.data.description}`,
    `Source: ${absoluteUrl(page.url)}`,
    body,
  ].join("\n\n")
}
