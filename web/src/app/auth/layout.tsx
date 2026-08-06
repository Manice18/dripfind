import type { Metadata } from "next";

import { constructMetaData } from "@/lib/metadata";

export const metadata: Metadata = constructMetaData({
  title: "Log in or sign up — DRIPFIND",
  description:
    "Create a DRIPFIND account or log in to turn outfit looks into shoppable wardrobes.",
  canonical: "https://dripfind.app/auth",
});

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
