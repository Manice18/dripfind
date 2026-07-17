"use client";

import { FormEvent, useState } from "react";

export function AnalyzeForm({
  onSubmit,
  loading,
}: {
  onSubmit: (url: string) => void;
  loading?: boolean;
}) {
  const [url, setUrl] = useState("");

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = url.trim();
    if (!trimmed || loading) return;
    onSubmit(trimmed);
  }

  return (
    <form onSubmit={handleSubmit} className="analyze-form">
      <label htmlFor="outfit-url" className="sr-only">
        Pinterest URL
      </label>
      <input
        id="outfit-url"
        type="url"
        required
        placeholder="Paste a Pinterest pin URL"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        disabled={loading}
      />
      <button type="submit" disabled={loading || !url.trim()}>
        {loading ? "Reading look…" : "Recreate outfit"}
      </button>
    </form>
  );
}
