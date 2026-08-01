# syntax=docker/dockerfile:1

# ---------- build ----------
FROM node:22-alpine AS build
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# CRITICAL: Next.js inlines NEXT_PUBLIC_* into the client bundle during
# `next build`. This MUST be a build arg — setting it at runtime in compose or
# in .env has NO effect, because the value is already compiled into the JS.
#
# Getting this wrong is the single most common way this deploy breaks. Symptoms:
# the app loads but every API call goes to localhost:8080, or the browser blocks
# them as mixed content.
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

RUN npm run build

# ---------- runtime ----------
FROM node:22-alpine
WORKDIR /app

ENV NODE_ENV=production \
    PORT=3000 \
    HOSTNAME=0.0.0.0

RUN addgroup -g 1001 nodejs && adduser -D -u 1001 -G nodejs nextjs

# Requires `output: "standalone"` in next.config.ts. Next traces exactly which
# node_modules files are reachable at runtime, taking the image from ~1GB to
# ~150MB. Smaller image means faster pulls on every deploy.
COPY --from=build --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=build --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=build --chown=nextjs:nodejs /app/public ./public

USER nextjs
EXPOSE 3000

# HOSTNAME=0.0.0.0 above is required: the standalone server binds localhost
# inside the container by default, making it unreachable from Nginx on the host.
CMD ["node", "server.js"]
