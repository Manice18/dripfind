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
            <Link href="/">← Back to LOOKBOOK</Link>
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
