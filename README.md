# DRIPFIND — AI Outfit Finder

Paste a Pinterest pin URL → detect the outfit → search multiple Indian retailers → return ranked shoppable matches.

Also supports **photo upload** (drag-and-drop / screenshot / file picker).

MVP sources: **Pinterest URL** or **uploaded image**.

**Search providers:** Myntra, Ajio, Flipkart, Bewakoof, H&M, Snitch, Veirdo, Westside, Off Duty, Freakins, The Pant Project, The Bear House, Powerlook, Bluorng, Rare Rabbit (+ optional SerpAPI).

Not yet wired (blocked / SPA / timeout): Amazon, The Souled Store, Bonkers Corner, Uniqlo, Zara, MANGO, M&S, Savana, Newme, March Tee, Comma Casuals, Allen Solly / Louis Philippe brand sites.

## Stack

- **Frontend:** Next.js, TypeScript, Tailwind, TanStack Query
- **Backend:** Go API + worker, Chi, pgx, slog
- **Database:** PostgreSQL 17 (Docker)
- **Object storage:** MinIO / S3 (`STORAGE_BACKEND=s3` by default; `local` for disk)

## Architecture

Two Go processes share Postgres and object storage. The **API** handles auth and enqueues work; the **worker** runs the heavy pipeline asynchronously.

```text
┌─────────────┐     session cookie      ┌──────────────────┐
│  Next.js    │ ───────────────────────▶│  Go API (:8080)  │
│  :3000      │◀── poll /result/{id} ───│  Chi + auth      │
└─────────────┘                         └────────┬─────────┘
                                                 │
                    create outfit + enqueue job  │
                                                 ▼
                                        ┌────────────────┐
                                        │   PostgreSQL   │
                                        │ outfits, jobs, │
                                        │ users, history │
                                        └────────┬───────┘
                                                 │
                                          claim job│
                                                 ▼
┌──────────────┐   read/write images   ┌──────────────────┐
│ MinIO / S3   │◀──────────────────────│  Go worker       │
└──────────────┘                       │  pipeline runner │
                                       └────────┬─────────┘
                                                │
              ┌─────────────────────────────────┼──────────────────────────┐
              ▼                                 ▼                          ▼
     Pinterest extract                  OpenAI vision              Retail providers
     + image download                   (or demo mode)             (Myntra, Ajio, …)
                                                                       │
                                                                       ▼
                                                                  Rank + dedupe
                                                                       │
                                                                       ▼
                                                              Persist results → DB
```

### Request flow

1. **Sign in** — browser hits `/auth/*`; API sets a DB-backed session cookie.
2. **Submit** — `POST /analyze` (Pinterest URL) or `POST /analyze/upload` (image file).
3. **API (fast path)** — validates input, creates an outfit row (`pending`), for uploads saves the image to storage, enqueues a job in Postgres, returns `{ "id": "..." }`.
4. **Worker (async)** — claims the job, marks the outfit `processing`, then:
   - **URL jobs:** extract Pinterest `og:image` → download → save to storage
   - **Upload jobs:** open the already-stored image key
   - Vision analysis → per-item search queries → concurrent marketplace/Shopify search → rank/dedupe → save outfit items + products (`completed` / `failed`)
5. **Poll** — frontend calls `GET /result/{id}` until status is terminal; history is loaded via `GET /history`.

```text
Browser                API                   Postgres              Worker              Storage / External
  │                     │                       │                    │                      │
  │── POST /analyze ───▶│                       │                    │                      │
  │                     │── insert outfit ─────▶│                    │                      │
  │                     │── enqueue job ───────▶│                    │                      │
  │◀──── { id } ────────│                       │                    │                      │
  │                     │                       │◀── claim job ──────│                      │
  │                     │                       │                    │── extract/download ─▶│
  │                     │                       │                    │── vision (OpenAI) ──▶│
  │                     │                       │                    │── search retailers ─▶│
  │                     │                       │◀── save results ───│                      │
  │── GET /result/id ──▶│── read outfit ───────▶│                    │                      │
  │◀── status+outfit ───│                       │                    │                      │
```

### Component roles

| Piece | Responsibility |
| ----- | -------------- |
| **Web** | Auth UI, submit URL/upload, poll results, show ranked products |
| **API** | Sessions, analyze/upload entrypoints, job enqueue, result/history reads |
| **Worker** | Durable job claim/retry, full outfit pipeline |
| **Postgres** | Users, sessions, outfits, items, products, job queue |
| **MinIO/S3** | Outfit images (API serves `/images/*` only when `STORAGE_BACKEND=local`) |
| **Vision** | Detect clothing items + style metadata (`OPENAI_API_KEY` or demo) |
| **Search + rank** | Concurrent provider fetch, similarity scoring, dedupe |

## Quick start

```bash
# 1. Environment
cp .env.example .env
# Optional: set OPENAI_API_KEY for real vision analysis

# 2. Install deps (first time)
make install

# 3. Start Postgres + MinIO + API + worker + web
make start

# Stop everything
make stop
```

- App: http://localhost:3000
- API: http://localhost:8080
- MinIO API: http://localhost:9000 (console: http://localhost:9001)
- Logs: `.run/api.log`, `.run/worker.log`, `.run/web.log`
- Adminer (optional): `make tools` → http://localhost:8081

Or run pieces separately: `make up`, `make api`, `make worker`, `make web`.

Without `OPENAI_API_KEY`, the API runs in **demo vision mode** (sample Old Money outfit) so the full pipeline still works.

Product search hits **live** retailers when they allow the request (Myntra, Ajio, Bewakoof, and others). Optional `SERPAPI_KEY` adds Google Shopping — useful when Flipkart or others block scrapers.

Analyze and history endpoints require a signed-in session (email/password, OTP, or Google OAuth).

## API

| Method   | Path              | Description                                |
| -------- | ----------------- | ------------------------------------------ |
| `POST`   | `/analyze`        | `{ "url": "..." }` → `{ "id": "..." }`     |
| `POST`   | `/analyze/upload` | multipart `image` file → `{ "id": "..." }` |
| `GET`    | `/result/{id}`    | Poll analysis status + outfit              |
| `GET`    | `/history`        | Recent searches                            |
| `DELETE` | `/history/{id}`   | Remove history entry                       |
| `GET`    | `/health`         | Health check                               |
| `GET`    | `/images/...`     | Saved outfit images (local storage only)   |

Auth: `POST /auth/register`, `POST /auth/login`, `POST /auth/logout`, `GET /auth/me`, `POST /auth/otp/request`, `POST /auth/otp/verify`, `GET /auth/google` (+ callback).

## Pipeline

1. Validate URL & extract image (Pinterest og:image / direct image)
2. Download & store in object storage (MinIO/S3, or local disk)
3. Enqueue job → worker runs vision analysis (OpenAI or demo)
4. Deterministic search-query generation
5. Concurrent provider search across marketplaces + homegrown Shopify brands; optional SerpAPI
6. Similarity ranking + dedupe
7. Persist outfit, items, products, history

## Optional keys

- `OPENAI_API_KEY` — real outfit detection via GPT-4o-mini vision
- `SERPAPI_KEY` — live Google Shopping enrichment alongside catalog providers
- `GOOGLE_CLIENT_ID` / `GOOGLE_CLIENT_SECRET` — Google sign-in
- SMTP vars — email OTP (without them, OTP codes are logged by the API)
- `STORAGE_BACKEND=local` — keep images on disk instead of MinIO/S3
