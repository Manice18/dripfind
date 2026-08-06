import type { Metadata } from "next";

import { constructMetaData } from "@/lib/metadata";

export const metadata: Metadata = constructMetaData({
  title: "Studio — DRIPFIND",
  description: "Analyze outfits and shop matched products.",
  noIndex: true,
  canonical: "https://dripfind.app/studio",
});

export default function StudioLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return children;
}
