"use client";

import posthog from "posthog-js";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import { api, type AuthUser } from "@/lib/api";

type AuthContextValue = {
  user: AuthUser | null;
  loading: boolean;
  refresh: () => Promise<void>;
  setUser: (user: AuthUser | null) => void;
  logout: () => Promise<void>;
  deleteAccount: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [loading, setLoading] = useState(true);

  const setAuthenticatedUser = useCallback((nextUser: AuthUser | null) => {
    setUser(nextUser);

    if (nextUser?.id) {
      posthog.identify(nextUser.id, {
        email: nextUser.email,
        name: nextUser.name,
      });
    }
  }, []);

  const refresh = useCallback(async () => {
    try {
      const data = await api.me();
      setAuthenticatedUser(data.user);
    } catch {
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, [setAuthenticatedUser]);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await api.me();
        if (!cancelled) setAuthenticatedUser(data.user);
      } catch {
        if (!cancelled) setUser(null);
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [setAuthenticatedUser]);

  const logout = useCallback(async () => {
    try {
      await api.logout();
      posthog.capture("logout_completed");
    } finally {
      posthog.reset();
      setUser(null);
    }
  }, []);

  const deleteAccount = useCallback(async () => {
    await api.deleteAccount();
    posthog.reset();
    setUser(null);
  }, []);

  const value = useMemo(
    () => ({
      user,
      loading,
      refresh,
      setUser: setAuthenticatedUser,
      logout,
      deleteAccount,
    }),
    [user, loading, refresh, setAuthenticatedUser, logout, deleteAccount],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
