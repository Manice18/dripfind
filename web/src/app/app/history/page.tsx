"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth-provider";
import { HistoryList } from "@/components/history-list";
import { SiteHeader } from "@/components/site-header";

export default function HistoryPage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.replace("/auth?mode=login");
    }
  }, [loading, user, router]);

  if (loading || !user) {
    return (
      <main className="workspace">
        <p className="muted">Checking session…</p>
      </main>
    );
  }

  return (
    <main className="workspace history-page">
      <SiteHeader variant="app" />

      <section className="history-section">
        <p className="eyebrow">Archive</p>
        <h1>Search history</h1>
        <p className="lede narrow">
          Reopen a previous analysis or clear entries you no longer need.
        </p>
        <HistoryList />
      </section>
    </main>
  );
}
