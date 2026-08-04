/**
 * Fix known-bad retailer URL shapes before opening or affiliate-wrapping.
 * Bewakoof PDPs must be /p/<slug> — older results and some APIs omit /p/.
 */
export function normalizeProductURL(
  website: string | undefined,
  rawURL: string,
): string {
  if (!rawURL) return rawURL;
  const w = (website ?? "").trim().toLowerCase();
  const u = rawURL.trim();
  if (w === "bewakoof" || u.toLowerCase().includes("bewakoof.com")) {
    return fixBewakoofURL(u);
  }
  return u;
}

function fixBewakoofURL(raw: string): string {
  try {
    const abs = raw.startsWith("http://") || raw.startsWith("https://")
      ? raw
      : `https://www.bewakoof.com/${raw.replace(/^\//, "")}`;
    const u = new URL(abs);
    let path = u.pathname || "/";
    if (!path.startsWith("/")) path = `/${path}`;
    const lower = path.toLowerCase();
    if (!lower.startsWith("/p/") && !lower.startsWith("/search")) {
      path = `/p${path === "/" ? "" : path}`;
    }
    u.pathname = path;
    u.protocol = "https:";
    u.hostname = u.hostname || "www.bewakoof.com";
    return u.toString();
  } catch {
    const slug = raw.replace(/^https?:\/\/(www\.)?bewakoof\.com\//i, "").replace(/^\//, "");
    if (!slug || slug.startsWith("p/")) {
      return `https://www.bewakoof.com/${slug}`;
    }
    return `https://www.bewakoof.com/p/${slug}`;
  }
}
