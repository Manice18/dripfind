"use client";

import Link from "next/link";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth-provider";
import { SiteHeader } from "@/components/site-header";

export default function LandingPage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) {
      router.replace("/app");
    }
  }, [loading, user, router]);

  return (
    <main className="landing">
      <SiteHeader variant="landing" />

      <section className="hero landing-hero">
        <div className="hero-visual" aria-hidden>
          <div className="hero-grain" />
          <div className="hero-photo" />
          <div className="hero-wash" />
        </div>
        <div className="hero-copy">
          <p className="brand-mark">LOOKBOOK</p>
          <h1>Turn any look into a shoppable wardrobe.</h1>
          <p className="lede">
            Paste a Pinterest pin or drop a photo. We read the outfit, then shop
            it across Myntra, Snitch, Off Duty, Bewakoof, Westside, and more.
          </p>
          <div className="hero-actions">
            <Link href="/auth?mode=signup" className="analyze-cta">
              Get started
            </Link>
            <Link href="/auth?mode=login" className="ghost-link hero-secondary">
              I already have an account
            </Link>
          </div>
        </div>
      </section>

      <section id="how" className="landing-section">
        <p className="eyebrow">How it works</p>
        <h2>Three steps from pin to cart.</h2>
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

      <section id="demo" className="landing-section landing-demo">
        <p className="eyebrow">See it</p>
        <h2>Watch LOOKBOOK in action.</h2>
        <p className="lede narrow">
          A short walkthrough of paste → analyze → shop. Video coming soon —
          drop your own recording here later.
        </p>
        <div className="video-slot" aria-label="Product demo video placeholder">
          <div className="video-slot-inner">
            <span className="video-label">Demo video</span>
            <p>Replace this block with your usage walkthrough.</p>
          </div>
        </div>
      </section>

      <section id="retailers" className="landing-section">
        <p className="eyebrow">Coverage</p>
        <h2>Built for the stores you actually browse.</h2>
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

      <section className="landing-section landing-cta">
        <p className="eyebrow">Ready</p>
        <h2>Start with one look.</h2>
        <p className="lede narrow">
          Create a free account and run your first analysis in under a minute.
        </p>
        <Link href="/auth?mode=signup" className="analyze-cta">
          Sign up free
        </Link>
      </section>

      <footer className="landing-footer">
        <span>LOOKBOOK</span>
        <span className="muted">AI outfit finder</span>
      </footer>
    </main>
  );
}
