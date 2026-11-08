import { getPageMarkdown } from "@/lib/site"

export const dynamic = "force-static"

export async function GET() {
  return new Response(`${await getPageMarkdown(["agents"])}\n`, {
    headers: { "Content-Type": "text/markdown; charset=utf-8" },
  })
}
