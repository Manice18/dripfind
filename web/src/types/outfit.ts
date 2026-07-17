export type OutfitStatus = "pending" | "processing" | "completed" | "failed";

export interface Product {
  id: string;
  title: string;
  brand: string;
  price: number;
  currency: string;
  website: string;
  url: string;
  image: string;
  match_score: number;
}

export interface ClothingItem {
  id: string;
  category: string;
  color: string;
  material: string;
  fit: string;
  pattern: string;
  confidence: number;
  search_query?: string;
  products?: Product[];
}

export interface Outfit {
  id: string;
  image_path: string;
  image_url?: string;
  source_url: string;
  style: string;
  gender: string;
  season: string;
  occasion: string;
  status: OutfitStatus;
  error_message?: string;
  items?: ClothingItem[];
  created_at: string;
  updated_at: string;
}

export interface HistoryEntry {
  id: string;
  outfit_id: string;
  source_url: string;
  style?: string;
  gender?: string;
  status?: string;
  image_url?: string;
  created_at: string;
}
