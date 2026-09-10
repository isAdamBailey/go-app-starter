FROM node:26-alpine AS base
WORKDIR /app

FROM base AS deps
# Node 22's bundled npm (10.x) has a known arborist crash on some dependency
# graphs (npm/cli#7724-class bug); npm 11+ fixes it.
RUN npm install -g npm@latest
COPY package.json package-lock.json ./
RUN npm ci

FROM deps AS build
COPY . .
RUN npm run build

FROM base AS runtime
ENV NODE_ENV=production
RUN addgroup -S nuxt && adduser -S nuxt -G nuxt
COPY --from=build --chown=nuxt:nuxt /app/.output ./.output
USER nuxt
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
