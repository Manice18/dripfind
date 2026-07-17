CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS outfits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_path TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL,
    style TEXT NOT NULL DEFAULT '',
    gender TEXT NOT NULL DEFAULT '',
    season TEXT NOT NULL DEFAULT '',
    occasion TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'pending',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS clothing_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outfit_id UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    category TEXT NOT NULL DEFAULT '',
    color TEXT NOT NULL DEFAULT '',
    material TEXT NOT NULL DEFAULT '',
    fit TEXT NOT NULL DEFAULT '',
    pattern TEXT NOT NULL DEFAULT '',
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    search_query TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID NOT NULL REFERENCES clothing_items(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '',
    brand TEXT NOT NULL DEFAULT '',
    price DOUBLE PRECISION NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'INR',
    website TEXT NOT NULL DEFAULT '',
    url TEXT NOT NULL DEFAULT '',
    image TEXT NOT NULL DEFAULT '',
    match_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outfit_id UUID NOT NULL REFERENCES outfits(id) ON DELETE CASCADE,
    source_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clothing_items_outfit_id ON clothing_items(outfit_id);
CREATE INDEX IF NOT EXISTS idx_products_item_id ON products(item_id);
CREATE INDEX IF NOT EXISTS idx_history_created_at ON history(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_outfits_created_at ON outfits(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_outfits_status ON outfits(status);
