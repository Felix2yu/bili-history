FROM oven/bun:1 AS base
WORKDIR /app

FROM base AS install
COPY package.json bun.lock* ./
RUN bun install

FROM base AS build
COPY --from=install /app/node_modules ./node_modules
COPY . .
ARG NUXT_BACKEND_URL_ARG=http://backend:8899
ENV NUXT_BACKEND_URL=${NUXT_BACKEND_URL_ARG}
ENV NUXT_PUBLIC_DEFAULT_BACKEND_URL=/api
RUN bun run build

FROM node:20-alpine AS release
WORKDIR /app
COPY --from=build /app/.output ./.output

ENV HOST=0.0.0.0
ENV PORT=3000
ENV NUXT_BACKEND_URL=http://backend:8899
ENV NUXT_PUBLIC_DEFAULT_BACKEND_URL=/api
ENV NODE_ENV=production

EXPOSE 3000

CMD ["node", ".output/server/index.mjs"]
