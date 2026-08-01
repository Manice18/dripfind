"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth-provider";

export function LandingGate({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) {
      router.replace("/studio");
    }
  }, [loading, user, router]);

  if (loading || user) {
    return (
      <main className="landing landing-gate" aria-busy="true">
        <p className="muted" role="status">
          {user ? "Opening your LOOKBOOK…" : "Loading LOOKBOOK…"}
        </p>
      </main>
    );
  }

  return <>{children}</>;
}
