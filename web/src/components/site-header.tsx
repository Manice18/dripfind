"use client";

import Link from "next/link";
import { useAuth } from "@/components/auth-provider";

export function SiteHeader({ variant = "landing" }: { variant?: "landing" | "app" }) {
  const { user, loading, logout } = useAuth();

  return (
    <header className="site-header">
      <Link href={user ? "/app" : "/"} className="brand">
        LOOKBOOK
      </Link>
      <nav className="site-nav">
        {variant === "landing" && (
          <>
            <a href="#how">How it works</a>
            <a href="#demo">Demo</a>
            <a href="#retailers">Retailers</a>
          </>
        )}
        {variant === "app" && user && (
          <>
            <Link href="/app">Search</Link>
            <Link href="/app/history">History</Link>
          </>
        )}
        {!loading && !user && (
          <>
            <Link href="/auth?mode=login" className="nav-ghost">
              Log in
            </Link>
            <Link href="/auth?mode=signup" className="nav-cta">
              Sign up
            </Link>
          </>
        )}
        {!loading && user && (
          <>
            <span className="nav-user">{user.name || user.email}</span>
            <button type="button" className="nav-ghost-btn" onClick={() => void logout().then(() => {
              window.location.href = "/";
            })}>
              Log out
            </button>
          </>
        )}
      </nav>
    </header>
  );
}
