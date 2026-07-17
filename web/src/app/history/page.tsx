"use client";

import Link from "next/link";
import { HistoryList } from "@/components/history-list";

export default function HistoryPage() {
  return (
    <main className="workspace history-page">
      <header className="site-header">
        <Link href="/" className="brand">
          LOOKBOOK
        </Link>
        <nav>
          <Link href="/">New search</Link>
        </nav>
      </header>

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
