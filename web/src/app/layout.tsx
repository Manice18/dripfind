import type { Metadata } from "next";
import { Figtree, Syne } from "next/font/google";
import { Providers } from "@/components/providers";
import "./globals.css";

const syne = Syne({
  subsets: ["latin"],
  variable: "--font-display",
  weight: ["600", "700", "800"],
});

const figtree = Figtree({
  subsets: ["latin"],
  variable: "--font-body",
  weight: ["400", "500", "600"],
});

export const metadata: Metadata = {
  title: "LOOKBOOK — AI Outfit Finder",
  description: "Turn any Pinterest outfit into a shoppable wardrobe.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${syne.variable} ${figtree.variable}`}>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
