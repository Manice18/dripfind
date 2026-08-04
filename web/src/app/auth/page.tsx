"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useEffect, useMemo, useState } from "react";

import { useAuth } from "@/components/auth-provider";
import { SessionLoader } from "@/components/session-loader";
import { SiteHeader } from "@/components/site-header";
import { api, googleAuthURL } from "@/lib/api";

type Mode = "login" | "signup";

function AuthInner() {
  const router = useRouter();
  const params = useSearchParams();
  const { setUser, user, loading: authLoading } = useAuth();

  const modeFromUrl = (
    params.get("mode") === "signup" ? "signup" : "login"
  ) as Mode;
  const oauthError = params.get("error");

  const [mode, setMode] = useState<Mode>(modeFromUrl);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const title = mode === "signup" ? "Create your account" : "Welcome back";
  const subtitle = useMemo(() => {
    if (mode === "signup")
      return "Save looks, reopen history, and shop your wardrobe.";
    return "Sign in to continue finding shoppable outfits.";
  }, [mode]);

  useEffect(() => {
    setMode(modeFromUrl);
  }, [modeFromUrl]);

  useEffect(() => {
    if (!authLoading && user) {
      router.replace("/studio");
    }
  }, [authLoading, user, router]);

  function switchMode(next: Mode) {
    setMode(next);
    setError("");
    const q = new URLSearchParams(params.toString());
    q.set("mode", next);
    q.delete("error");
    router.replace(`/auth?${q.toString()}`);
  }

  async function onPasswordSubmit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const res =
        mode === "signup"
          ? await api.register({ email, password, name })
          : await api.login({ email, password });
      setUser(res.user);
      router.push("/studio");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }

  if (authLoading || user) {
    return <SessionLoader />;
  }

  return (
    <main className="auth-page">
      <SiteHeader variant="auth" />
      <section className="auth-shell">
        <div className="auth-panel">
          <h1>{title}</h1>
          <p className="lede narrow">{subtitle}</p>

          {oauthError && (
            <p className="error" role="alert">
              Google sign-in didn’t complete. Try again, or continue with email.
            </p>
          )}

          <form className="auth-form" onSubmit={onPasswordSubmit} noValidate>
            {mode === "signup" && (
              <label>
                Name
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  autoComplete="name"
                  placeholder="Optional"
                  maxLength={80}
                />
              </label>
            )}
            <label>
              Email
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoComplete="email"
                aria-invalid={!!error}
                aria-describedby={error ? "auth-error" : undefined}
              />
            </label>
            <label>
              Password
              <input
                type="password"
                required
                minLength={8}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete={
                  mode === "signup" ? "new-password" : "current-password"
                }
                aria-invalid={!!error}
                aria-describedby={error ? "auth-error" : undefined}
              />
            </label>
            <button
              type="submit"
              className="analyze-cta"
              disabled={busy}
              aria-busy={busy}
            >
              {busy
                ? "Working…"
                : mode === "signup"
                  ? "Create account"
                  : "Log in"}
            </button>
          </form>

          {error && (
            <p id="auth-error" className="error" role="alert">
              {error}
            </p>
          )}

          <div className="auth-divider">or</div>

          <a className="google-btn" href={googleAuthURL()}>
            <svg
              className="google-btn-icon"
              viewBox="0 0 24 24"
              width="18"
              height="18"
              aria-hidden="true"
              focusable="false"
            >
              <path
                fill="#4285F4"
                d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
              />
              <path
                fill="#34A853"
                d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
              />
              <path
                fill="#FBBC05"
                d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
              />
              <path
                fill="#EA4335"
                d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
              />
            </svg>
            Continue with Google
          </a>

          <p className="auth-switch">
            {mode === "signup" ? (
              <>
                Already have an account?{" "}
                <button type="button" onClick={() => switchMode("login")}>
                  Log in
                </button>
              </>
            ) : (
              <>
                New here?{" "}
                <button type="button" onClick={() => switchMode("signup")}>
                  Sign up
                </button>
              </>
            )}
          </p>

          <p className="auth-back">
            <Link href="/">← Back to DRIPFIND</Link>
          </p>
        </div>
      </section>
    </main>
  );
}

export default function AuthPage() {
  return (
    <Suspense fallback={<SessionLoader />}>
      <AuthInner />
    </Suspense>
  );
}
