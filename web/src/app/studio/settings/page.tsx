"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth-provider";
import { IconLogout } from "@/components/icons";
import { SessionLoader } from "@/components/session-loader";
import { SiteHeader } from "@/components/site-header";

export default function SettingsPage() {
  const { user, loading, logout } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !user) {
      router.replace("/auth?mode=login");
    }
  }, [loading, user, router]);

  if (loading || !user) {
    return <SessionLoader />;
  }

  const displayName = user.name || "Account";
  const joined = user.created_at
    ? new Date(user.created_at).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      })
    : null;

  return (
    <main className="workspace settings-page">
      <SiteHeader variant="studio" />

      <section className="settings-section" aria-labelledby="settings-heading">
        <h1 id="settings-heading">Account</h1>
        <p className="lede narrow">Your LOOKBOOK profile and session.</p>

        <dl className="settings-list">
          <div>
            <dt>Name</dt>
            <dd>{displayName}</dd>
          </div>
          <div>
            <dt>Email</dt>
            <dd>{user.email}</dd>
          </div>
          {joined && (
            <div>
              <dt>Joined</dt>
              <dd>{joined}</dd>
            </div>
          )}
        </dl>

        <button
          type="button"
          className="ghost settings-logout"
          onClick={() =>
            void logout().then(() => {
              window.location.href = "/";
            })
          }
        >
          <IconLogout />
          Log out
        </button>
      </section>
    </main>
  );
}
