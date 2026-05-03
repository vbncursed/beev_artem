# --- stage 1: build the SPA ----------------------------------------------
# node:22 is current LTS; alpine keeps the builder small. Yarn is the
# repo's chosen package manager (yarn.lock present), --frozen-lockfile
# enforces reproducible installs across CI / local / prod.
FROM node:22-alpine AS builder

WORKDIR /app

# Install deps first — separate COPY exists so changes to source don't
# bust the dependency layer cache. yarn.lock pins everything.
COPY frontend/package.json frontend/yarn.lock ./
RUN yarn install --frozen-lockfile

# Bring in the rest of the source and build. `yarn build` runs `tsc -b`
# then `vite build` per package.json scripts; output lands in /app/dist.
COPY frontend/ ./
RUN yarn build


# --- stage 2: serve via nginx --------------------------------------------
# nginx serves the static bundle and proxies /api → gateway:8080 inside
# the compose network so the SPA is same-origin (no CORS at runtime).
FROM nginx:1.27-alpine

# Drop the stock default config and ship ours from the repo so changes
# go through review.
RUN rm /etc/nginx/conf.d/default.conf
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

# Use the stock entrypoint; nginx -g "daemon off;" runs in foreground
# so docker keeps the container alive.
