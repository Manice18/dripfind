# LOOKBOOK — AI Outfit Finder

Paste a Pinterest pin URL → detect the outfit → search multiple Indian retailers → return ranked shoppable matches.

MVP source: **Pinterest only**.

**Search providers:** Myntra, Ajio, Flipkart, Bewakoof, H&M, Snitch, Veirdo, Westside, Off Duty, Freakins, The Pant Project, The Bear House, Powerlook, Bluorng, Rare Rabbit (+ optional SerpAPI).

Not yet wired (blocked / SPA / timeout): Amazon, The Souled Store, Bonkers Corner, Uniqlo, Zara, MANGO, M&S, Savana, Newme, March Tee, Comma Casuals, Allen Solly / Louis Philippe brand sites.

## Stack

- **Frontend:** Next.js, TypeScript, Tailwind, TanStack Query
- **Backend:** Go, Chi, pgx, slog
- **Database:** PostgreSQL 17 (Docker)

## Quick start

```bash
# 1. Environment
cp .env.example .env
# Optional: set OPENAI_API_KEY for real vision analysis

# 2. Install deps (first time)
make install

# 3. Start Postgres + API + web
make start

# Stop everything
make stop
```

- App: http://localhost:3000
- API: http://localhost:8080
- Logs: `.run/api.log`, `.run/web.log`
- Adminer (optional): `make tools` → http://localhost:8081

Or run pieces separately: `make up`, `make api`, `make web`.

Without `OPENAI_API_KEY`, the API runs in **demo vision mode** (sample Old Money outfit) so the full pipeline still works.

Product search hits **live** retailers when they allow the request (Myntra, Ajio, Bewakoof, and others). Optional `SERPAPI_KEY` adds Google Shopping — useful when Flipkart or others block scrapers.

## API

| Method   | Path            | Description                            |
| -------- | --------------- | -------------------------------------- |
| `POST`   | `/analyze`      | `{ "url": "..." }` → `{ "id": "..." }` |
| `GET`    | `/result/{id}`  | Poll analysis status + outfit          |
| `GET`    | `/history`      | Recent searches                        |
| `DELETE` | `/history/{id}` | Remove history entry                   |
| `GET`    | `/health`       | Health check                           |
| `GET`    | `/images/...`   | Saved outfit images                    |

## Pipeline

1. Validate URL & extract image (Pinterest og:image / direct image)
2. Download & store locally under `backend/images/originals/`
3. Vision analysis (OpenAI or demo)
4. Deterministic search-query generation
5. Concurrent provider search across marketplaces + homegrown Shopify brands; optional SerpAPI
6. Similarity ranking + dedupe
7. Persist outfit, items, products, history

## Optional keys

- `OPENAI_API_KEY` — real outfit detection via GPT-4o-mini vision
- `SERPAPI_KEY` — live Google Shopping enrichment alongside catalog providers
