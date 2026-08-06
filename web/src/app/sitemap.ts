import type { MetadataRoute } from "next";

const SITE = "https://dripfind.app";

export default function sitemap(): MetadataRoute.Sitemap {
  const lastModified = new Date();

  return [
    {
      url: SITE,
      lastModified,
      changeFrequency: "weekly",
      priority: 1,
    },
    {
      url: `${SITE}/auth`,
      lastModified,
      changeFrequency: "monthly",
      priority: 0.6,
    },
  ];
}
