import { RootProvider } from "fumadocs-ui/provider/next"
import type { Metadata } from "next"
import localFont from "next/font/local"

import { TooltipProvider } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"

import "./globals.css"

const satoshi = localFont({
  src: "./fonts/Satoshi-Variable.woff2",
  variable: "--font-satoshi",
  display: "swap",
})

const clashGrotesk = localFont({
  src: "./fonts/ClashGrotesk-Variable.woff2",
  variable: "--font-clash",
  display: "swap",
})

const clashDisplay = localFont({
  src: "./fonts/ClashDisplay-Variable.woff2",
  variable: "--font-clash-display",
  display: "swap",
})

const iosevka = localFont({
  src: [
    {
      path: "./fonts/IosevkaSS02-Regular.woff2",
      weight: "400",
      style: "normal",
    },
    {
      path: "./fonts/IosevkaSS02-Medium.woff2",
      weight: "500",
      style: "normal",
    },
    {
      path: "./fonts/IosevkaSS02-SemiBold.woff2",
      weight: "600",
      style: "normal",
    },
    { path: "./fonts/IosevkaSS02-Bold.woff2", weight: "700", style: "normal" },
  ],
  variable: "--font-iosevka",
  display: "swap",
})

export const metadata: Metadata = {
  title: "Obscuro — Encrypted secrets your agents can use without reading",
  description:
    "Encrypted secrets in git that AI coding agents can list, reference, run, and deploy with, without the values entering their context.",
  metadataBase: new URL(
    process.env.SITE_URL || "https://cranklabs.github.io/obscuro/",
  ),
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang="en"
      className={cn(
        satoshi.variable,
        clashGrotesk.variable,
        clashDisplay.variable,
        iosevka.variable,
        "dark scroll-smooth",
      )}
      data-scroll-behavior="smooth"
      suppressHydrationWarning
    >
      <body className="min-h-screen bg-background text-foreground antialiased">
        <RootProvider
          search={{
            options: {
              type: "static",
              api: `${process.env.NEXT_PUBLIC_BASE_PATH || ""}/api/search.json`,
            },
          }}
          theme={{
            enabled: false,
          }}
        >
          <TooltipProvider>{children}</TooltipProvider>
        </RootProvider>
      </body>
    </html>
  )
}
