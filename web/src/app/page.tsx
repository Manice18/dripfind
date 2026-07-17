"use client";

import { useMutation, useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useMemo, useState } from "react";
import { AnalyzeForm } from "@/components/analyze-form";
import { LoadingState } from "@/components/loading-state";
import { OutfitResult } from "@/components/outfit-result";
import { api } from "@/lib/api";

function HomeInner() {
  const router = useRouter();
  const params = useSearchParams();
  const id = params.get("id");
  const [stage, setStage] = useState(0);

  const analyze = useMutation({
    mutationFn: api.analyze,
    onSuccess: (data) => {
      router.push(`/?id=${data.id}`);
    },
  });

  const result = useQuery({
    queryKey: ["result", id],
    queryFn: () => api.result(id!),
    enabled: !!id,
    refetchInterval: (q) => {
      const status = q.state.data?.status;
      if (status === "completed" || status === "failed") return false;
      return 1500;
    },
  });

  useEffect(() => {
    if (!id || result.data?.status === "completed" || result.data?.status === "failed") {
      return;
    }
    const t = setInterval(() => setStage((s) => (s + 1) % 5), 2200);
    return () => clearInterval(t);
  }, [id, result.data?.status]);

  const showHero = !id;
  const showLoading =
    !!id &&
    (result.isLoading ||
      result.data?.status === "pending" ||
      result.data?.status === "processing");
  const showResult = result.data?.status === "completed";
  const showError =
    analyze.isError ||
    result.isError ||
    result.data?.status === "failed";

  const errorMessage = useMemo(() => {
    if (analyze.isError) return (analyze.error as Error).message;
    if (result.isError) return (result.error as Error).message;
    if (result.data?.status === "failed") return result.data.error_message || "Analysis failed";
    return "";
  }, [analyze.isError, analyze.error, result.isError, result.error, result.data]);

  return (
    <main>
      <header className="site-header">
        <Link href="/" className="brand">
          LOOKBOOK
        </Link>
        <nav>
          <Link href="/history">History</Link>
        </nav>
      </header>

      {showHero && (
        <section className="hero">
          <div className="hero-visual" aria-hidden>
            <div className="hero-grain" />
            <div className="hero-photo" />
            <div className="hero-wash" />
          </div>
          <div className="hero-copy">
            <p className="brand-mark">LOOKBOOK</p>
            <h1>Turn any Pinterest look into a shoppable wardrobe.</h1>
            <p className="lede">
              Paste a Pinterest pin URL. We read the outfit, then shop it across
              Myntra, Snitch, Off Duty, Bewakoof, Westside, and more.
            </p>
            <AnalyzeForm
              onSubmit={(url) => analyze.mutate(url)}
              loading={analyze.isPending}
            />
            {analyze.isError && <p className="error">{errorMessage}</p>}
          </div>
        </section>
      )}

      {!showHero && (
        <section className="workspace">
          <div className="workspace-top">
            <Link href="/" className="ghost-link">
              ← New search
            </Link>
            <AnalyzeForm
              onSubmit={(url) => analyze.mutate(url)}
              loading={analyze.isPending}
            />
          </div>

          {showLoading && <LoadingState stage={stage} />}
          {showError && !showLoading && (
            <div className="error-panel">
              <h2>Couldn’t finish this look</h2>
              <p>{errorMessage}</p>
              <Link href="/">Try another URL</Link>
            </div>
          )}
          {showResult && result.data && <OutfitResult outfit={result.data} />}
        </section>
      )}
    </main>
  );
}

export default function HomePage() {
  return (
    <Suspense fallback={<main className="workspace"><p className="muted">Loading…</p></main>}>
      <HomeInner />
    </Suspense>
  );
}
