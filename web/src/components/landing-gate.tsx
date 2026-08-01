"use client";

import { useEffect, type ReactNode } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth-provider";
import { SessionLoader } from "@/components/session-loader";

export function LandingGate({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) {
      router.replace("/studio");
    }
  }, [loading, user, router]);

  if (loading || user) {
    return <SessionLoader />;
  }

  return <>{children}</>;
}
