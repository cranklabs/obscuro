import Link from "next/link"

import { CodeBlock } from "./_components/code-block"
import { Footer } from "./_components/footer"
import { Nav } from "./_components/nav"

const INSTALL_SCRIPT =
  "curl -fsSL https://raw.githubusercontent.com/cranklabs/obscuro/main/install.sh | sh"

const AGENT_PROMPT =
  "Read https://cranklabs.github.io/obscuro/agents.md and follow it for secrets in this repo."

const steps = [
  {
    title: "You store",
    description:
      "Create the vault, add secrets, and keep the password in your keychain.",
    code: "obscuro init\nobscuro set API_KEY",
  },
  {
    title: "Agents reference",
    description: "Agents list names, never values, and use them in config.",
    code: "obscuro list\n# api_key: __API_KEY__",
  },
  {
    title: "Agents run",
    description: "Secrets go straight into the process or deploy tool.",
    code: "obscuro run -- npm test\nobscuro inject --strict < app.yaml | kubectl apply -f -",
  },
]

const reasons = [
  {
    title: "Values stay out of context",
    description:
      "list, run, and inject let agents work with secrets without printing them into transcripts or logs.",
  },
  {
    title: "Non-interactive by design",
    description:
      "Once the password is in your keychain, every command works in an agent shell with no TTY.",
  },
  {
    title: "Safe to diff and commit",
    description:
      "The vault is ciphertext in your repo, so agent commits and reviews never contain plaintext.",
  },
  {
    title: "Generated secrets stay hidden",
    description:
      "Pipe a generator into obscuro set and the new value goes into the vault and nowhere else.",
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
            Secrets your agents can use,
            <br />
            <span className="text-terminal">without reading them.</span>
          </h1>
          <p className="text-muted-foreground mt-5 max-w-xl text-base/relaxed">
            Obscuro keeps encrypted secrets in your repository. AI coding agents
            can discover, reference, and deploy with them while the values stay
            out of their context.
          </p>
          <div className="mt-6 flex flex-wrap items-center gap-5 text-sm">
            <Link
              href="/docs"
              className="bg-primary text-primary-foreground hover:bg-primary/90 inline-flex rounded-md px-4 py-2 font-medium transition-colors"
            >
              Read the docs
            </Link>
            <Link
              href="/docs/agents"
              className="text-muted-foreground hover:text-foreground transition-colors"
            >
              Agent guide
            </Link>
          </div>
          <div className="mt-10 grid gap-6">
            <div>
              <h2 className="mb-3 text-sm font-medium">Install</h2>
              <CodeBlock code={INSTALL_SCRIPT} />
            </div>
            <div>
              <h2 className="mb-3 text-sm font-medium">Tell your agent</h2>
              <CodeBlock code={AGENT_PROMPT} prompt={false} />
            </div>
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
        </section>

        <section
          className="border-border border-t py-10 sm:py-12"
          aria-labelledby="agents-heading"
        >
          <h2 id="agents-heading" className="text-xl font-semibold">
            Built for coding agents
          </h2>
          <ul className="mt-6 grid gap-6 sm:grid-cols-2">
            {reasons.map((reason) => (
              <li key={reason.title}>
                <h3 className="text-sm font-medium">{reason.title}</h3>
                <p className="text-muted-foreground mt-2 text-sm/relaxed">
                  {reason.description}
                </p>
              </li>
            ))}
          </ul>
          <p className="text-muted-foreground mt-6 text-sm/relaxed">
            Agent rules are a convention, not a sandbox. Read the{" "}
            <Link href="/docs/agents" className="hover:text-terminal underline">
              agent guide
            </Link>{" "}
            or give agents the plain{" "}
            <a
              href={`${process.env.NEXT_PUBLIC_BASE_PATH || ""}/agents.md`}
              className="hover:text-terminal underline"
            >
              agents.md
            </a>
            .
          </p>
          <div className="mt-6 flex flex-wrap items-center gap-x-5 gap-y-2 text-sm">
            <span className="text-muted-foreground">Guides</span>
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
            <Link
              href="/docs/guides/helm-integration"
              className="hover:text-terminal transition-colors"
            >
              Helm
            </Link>
          </div>
        </section>
      </main>
      <Footer />
    </>
  )
}
