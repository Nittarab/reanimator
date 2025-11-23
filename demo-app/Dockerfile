# Development stage
FROM node:20-alpine AS development

WORKDIR /app

# Install wget for health checks
RUN apk --no-cache add wget

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY . .

# Expose port (default 3002, can be overridden by PORT env var)
EXPOSE 3002

# Development mode - will be overridden by docker-compose
CMD ["npm", "run", "dev"]

# Build stage
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (production only)
RUN npm ci --only=production

# Copy source code
COPY . .

# Production stage
FROM node:20-alpine AS production

WORKDIR /app

# Install wget for health checks
RUN apk --no-cache add wget

# Copy dependencies and source from builder
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/src ./src
COPY --from=builder /app/public ./public

# Expose port (default 3002, can be overridden by PORT env var)
EXPOSE 3002

# Health check
HEALTHCHECK --interval=30s --timeout=10s --retries=3 --start-period=10s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-3002}/api/health || exit 1

CMD ["node", "src/index.js"]
