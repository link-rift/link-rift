# ── Stage 1: Dependencies ────────────────────
FROM node:20-alpine AS deps

WORKDIR /app
COPY web/package.json web/package-lock.json ./
RUN npm ci

# ── Stage 2: Build ───────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY web/ .
RUN npm run build

# ── Stage 3: Serve ───────────────────────────
FROM nginx:alpine

COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD wget -qO- http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
