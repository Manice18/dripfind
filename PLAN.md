# 👕 AI Outfit Finder

### Turn Any Pinterest Outfit into a Shoppable Wardrobe

> **Goal**
>
> Paste a Pinterest URL →
> Download the image →
> AI understands the outfit →
> Search multiple shopping websites →
> Rank similar products →
> Return a complete shoppable outfit.

---

# 🎯 Vision

Unlike Google Lens or Pinterest Lens, this application is **fashion-first**.

Instead of answering:

> "What is this?"

It answers:

> "How can I recreate this entire outfit?"

The system should:

- Detect every clothing item
- Understand style and aesthetics
- Search multiple shopping websites
- Recommend similar products
- Rank by similarity
- Build complete outfits

---

# 🏗 System Architecture

For the MVP, we'll build a **Modular Monolith**.

This gives us:

- Simple development
- Easy debugging
- Easy deployment later
- Can evolve into microservices if needed

```text
                    Next.js

                       │
                  REST API
                       │
          ┌────────────────────────────┐
          │                            │
          │      Go Backend            │
          │                            │
          ├────────────────────────────┤
          │                            │
          │ URL Extraction             │
          │ Image Downloader           │
          │ Vision Analysis            │
          │ Query Generator            │
          │ Product Search             │
          │ Ranking Engine             │
          │ History                    │
          │                            │
          └────────────┬───────────────┘
                       │
                  PostgreSQL
                 (Docker Container)

                       │

               Local Image Storage
```

---

# 🧱 Technology Stack

## Frontend

- Next.js
- TypeScript
- TailwindCSS
- shadcn/ui
- TanStack Query

---

## Backend

- Go 1.24+
- Chi Router
- pgx
- sqlc
- slog

---

## Database

PostgreSQL 17

Running inside Docker.

---

## Database Migrations

golang-migrate

---

## AI

Initially

OpenAI Vision

Later can support

- Gemini
- Claude Vision

through an interface.

---

## Image Storage

For MVP

```
backend/images/
```

Later

```
S3
Cloudflare R2
GCS
```

---

# 📁 Folder Structure

```text
ai-outfit-finder/

│
├── backend/
│
│   ├── cmd/
│   │    └── api/
│   │         └── main.go
│   │
│   ├── internal/
│   │
│   │    ├── api/
│   │    │      routes.go
│   │    │      middleware.go
│   │    │
│   │    ├── config/
│   │    │
│   │    ├── database/
│   │    │
│   │    ├── models/
│   │    │
│   │    ├── pinterest/
│   │    │
│   │    ├── image/
│   │    │
│   │    ├── vision/
│   │    │
│   │    ├── prompt/
│   │    │
│   │    ├── search/
│   │    │
│   │    │     provider.go
│   │    │     amazon.go
│   │    │     myntra.go
│   │    │     ajio.go
│   │    │
│   │    ├── ranking/
│   │    │
│   │    ├── history/
│   │    │
│   │    └── utils/
│   │
│   ├── images/
│   │
│   │     ├── originals/
│   │     └── processed/
│   │
│   ├── migrations/
│   │
│   ├── sql/
│   │
│   └── go.mod
│
├── web/
│
│   ├── app/
│   ├── components/
│   ├── hooks/
│   ├── lib/
│   └── types/
│
├── docker/
│   └── postgres/
│       └── init.sql
│
├── docker-compose.yml
│
├── Makefile
│
├── .env
│
└── README.md
```

---

# 🔄 Complete Request Flow

```text
User

↓

Paste Pinterest URL

↓

POST /analyze

↓

Backend validates URL

↓

Extract image URL

↓

Download image

↓

Store locally

↓

Vision AI analyzes image

↓

Generate structured JSON

↓

Generate search queries

↓

Search providers concurrently

↓

Merge products

↓

Rank products

↓

Store history

↓

Return response

↓

Frontend renders outfit
```

---

# 📦 Pipeline Architecture

Think of the backend as a pipeline.

```text
Input URL

        │

        ▼

Image Acquisition

        │

        ▼

Image Processing

        │

        ▼

Vision Understanding

        │

        ▼

Structured Fashion Model

        │

        ▼

Search Query Generation

        │

        ▼

Provider Search

        │

        ▼

Result Normalization

        │

        ▼

Ranking

        │

        ▼

Response Builder

        │

        ▼

Frontend
```

Each stage should know nothing about the next stage.

---

# 1️⃣ URL Extraction Module

Input

```
Pinterest URL
```

Responsibilities

- Validate URL
- Resolve redirects
- Extract metadata
- Obtain image URL

Output

```
Image URL
```

---

# 2️⃣ Image Downloader

Responsibilities

- Download image
- Save locally
- Generate UUID filename
- Verify image format

Output

```
images/originals/image-123.jpg
```

---

# 3️⃣ Vision Module

Input

```
Image Path
```

Output

```json
{
  "gender": "male",
  "style": "Old Money",
  "season": "Summer",
  "occasion": "Casual",
  "items": [
    {
      "category": "Shirt",
      "color": "White",
      "material": "Linen",
      "fit": "Oversized",
      "pattern": "Solid",
      "confidence": 0.97
    }
  ]
}
```

The Vision module ONLY understands clothes.

It does NOT search.

---

# 4️⃣ Search Query Generator

Converts

```json
{
  "category": "Shirt",
  "color": "White",
  "material": "Linen",
  "fit": "Oversized"
}
```

into

```
White Oversized Linen Shirt Men
```

The AI should **not** generate search strings.

---

# 5️⃣ Product Provider Layer

Every shopping website implements

```go
type Provider interface {
    Search(
        ctx context.Context,
        item ClothingItem,
    ) ([]Product, error)
}
```

Example

```
AmazonProvider

↓

MyntraProvider

↓

AjioProvider

↓

FlipkartProvider
```

Adding a new provider should only require implementing this interface.

---

# 6️⃣ Concurrent Search

Search every provider simultaneously.

```text
                    Shirt

                      │

      ┌───────────────┼───────────────┐

      ▼               ▼               ▼

 Amazon         Myntra          Ajio

      ▼               ▼               ▼

      └───────────────┼───────────────┘

              Merge Results
```

Use goroutines and channels.

No Redis.

No workers.

---

# 7️⃣ Result Normalization

Different providers return different schemas.

Normalize everything into

```go
type Product struct {
    Title string
    Brand string
    Price float64
    URL string
    Image string
    Website string
}
```

---

# 8️⃣ Ranking Engine

Input

```
Products
```

Output

```
Sorted Products
```

Scoring factors

- Color
- Material
- Fit
- Pattern
- Similarity
- Brand
- Price

Example

```
96%

91%

87%

83%
```

---

# 🗄 Database Schema

## outfits

```sql
id

image_path

source_url

style

gender

season

occasion

created_at
```

---

## clothing_items

```sql
id

outfit_id

category

color

material

fit

pattern

confidence
```

---

## products

```sql
id

item_id

title

brand

price

currency

website

url

image

match_score
```

---

## history

```sql
id

source_url

created_at
```

---

# 🌐 REST API

## Analyze Outfit

```
POST /analyze
```

Body

```json
{
  "url": "..."
}
```

Response

```json
{
  "id": "123"
}
```

---

## Get Result

```
GET /result/{id}
```

---

## Search History

```
GET /history
```

---

## Delete History

```
DELETE /history/{id}
```

---

# 📷 Local Image Storage

```
backend/images/

    originals/

        uuid.jpg

        uuid2.jpg

    processed/

        shirt.png

        shoes.png
```

Later

```
ImageStorage

↓

LocalStorage

↓

S3Storage
```

---

# 🐳 Docker

Only PostgreSQL is containerized.

```
docker compose up -d
```

Starts

- PostgreSQL

Optional

- Adminer

Nothing else.

---

# 📋 Development Workflow

```text
make up

↓

Postgres starts

↓

make migrate-up

↓

go run cmd/api/main.go

↓

cd web

↓

npm run dev
```

---

# 🔍 Logging

Use

```
log/slog
```

Log

- Incoming request
- Image downloaded
- AI latency
- Provider latency
- Ranking duration

---

# ⚙️ Configuration

Use

```
.env
```

Example

```
POSTGRES_HOST=localhost

POSTGRES_PORT=5432

POSTGRES_DB=outfitfinder

OPENAI_API_KEY=...
```

---

# 🧩 Interfaces

## Vision

```go
type Analyzer interface {
    Analyze(
        imagePath string,
    ) (*Outfit, error)
}
```

---

## Provider

```go
type Provider interface {
    Search(
        ctx context.Context,
        item ClothingItem,
    ) ([]Product, error)
}
```

---

## Ranking

```go
type Ranker interface {
    Rank(
        products []Product,
    ) []Product
}
```

---

## Image Storage

```go
type ImageStorage interface {
    Save(
        image []byte,
    ) (string, error)
}
```

Current implementation

```
LocalStorage
```

Future

```
S3Storage
```

---

# 🚀 Development Roadmap

## Phase 1 — Foundation

- [x] Project setup
- [x] Docker PostgreSQL
- [x] Database migrations
- [x] Next.js setup
- [x] Go API
- [x] Basic routing

---

## Phase 2 — Image Pipeline

- [x] Validate URL
- [x] Extract Pinterest image
- [x] Download image
- [x] Save locally

---

## Phase 3 — AI

- [x] Integrate Vision AI
- [x] Build prompts
- [x] Parse structured JSON
- [x] Display detected outfit

---

## Phase 4 — Product Search

- [x] Provider abstraction
- [x] Amazon provider
- [x] Myntra provider
- [x] Ajio provider
- [x] Concurrent search
- [x] Result normalization

---

## Phase 5 — Ranking

- [x] Similarity scoring
- [x] Price-aware ranking
- [x] Best match selection
- [x] Deduplication

---

## Phase 6 — Frontend

- [x] Landing page
- [x] Loading state
- [x] Outfit summary
- [x] Product cards
- [x] Search history

---

# 📈 Future Improvements (Post-MVP)

- Browser extension
- User authentication
- Saved outfits
- Wishlist
- AI stylist ("Make this outfit more formal")
- Budget-based outfit generation
- Wardrobe matching ("You already own similar trousers")
- Additional sources (Instagram, TikTok, screenshots)
- Affiliate link integration
- Vector embeddings for semantic product similarity
- Background jobs and Redis for long-running analyses
- Cloud object storage (S3/R2)
- Monitoring and observability

---

# 💡 Core Design Principles

1. **Pipeline over tightly coupled logic** — each stage has one responsibility.
2. **Interfaces over concrete implementations** — Vision, Providers, Ranking, and Storage should all be swappable.
3. **Provider abstraction** — adding a new retailer should only require implementing the `Provider` interface.
4. **AI for understanding, deterministic code for searching** — let the LLM describe the outfit, but let your code generate search queries and ranking.
5. **Modular Monolith first** — keep everything in one codebase until scale justifies extracting services.

This architecture gives you a clean MVP while leaving clear extension points for future capabilities without requiring major rewrites.
