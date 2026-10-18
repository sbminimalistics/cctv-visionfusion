# build image (compiles typescript to javascript)
FROM node:26-alpine AS builder

WORKDIR /app

# copy dependency manifests first to leverage docker layer caching
COPY package*.json ./


# install dependencies
RUN npm ci

# copy source files
# COPY tsconfig.json ./
COPY src/ ./src/
COPY public/ ./public/

# compile typescript to javascript
RUN npm run build

# prune node_modules to remove heavy build-time devdependencies
RUN npm prune --production

# runtime image (hardened production execution layer)
FROM node:26-alpine AS runner

WORKDIR /app

# set execution state tracking to optimized production environment flags
ENV NODE_ENV=production

# create a secure, non-root system user context for process insulation
USER node

# copy production runtime dependency modules from the builder stage
COPY --chown=node:node --from=builder /app/node_modules/amqplib ./node_modules/amqplib
COPY --chown=node:node --from=builder /app/node_modules/dotenv ./node_modules/dotenv

# copy compiled distribution artifacts, assets, and base layouts
COPY --chown=node:node --from=builder /app/dist ./dist
COPY --chown=node:node --from=builder /app/public ./public
COPY --chown=node:node --from=builder /app/package.json ./package.json

# expose the application network gateway socket port mapping target
EXPOSE 3000

# execute the compiled application entry point
CMD ["node", "dist/app.js"]
