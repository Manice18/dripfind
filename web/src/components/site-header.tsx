"use client";

import Link from "next/link";

import { useAuth } from "@/components/auth-provider";
import { IconLogout, IconUser } from "@/components/icons";

type HeaderVariant = "landing" | "studio" | "auth";

export function SiteHeader({
  variant = "landing",
}: {
  variant?: HeaderVariant;
}) {
  const { user, loading, logout } = useAuth();
  const displayName = user?.name || user?.email || "Account";
  const navLabel =
    variant === "landing"
      ? "Account"
      : variant === "auth"
        ? "Account"
        : "Studio";

  return (
    <header
      className={`site-header${variant === "landing" ? " site-header-landing" : ""}`}
    >
      <Link href={user ? "/studio" : "/"} className="brand">
        DRIPFIND
      </Link>

      {variant === "landing" && (
        <nav className="landing-jump" aria-label="On this page">
          <a href="#how">How it works</a>
          <a href="#demo">Demo</a>
          <a href="#retailers">Retailers</a>
        </nav>
      )}

      <nav className="site-nav site-nav-primary" aria-label={navLabel}>
        {variant === "studio" && user && (
          <>
            <Link href="/studio">Search</Link>
            <Link href="/studio/history">History</Link>
          </>
        )}
        {loading && (
          <span
            className="nav-status"
            role="status"
            aria-label="Checking session"
          >
            <span className="nav-status-spinner" aria-hidden="true" />
          </span>
        )}
        {!loading && !user && variant !== "auth" && (
          <>
            <Link href="/auth?mode=login" className="nav-ghost">
              Log in
            </Link>
            <Link href="/auth?mode=signup" className="nav-cta">
              Sign up
            </Link>
          </>
        )}
        {!loading && !user && variant === "auth" && (
          <Link href="/" className="nav-ghost">
            Home
          </Link>
        )}
        {!loading && user && (
          <>
            <Link
              href="/studio/settings"
              className="nav-icon-btn nav-icon-btn-profile"
              aria-label={`Account settings for ${displayName}`}
              title={displayName}
            >
              <IconUser />
            </Link>
            <button
              type="button"
              className="nav-icon-btn"
              aria-label="Log out"
              title="Log out"
              onClick={() =>
                void logout().then(() => {
                  window.location.href = "/";
                })
              }
            >
              <IconLogout />
            </button>
          </>
        )}
      </nav>
    </header>
  );
}
