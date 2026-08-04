import { Metadata } from "next";

export function constructMetaData({
  title = "DRIPFIND — AI Outfit Finder",
  description = "Turn any Pinterest outfit into a shoppable wardrobe.",
  image = "/thumbnail.png",
  authors = { name: "Manice18", url: "https://dripfind.app" },
  creator = "Manice18",
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
        },
      ],
    },
    twitter: {
      card: "summary_large_image",
      site: "@Manice18heree",
      creator: "@Manice18heree",
      title,
      description,
      images: [image],
    },
    metadataBase: new URL("https://dripfind.app"),
    robots,
  };
}
