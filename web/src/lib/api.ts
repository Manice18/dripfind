import type { HistoryEntry, Outfit } from "@/types/outfit";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  const isFormData = typeof FormData !== "undefined" && init?.body instanceof FormData;
  if (!isFormData && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error((data as { error?: string }).error ?? `Request failed (${res.status})`);
  }
  return data as T;
}

export function mediaURL(path?: string) {
  if (!path) return "";
  if (path.startsWith("http")) return path;
  return `${API_BASE}${path.startsWith("/") ? "" : "/"}${path}`;
}

export const api = {
  analyze: (url: string) =>
    request<{ id: string }>("/analyze", {
      method: "POST",
      body: JSON.stringify({ url }),
    }),

  analyzeUpload: (file: File) => {
    const body = new FormData();
    body.append("image", file);
    return request<{ id: string }>("/analyze/upload", {
      method: "POST",
      body,
    });
  },

  result: (id: string) => request<Outfit>(`/result/${id}`),

  history: () => request<{ items: HistoryEntry[] }>("/history"),

  deleteHistory: (id: string) =>
    request<{ status: string }>(`/history/${id}`, { method: "DELETE" }),
};
