/**
 * inr.deals CPS partners. Wrap product card hrefs for these retailers.
 * Tracking id from https://inr.deals (account man779003531).
 */
const INR_DEALS_ID = "man779003531";

function partnerKey(
  website: string | undefined,
  rawURL: string,
): string | null {
  const w = (website ?? "").trim().toLowerCase();
  const u = (rawURL ?? "").trim().toLowerCase();

  if (w === "myntra" || u.includes("myntra.com")) return "myntra";
  if (w === "ajio" || u.includes("ajio.com")) return "ajio";
  if (w === "hm" || u.includes("www2.hm.com") || u.includes("hm.com/"))
    return "hm";
  return null;
}

/** True when this product should go through inr.deals. */
export function isAffiliateProduct(
  website: string | undefined,
  rawURL: string,
): boolean {
  return partnerKey(website, rawURL) !== null;
}

/** Returns an inr.deals tracking URL for enrolled merchants; otherwise the original URL. */
export function affiliateProductURL(
  website: string | undefined,
  rawURL: string,
): string {
  if (!rawURL) return rawURL;
  if (rawURL.includes("inr.deals/track")) return rawURL;
  if (!partnerKey(website, rawURL)) return rawURL;

  const params = new URLSearchParams({
    id: INR_DEALS_ID,
    src: "merchant-detail-backend",
    campaign: "cps",
    url: rawURL,
  });
  return `https://inr.deals/track?${params.toString()}`;
}
