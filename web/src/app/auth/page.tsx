"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { FormEvent, Suspense, useEffect, useMemo, useState } from "react";
import { useAuth } from "@/components/auth-provider";
import { SiteHeader } from "@/components/site-header";
import { api, googleAuthURL } from "@/lib/api";

type Mode = "login" | "signup";
// type Tab = "password" | "otp"; // email OTP — re-enable later

function AuthInner() {
  const router = useRouter();
  const params = useSearchParams();
  const { setUser, user, loading: authLoading } = useAuth();

  const initialMode = (params.get("mode") === "signup" ? "signup" : "login") as Mode;
  const oauthError = params.get("error");

  const [mode, setMode] = useState<Mode>(initialMode);
  // const [tab, setTab] = useState<Tab>("password");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  // const [otpSent, setOtpSent] = useState(false);
  // const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  // const [info, setInfo] = useState("");

  const title = mode === "signup" ? "Create your account" : "Welcome back";
  const subtitle = useMemo(() => {
    // if (tab === "otp") return "We’ll email you a one-time code — no password needed.";
    if (mode === "signup") return "Save looks, reopen history, and shop your wardrobe.";
    return "Sign in to continue finding shoppable outfits.";
  }, [mode]);

  useEffect(() => {
    if (!authLoading && user) {
      router.replace("/app");
    }
  }, [authLoading, user, router]);

  async function onPasswordSubmit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    // setInfo("");
    try {
      const res =
        mode === "signup"
          ? await api.register({ email, password, name })
          : await api.login({ email, password });
      setUser(res.user);
      router.push("/app");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }

  /*
  async function onRequestOTP(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    setInfo("");
    try {
      await api.requestOTP(email);
      setOtpSent(true);
      setInfo("Check your email for a 6-digit code. In local dev, the code is printed in the API logs.");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }

  async function onVerifyOTP(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      const res = await api.verifyOTP(email, code);
      setUser(res.user);
      router.push("/app");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy(false);
    }
  }
  */

  return (
    <main className="auth-page">
      <SiteHeader />
      <section className="auth-shell">
        <div className="auth-panel">
          <p className="eyebrow">Account</p>
          <h1>{title}</h1>
          <p className="lede narrow">{subtitle}</p>

          {oauthError && (
            <p className="error">
              Google sign-in failed{oauthError !== "google_auth_failed" ? `: ${oauthError}` : ""}.
              Try again or use email.
            </p>
          )}

          {/* Email OTP tabs — re-enable later
          <div className="auth-tabs" role="tablist">
            <button
              type="button"
              role="tab"
              aria-selected={tab === "password"}
              className={tab === "password" ? "is-active" : ""}
              onClick={() => {
                setTab("password");
                setError("");
                setInfo("");
              }}
            >
              Email & password
            </button>
            <button
              type="button"
              role="tab"
              aria-selected={tab === "otp"}
              className={tab === "otp" ? "is-active" : ""}
              onClick={() => {
                setTab("otp");
                setError("");
                setInfo("");
              }}
            >
              Email code
            </button>
          </div>
          */}

          <form className="auth-form" onSubmit={onPasswordSubmit}>
            {mode === "signup" && (
              <label>
                Name
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  autoComplete="name"
                  placeholder="Optional"
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
                autoComplete={mode === "signup" ? "new-password" : "current-password"}
              />
            </label>
            <button type="submit" className="analyze-cta" disabled={busy}>
              {busy ? "Working…" : mode === "signup" ? "Create account" : "Log in"}
            </button>
          </form>

          {/* Email OTP forms — re-enable later
          {tab === "otp" && !otpSent && (
            <form className="auth-form" onSubmit={onRequestOTP}>
              <label>
                Email
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  autoComplete="email"
                />
              </label>
              <button type="submit" className="analyze-cta" disabled={busy}>
                {busy ? "Sending…" : "Send code"}
              </button>
            </form>
          )}

          {tab === "otp" && otpSent && (
            <form className="auth-form" onSubmit={onVerifyOTP}>
              <label>
                6-digit code
                <input
                  inputMode="numeric"
                  pattern="[0-9]{6}"
                  maxLength={6}
                  required
                  value={code}
                  onChange={(e) => setCode(e.target.value.replace(/\D/g, "").slice(0, 6))}
                  autoComplete="one-time-code"
                />
              </label>
              <button type="submit" className="analyze-cta" disabled={busy}>
                {busy ? "Verifying…" : "Verify & continue"}
              </button>
              <button
                type="button"
                className="ghost-link"
                onClick={() => {
                  setOtpSent(false);
                  setCode("");
                }}
              >
                Use a different email
              </button>
            </form>
          )}
          */}

          {error && <p className="error">{error}</p>}
          {/* {info && <p className="muted">{info}</p>} */}

          <div className="auth-divider">or</div>

          <a className="google-btn" href={googleAuthURL()}>
            Continue with Google
          </a>

          <p className="auth-switch">
            {mode === "signup" ? (
              <>
                Already have an account?{" "}
                <button type="button" onClick={() => setMode("login")}>
                  Log in
                </button>
              </>
            ) : (
              <>
                New here?{" "}
                <button type="button" onClick={() => setMode("signup")}>
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
    <Suspense fallback={<main className="auth-page"><p className="muted">Loading…</p></main>}>
      <AuthInner />
    </Suspense>
  );
}
