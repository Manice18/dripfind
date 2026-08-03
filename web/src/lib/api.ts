import type { HistoryEntry, Outfit } from "@/types/outfit";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type AuthUser = {
  id: string;
  email: string;
  name: string;
  email_verified: boolean;
  created_at: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  const isFormData =
    typeof FormData !== "undefined" && init?.body instanceof FormData;
  if (!isFormData && !headers.has("Content-Type") && init?.body) {
    headers.set("Content-Type", "application/json");
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers,
    credentials: "include",
  });

  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(
      (data as { error?: string }).error ?? `Request failed (${res.status})`,
    );
  }
  return data as T;
}

export function mediaURL(path?: string) {
  if (!path) return "";
  if (path.startsWith("http")) return path;
  return `${API_BASE}${path.startsWith("/") ? "" : "/"}${path}`;
}

export function googleAuthURL() {
  return `${API_BASE}/auth/google`;
}

export const api = {
  me: () => request<{ user: AuthUser }>("/auth/me"),

  register: (body: { email: string; password: string; name?: string }) =>
    request<{ user: AuthUser }>("/auth/register", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  login: (body: { email: string; password: string }) =>
    request<{ user: AuthUser }>("/auth/login", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  logout: () =>
    request<{ status: string }>("/auth/logout", {
      method: "POST",
      body: JSON.stringify({}),
    }),

  deleteAccount: () =>
    request<{ status: string }>("/auth/account", {
      method: "DELETE",
    }),

  requestOTP: (email: string) =>
    request<{ status: string; message: string }>("/auth/otp/request", {
      method: "POST",
      body: JSON.stringify({ email }),
    }),

  verifyOTP: (email: string, code: string) =>
    request<{ user: AuthUser }>("/auth/otp/verify", {
      method: "POST",
      body: JSON.stringify({ email, code }),
    }),

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
