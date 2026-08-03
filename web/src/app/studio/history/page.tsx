"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { HistoryList } from "@/components/history-list";
import { SessionLoader } from "@/components/session-loader";
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
    return <SessionLoader />;
  }

  return (
    <main className="workspace history-page">
      <SiteHeader variant="studio" />

      <section className="history-section">
        <h1>Search history</h1>
        <p className="lede narrow">
          Reopen a previous analysis or clear entries you no longer need.
        </p>
        <HistoryList />
      </section>
    </main>
  );
}
