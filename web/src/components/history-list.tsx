"use client";

import Link from "next/link";
import posthog from "posthog-js";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, mediaURL } from "@/lib/api";

export function HistoryList() {
  const qc = useQueryClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ["history"],
    queryFn: api.history,
  });

  const del = useMutation({
    mutationFn: api.deleteHistory,
    onSuccess: () => {
      posthog.capture("history_entry_deleted");
      qc.invalidateQueries({ queryKey: ["history"] });
    },
  });

  function deleteHistoryEntry(entryId: string) {
    del.mutate(entryId);
  }

  if (isLoading) return <p className="muted">Loading history…</p>;
  if (error) return <p className="error">{(error as Error).message}</p>;

  const items = data?.items ?? [];
  if (items.length === 0) {
    return (
      <p className="muted">
        No searches yet.{" "}
        <Link href="/studio" className="ghost-link">
          Run a new search
        </Link>
        .
      </p>
    );
  }

  return (
    <ul className="history-list">
      {items.map((entry) => (
        <li key={entry.id}>
          <Link href={`/studio?id=${entry.outfit_id}`} className="history-link">
            {entry.image_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img
                src={mediaURL(entry.image_url)}
                alt=""
                loading="lazy"
                decoding="async"
              />
            ) : (
              <div className="history-placeholder" />
            )}
            <div>
              <strong>{entry.style || "Outfit"}</strong>
              <p>{new Date(entry.created_at).toLocaleString()}</p>
              <span className="site">{entry.status}</span>
            </div>
          </Link>
          <button
            type="button"
            className="ghost"
            onClick={() => deleteHistoryEntry(entry.id)}
            aria-label="Delete history entry"
          >
            Remove
          </button>
        </li>
      ))}
    </ul>
  );
}
