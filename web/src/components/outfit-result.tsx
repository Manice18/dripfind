"use client";

import type { ClothingItem, Outfit, Product } from "@/types/outfit";
import { mediaURL } from "@/lib/api";

function formatPrice(p: Product) {
  try {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: p.currency || "INR",
      maximumFractionDigits: 0,
    }).format(p.price);
  } catch {
    return `₹${Math.round(p.price)}`;
  }
}

function ProductCard({ product }: { product: Product }) {
  return (
    <a
      className="product-card"
      href={product.url}
      target="_blank"
      rel="noopener noreferrer"
    >
      <div className="product-media">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src={product.image} alt={product.title} loading="lazy" />
        <span className="match">{Math.round(product.match_score)}% match</span>
      </div>
      <div className="product-meta">
        <p className="brand">{product.brand}</p>
        <h4>{product.title}</h4>
        <div className="row">
          <strong>{formatPrice(product)}</strong>
          <span className="site">{product.website}</span>
        </div>
      </div>
    </a>
  );
}

function ItemBlock({ item }: { item: ClothingItem }) {
  return (
    <section className="item-block">
      <header>
        <div>
          <p className="eyebrow">{item.category}</p>
          <h3>
            {item.color} {item.material} · {item.fit}
          </h3>
          <p className="muted">
            {item.pattern} pattern · {Math.round(item.confidence * 100)}% confidence
          </p>
        </div>
      </header>
      <div className="product-grid">
        {(item.products ?? []).length === 0 ? (
          <p className="muted">No products found for this piece.</p>
        ) : (
          (item.products ?? []).map((p) => (
            <ProductCard key={p.id} product={p} />
          ))
        )}
      </div>
    </section>
  );
}

export function OutfitResult({ outfit }: { outfit: Outfit }) {
  return (
    <div className="result-layout">
      <aside className="result-summary">
        <div className="summary-media">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src={mediaURL(outfit.image_url)} alt="Analyzed outfit" />
        </div>
        <div className="summary-copy">
          <p className="eyebrow">Detected look</p>
          <h2>{outfit.style || "Untitled look"}</h2>
          <p>
            {outfit.gender} · {outfit.season} · {outfit.occasion}
          </p>
          <ul>
            {(outfit.items ?? []).map((item) => (
              <li key={item.id}>
                <span>{item.category}</span>
                <em>{item.color}</em>
              </li>
            ))}
          </ul>
        </div>
      </aside>

      <div className="result-items">
        {(outfit.items ?? []).map((item) => (
          <ItemBlock key={item.id} item={item} />
        ))}
      </div>
    </div>
  );
}
