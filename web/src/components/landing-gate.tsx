"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { SessionLoader } from "@/components/session-loader";

/**
 * Redirect signed-in users to the studio, but always SSR/paint the landing
 * HTML while the session check is in flight so crawlers and hash deep-links
 * see real content (not the "hold up" loader).
 */
export function LandingGate({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) {
      router.replace("/studio");
    }
  }, [loading, user, router]);

  if (!loading && user) {
    return <SessionLoader />;
  }

  return <>{children}</>;
}
