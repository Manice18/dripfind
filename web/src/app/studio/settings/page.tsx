"use client";

import { useEffect, useId, useRef, useState } from "react";
import { useRouter } from "next/navigation";

import { useAuth } from "@/components/auth-provider";
import { IconLogout } from "@/components/icons";
import { SessionLoader } from "@/components/session-loader";
import { SiteHeader } from "@/components/site-header";

export default function SettingsPage() {
  const { user, loading, logout, deleteAccount } = useAuth();
  const router = useRouter();
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const dialogRef = useRef<HTMLDialogElement>(null);
  const titleId = useId();
  const bodyId = useId();

  useEffect(() => {
    if (!loading && !user) {
      router.replace("/auth?mode=login");
    }
  }, [loading, user, router]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (confirmDelete) {
      if (!dialog.open) dialog.showModal();
    } else if (dialog.open) {
      dialog.close();
    }
  }, [confirmDelete]);

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

  function closeDeleteDialog() {
    if (busy) return;
    setConfirmDelete(false);
    setError("");
  }

  async function onDeleteAccount() {
    setBusy(true);
    setError("");
    try {
      await deleteAccount();
      window.location.href = "/";
    } catch (err) {
      setError((err as Error).message || "Couldn’t delete account");
      setBusy(false);
    }
  }

  return (
    <main className="workspace settings-page">
      <SiteHeader variant="studio" />

      <section className="settings-section" aria-labelledby="settings-heading">
        <h1 id="settings-heading">Account</h1>
        <p className="lede narrow">Your DRIPFIND profile and session.</p>

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

        <div className="settings-actions">
          <div className="settings-action-row">
            <button
              type="button"
              className="ghost settings-logout"
              disabled={busy}
              onClick={() =>
                void logout().then(() => {
                  window.location.href = "/";
                })
              }
            >
              <IconLogout />
              Log out
            </button>
            <button
              type="button"
              className="ghost settings-delete"
              disabled={busy}
              onClick={() => {
                setConfirmDelete(true);
                setError("");
              }}
            >
              Delete account
            </button>
          </div>
        </div>
      </section>

      <dialog
        ref={dialogRef}
        className="settings-delete-dialog"
        aria-labelledby={titleId}
        aria-describedby={bodyId}
        onCancel={(event) => {
          event.preventDefault();
          closeDeleteDialog();
        }}
        onClose={() => {
          setConfirmDelete(false);
          setError("");
        }}
        onClick={(event) => {
          if (event.target === dialogRef.current) {
            closeDeleteDialog();
          }
        }}
      >
        <div className="settings-delete-dialog-panel">
          <p id={titleId} className="settings-delete-title">
            Warning: permanent deletion
          </p>
          <div id={bodyId}>
            <p>
              Deleting your account will permanently remove everything tied to{" "}
              <strong>{user.email}</strong>:
            </p>
            <ul>
              <li>Your profile and sign-in access</li>
              <li>All search history</li>
              <li>Every outfit analysis and match result</li>
            </ul>
            <p className="settings-delete-irreversible">
              This cannot be undone. You will need to create a new account to
              use DRIPFIND again.
            </p>
          </div>
          {error && (
            <p className="error" role="alert">
              {error}
            </p>
          )}
          <div className="settings-delete-row">
            <button
              type="button"
              className="ghost"
              disabled={busy}
              onClick={closeDeleteDialog}
            >
              Keep my account
            </button>
            <button
              type="button"
              className="danger-cta"
              disabled={busy}
              aria-busy={busy}
              onClick={() => void onDeleteAccount()}
            >
              {busy ? "Deleting…" : "Yes, delete forever"}
            </button>
          </div>
        </div>
      </dialog>
    </main>
  );
}
