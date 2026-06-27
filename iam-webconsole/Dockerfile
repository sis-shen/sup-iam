# ============================================================
# Stage 1: Build Vue 3 application
# ============================================================
FROM node:20-alpine AS build

WORKDIR /app

# Copy package files and install dependencies
COPY package.json ./
RUN npm install

# Copy source code and build
COPY src/ ./src/
COPY public/ ./public/
COPY index.html vite.config.js ./
RUN npm run build

# ============================================================
# Stage 2: Runtime with Nginx
# ============================================================
FROM nginx:1.27-alpine

# Install gettext for envsubst (already included in nginx:alpine)
# RUN apk add --no-cache gettext

# Copy nginx template
COPY nginx/nginx.conf.template /etc/nginx/templates/nginx.conf.template

# Copy entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Copy built static files from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Expose port
EXPOSE 8080

# Use custom entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
