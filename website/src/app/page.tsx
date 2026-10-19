import Link from "next/link"

import { CodeBlock } from "./_components/code-block"
import { Footer } from "./_components/footer"
import { Nav } from "./_components/nav"

const INSTALL_SCRIPT =
  "curl -fsSL https://raw.githubusercontent.com/cranklabs/obscuro/main/install.sh | sh"

const steps = [
  {
    title: "Store",
    description: "Create a vault and add a secret.",
    code: "obscuro init\nobscuro set API_KEY",
  },
  {
    title: "Reference",
    description: "Use placeholders in your config.yaml.",
    code: "api_key: __API_KEY__",
  },
  {
    title: "Inject",
    description: "Replace placeholders at deploy time.",
    code: "obscuro inject < config.yaml",
  },
]

export default function Home() {
  return (
    <>
      <Nav />
      <main className="mx-auto w-full max-w-4xl px-6">
        <section className="py-16 sm:py-20" aria-labelledby="hero-heading">
          <h1
            id="hero-heading"
            className="font-display text-4xl/tight font-semibold tracking-tight sm:text-5xl/tight"
          >
            Secrets in git,
            <br />
            <span className="text-terminal">not in plaintext.</span>
          </h1>
          <p className="text-muted-foreground mt-5 max-w-xl text-base/relaxed">
            Obscuro encrypts secrets in your repository and injects them into
            config files at deploy time.
          </p>
          <div className="mt-6 flex flex-wrap items-center gap-5 text-sm">
            <Link
              href="/docs"
              className="bg-primary text-primary-foreground hover:bg-primary/90 inline-flex rounded-md px-4 py-2 font-medium transition-colors"
            >
              Read the docs
            </Link>
            <Link
              href="/docs/installation"
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              Installation options
            </Link>
          </div>
          <div className="mt-10">
            <h2 className="mb-3 text-sm font-medium">Install</h2>
            <CodeBlock code={INSTALL_SCRIPT} />
          </div>
        </section>

        <section
          className="border-border border-t py-10 sm:py-12"
          aria-labelledby="how-it-works-heading"
        >
          <h2 id="how-it-works-heading" className="text-xl font-semibold">
            How it works
          </h2>
          <ol className="mt-6 grid gap-6 md:grid-cols-3">
            {steps.map((step, index) => (
              <li key={step.title} className="min-w-0">
                <h3 className="text-sm font-medium">
                  <span className="text-muted-foreground mr-2">
                    {index + 1}.
                  </span>
                  {step.title}
                </h3>
                <p className="text-muted-foreground mt-2 mb-3 text-sm/relaxed">
                  {step.description}
                </p>
                <CodeBlock code={step.code} prompt={false} />
              </li>
            ))}
          </ol>
          <p className="text-muted-foreground mt-6 text-sm/relaxed">
            Commit the encrypted vault and your templates, never decrypted
            output.
          </p>
          <div className="mt-6 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm">
            <span className="text-muted-foreground">Guides</span>
            <Link
              href="/docs/guides/helm-integration"
              className="hover:text-terminal transition-colors"
            >
              Helm
            </Link>
            <Link
              href="/docs/guides/docker-compose"
              className="hover:text-terminal transition-colors"
            >
              Docker Compose
            </Link>
            <Link
              href="/docs/guides/kubernetes"
              className="hover:text-terminal transition-colors"
            >
              Kubernetes
            </Link>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
