FROM node:26.12.1-alpine@sha256:40cd011b55415bc894e45b02fdd5a5dfde056921b7d23a64c74f80497af48797 AS build
WORKDIR /app
# Node 25+ no longer bundles Corepack. Keep the package manager pinned.
RUN npm install --global pnpm@10.32.1
COPY package.json pnpm-lock.yaml .npmrc* ./
RUN pnpm install --frozen-lockfile
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

FROM node:26.12.1-alpine@sha256:40cd011b55415bc894e45b02fdd5a5dfde056921b7d23a64c74f80497af48797 AS runtime
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000 HOSTNAME=0.0.0.0
RUN addgroup -S -g 1001 app && adduser -S -u 1001 -G app app
COPY --from=build --chown=app:app /app/.next/standalone ./
COPY --from=build --chown=app:app /app/.next/static ./.next/static
COPY --from=build --chown=app:app /app/public ./public
COPY --from=build --chown=app:app /app/scripts/migrate-auth.mjs /app/scripts/import-legacy-auth.mjs ./scripts/
COPY --from=build --chown=app:app /app/auth-migrations ./auth-migrations
USER app
EXPOSE 3000
HEALTHCHECK --interval=15s --timeout=3s --start-period=30s --retries=3 CMD node -e "require('http').get('http://127.0.0.1:3000/api/health',r=>process.exit(r.statusCode===200?0:1)).on('error',()=>process.exit(1))"
CMD ["node", "server.js"]
