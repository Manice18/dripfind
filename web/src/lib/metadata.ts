import { Metadata } from "next";

export function constructMetaData({
  title = "DRIPFIND — AI Outfit Finder for Shoppable Wardrobes",
  description = "Drop a Pinterest pin or photo. DRIPFIND finds matching clothes across Indian retailers so you can shop the look.",
  image = "/thumbnail.jpg",
  authors = { name: "Dripfind", url: "https://dripfind.app" },
  creator = "Dripfind",
  generator = "Next.js",
  publisher = "DRIPFIND",
  robots = "index, follow",
}: {
  title?: string;
  description?: string;
  image?: string;
  authors?: { name: string; url: string };
  creator?: string;
  generator?: string;
  publisher?: string;
  robots?: string;
} = {}): Metadata {
  return {
    title,
    description,
    authors,
    creator,
    generator,
    publisher,
    openGraph: {
      title,
      description,
      url: "https://dripfind.app",
      siteName: "DRIPFIND",
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
    metadataBase: new URL("https://dripfind.app"),
    robots,
  };
}
