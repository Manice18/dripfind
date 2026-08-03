import Link from "next/link";

import { HeroVisual } from "@/components/hero-visual";
import { LandingGate } from "@/components/landing-gate";
import { LandingScrollFX } from "@/components/landing-scroll-fx";
import { SiteHeader } from "@/components/site-header";

export default function LandingPage() {
  return (
    <LandingGate>
      <main className="landing" id="main-content">
        <LandingScrollFX />
        <a href="#how" className="skip-link">
          Skip to content
        </a>

        <SiteHeader variant="landing" />

        <section
          className="hero landing-hero"
          aria-labelledby="landing-hero-heading"
        >
          <HeroVisual priority />
          <div className="hero-copy">
            <p className="brand-mark">LOOKBOOK</p>
            <h1 id="landing-hero-heading">
              Turn any look into a shoppable wardrobe.
            </h1>
            <p className="lede">
              Paste a Pinterest pin or drop a photo. We read the outfit, then
              shop it across Myntra, Snitch, Off Duty, Bewakoof, Westside, and
              more.
            </p>
            <div className="hero-actions">
              <Link href="/auth?mode=signup" className="analyze-cta" prefetch>
                Get started
              </Link>
              <Link
                href="/auth?mode=login"
                className="ghost-link hero-secondary"
                prefetch
              >
                I already have an account
              </Link>
            </div>
          </div>
        </section>

        <section
          id="how"
          className="landing-section landing-section-how"
          data-landing-section
          aria-labelledby="landing-how-heading"
          tabIndex={-1}
        >
          <h2 id="landing-how-heading">Three steps from pin to cart.</h2>
          <p className="lede narrow">
            No moodboards to reverse-engineer by hand — drop a look and get
            matched products with prices and links.
          </p>
          <ol className="steps">
            <li>
              <strong>Drop a look</strong>
              <span>Pinterest URL or any outfit photo / screenshot.</span>
            </li>
            <li>
              <strong>We read the fit</strong>
              <span>Category, color, material, and vibe — piece by piece.</span>
            </li>
            <li>
              <strong>Shop the matches</strong>
              <span>Live results from Indian retailers you already use.</span>
            </li>
          </ol>
        </section>

        <section
          id="demo"
          className="landing-section landing-section-demo landing-demo"
          data-landing-section
          aria-labelledby="landing-demo-heading"
          tabIndex={-1}
        >
          <h2 id="landing-demo-heading">Watch LOOKBOOK in action.</h2>
          <p className="lede narrow">
            A short walkthrough of paste → analyze → shop — coming soon.
          </p>
          <div
            className="video-slot"
            role="img"
            aria-label="Product demo video coming soon"
          >
            <div className="video-slot-inner">
              <span className="video-label">Coming soon</span>
              <p>See paste, analyze, and shop in one pass.</p>
            </div>
          </div>
        </section>

        <section
          id="retailers"
          className="landing-section landing-section-retailers"
          data-landing-section
          aria-labelledby="landing-retailers-heading"
          tabIndex={-1}
        >
          <h2 id="landing-retailers-heading">
            Built for the stores you actually browse.
          </h2>
          <p className="lede narrow">
            Homegrown brands and big marketplaces in one pass — so a street look
            doesn’t stop at a single site.
          </p>
          <ul className="retailer-row">
            <li>Myntra</li>
            <li>Ajio</li>
            <li>Flipkart</li>
            <li>Snitch</li>
            <li>Bewakoof</li>
            <li>Westside</li>
            <li>H&M</li>
            <li>Off Duty</li>
          </ul>
        </section>

        <section
          className="landing-section landing-section-cta landing-cta"
          data-landing-section
          aria-labelledby="landing-cta-heading"
        >
          <h2 id="landing-cta-heading">Start with one look.</h2>
          <p className="lede narrow">
            Create a free account and run your first analysis in under a minute.
          </p>
          <Link href="/auth?mode=signup" className="analyze-cta" prefetch>
            Sign up free
          </Link>
        </section>

        <footer className="landing-footer">
          <span>LOOKBOOK</span>
          <span className="muted">AI outfit finder</span>
        </footer>
      </main>
    </LandingGate>
  );
}
