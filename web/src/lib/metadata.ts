import { Metadata } from "next";

const SITE = "https://dripfind.app";

export function constructMetaData({
  title = "DRIPFIND — AI Outfit Finder for Shoppable Wardrobes",
  description = "Drop a Pinterest pin or photo. DRIPFIND finds matching clothes across Indian retailers so you can shop the look.",
  image = "/thumbnail.jpg",
  authors = { name: "Dripfind", url: SITE },
  creator = "Dripfind",
  generator = "Next.js",
  publisher = "DRIPFIND",
  robots = "index, follow",
  canonical = SITE,
  noIndex = false,
}: {
  title?: string;
  description?: string;
  image?: string;
  authors?: { name: string; url: string };
  creator?: string;
  generator?: string;
  publisher?: string;
  robots?: string;
  canonical?: string;
  noIndex?: boolean;
} = {}): Metadata {
  const robotDirectives = noIndex
    ? { index: false, follow: false, googleBot: { index: false, follow: false } }
    : robots;

  return {
    title,
    description,
    authors,
    creator,
    generator,
    publisher,
    keywords: [
      "AI outfit finder",
      "shop the look",
      "Pinterest outfit search",
      "Indian fashion shopping",
      "Myntra",
      "Ajio",
      "Flipkart",
      "DRIPFIND",
    ],
    category: "fashion",
    metadataBase: new URL(SITE),
    alternates: {
      canonical,
      types: {
        "text/plain": [
          { url: `${SITE}/llms.txt`, title: "llms.txt" },
          { url: `${SITE}/llms-full.txt`, title: "llms-full.txt" },
        ],
      },
    },
    openGraph: {
      title,
      description,
      url: canonical,
      siteName: "DRIPFIND",
      locale: "en_IN",
      type: "website",
      images: [
        {
          url: image,
          width: 1200,
          height: 630,
          alt: "DRIPFIND — AI outfit finder",
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      title,
      description,
      images: [image],
    },
    robots: robotDirectives,
  };
}

/** JSON-LD for the public marketing homepage. */
export const landingJsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "WebSite",
      "@id": `${SITE}/#website`,
      url: SITE,
      name: "DRIPFIND",
      description:
        "AI outfit finder that turns a Pinterest pin or photo into a shoppable wardrobe across Indian retailers.",
      publisher: { "@id": `${SITE}/#organization` },
      inLanguage: "en-IN",
    },
    {
      "@type": "Organization",
      "@id": `${SITE}/#organization`,
      name: "DRIPFIND",
      url: SITE,
      logo: `${SITE}/thumbnail.jpg`,
    },
    {
      "@type": "SoftwareApplication",
      "@id": `${SITE}/#app`,
      name: "DRIPFIND",
      applicationCategory: "LifestyleApplication",
      operatingSystem: "Web",
      url: SITE,
      description:
        "Paste a Pinterest pin or drop an outfit photo. DRIPFIND identifies clothing pieces and finds matching products on Indian retailers.",
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "INR",
      },
      featureList: [
        "Pinterest URL and photo upload analysis",
        "Per-item clothing detection",
        "Product search across Indian retailers",
        "Shoppable outfit results with prices and links",
      ],
    },
  ],
} as const;
