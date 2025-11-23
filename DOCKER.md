# Docker Deployment Guide

This guide covers deploying the AI SRE Platform using Docker and Docker Compose.

## Overview

The platform uses multi-stage Docker builds for all services:

- **Incident Service**: Go backend with Alpine Linux base
- **Dashboard**: React frontend served by nginx
- **Demo App**: Node.js application

All services include health checks and are optimized for both development and production environments.

## Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- At least 4GB RAM available for Docker
- GitHub Personal Access Token with `workflow` scope

## Quick Start

### Development Environment

1. Copy environment template:
```bash
cp .env.example .env
```

2. Edit `.env` with your credentials (minimum required):
```bash
GITHUB_TOKEN=your_github_token_here
ENCRYPTION_KEY=your_32_byte_encryption_key_here_12
```

3. Start all services:
```bash
./scripts/dev.sh
```

This will:
- Start PostgreSQL and Redis
- Run database migrations
- Start Incident Service, Dashboard, and Demo App
- Seed the database with sample data
- Display service URLs

### Production Environment

1. Set up production environment variables:
```bash
cp .env.example .env
# Edit .env with production credentials
```

2. Deploy:
```bash
./scripts/prod.sh
```

Or manually:
```bash
docker-compose -f docker-compose.prod.yml up -d
docker-compose -f docker-compose.prod.yml exec incident-service ./migrate
```

## Architecture

### Multi-Stage Builds

All Dockerfiles use multi-stage builds for optimal image size and security:

1. **Development Stage**: Includes development tools, hot-reload support
2. **Builder Stage**: Compiles/builds the application
3. **Production Stage**: Minimal runtime image with only necessary files

### Health Checks

All services include Docker health checks:

- **Incident Service**: `GET /api/v1/health` (30s interval)
- **Dashboard**: `GET /health` via nginx (30s interval)
- **Demo App**: `GET /api/health` (30s interval)
- **PostgreSQL**: `pg_isready` (10s interval)
- **Redis**: `redis-cli ping` (10s interval)

## Service Configuration

### Incident Service

**Ports:**
- `8080`: HTTP API
- `9090`: Prometheus metrics

**Environment Variables:**
```bash
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_NAME=ai_sre
DATABASE_USER=postgres
DATABASE_PASSWORD=your_password
DATABASE_SSL_MODE=disable  # Use 'require' in production
REDIS_HOST=redis
REDIS_PORT=6379
GITHUB_TOKEN=your_token
GITHUB_API_URL=https://api.github.com
ENCRYPTION_KEY=your_32_byte_key
CONFIG_PATH=/app/config.yaml
```

**Volumes:**
- `./config.yaml:/app/config.yaml:ro` - Service mappings and configuration

### Dashboard

**Ports:**
- `3001`: HTTP (development)
- `80`: HTTP (production, mapped to host 3001)

**Environment Variables:**
```bash
VITE_API_URL=http://incident-service:8080  # Internal service URL
```

**Features:**
- Gzip compression enabled
- Static asset caching (1 year)
- SPA routing support
- Security headers configured

### Demo App

**Ports:**
- `3002`: HTTP

**Environment Variables:**
```bash
PORT=3002
SENTRY_DSN=your_sentry_dsn
```

## Docker Compose Files

### docker-compose.dev.yml

Development configuration with:
- Hot-reload for all services
- Source code mounted as volumes
- Development-friendly logging
- Separate container names (suffixed with `-dev`)

**Usage:**
```bash
docker-compose -f docker-compose.dev.yml up
```

### docker-compose.prod.yml

Production configuration with:
- Optimized production builds
- Restart policies (`unless-stopped`)
- Health checks enabled
- Separate volumes for data persistence
- Separate container names (suffixed with `-prod`)

**Usage:**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## Common Operations

### View Logs

Development:
```bash
docker-compose -f docker-compose.dev.yml logs -f
```

Production:
```bash
docker-compose -f docker-compose.prod.yml logs -f
```

Specific service:
```bash
docker-compose -f docker-compose.prod.yml logs -f incident-service
```

### Stop Services

Development:
```bash
docker-compose -f docker-compose.dev.yml down
```

Production:
```bash
docker-compose -f docker-compose.prod.yml down
```

### Rebuild Images

After code changes:
```bash
docker-compose -f docker-compose.dev.yml build
docker-compose -f docker-compose.dev.yml up -d
```

Force rebuild without cache:
```bash
docker-compose -f docker-compose.prod.yml build --no-cache
```

### Run Database Migrations

Development:
```bash
cd incident-service
DATABASE_HOST=localhost \
DATABASE_PORT=5432 \
DATABASE_NAME=ai_sre \
DATABASE_USER=postgres \
DATABASE_PASSWORD=postgres \
DATABASE_SSL_MODE=disable \
go run cmd/migrate/main.go
```

Production:
```bash
docker-compose -f docker-compose.prod.yml exec incident-service ./migrate
```

### Access Container Shell

```bash
docker-compose -f docker-compose.dev.yml exec incident-service sh
docker-compose -f docker-compose.dev.yml exec dashboard sh
docker-compose -f docker-compose.dev.yml exec demo-app sh
```

### Check Service Health

```bash
# Incident Service
curl http://localhost:8080/api/v1/health

# Dashboard
curl http://localhost:3001/health

# Demo App
curl http://localhost:3002/api/health

# Metrics
curl http://localhost:9090/metrics
```

## Troubleshooting

### Services Won't Start

1. Check if ports are already in use:
```bash
lsof -i :8080  # Incident Service
lsof -i :3001  # Dashboard
lsof -i :3002  # Demo App
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
```

2. Check Docker logs:
```bash
docker-compose -f docker-compose.dev.yml logs
```

3. Verify environment variables:
```bash
docker-compose -f docker-compose.dev.yml config
```

### Database Connection Issues

1. Ensure PostgreSQL is healthy:
```bash
docker-compose -f docker-compose.dev.yml ps postgres
```

2. Check PostgreSQL logs:
```bash
docker-compose -f docker-compose.dev.yml logs postgres
```

3. Test connection manually:
```bash
docker exec ai-sre-postgres-dev psql -U postgres -d ai_sre -c "SELECT 1"
```

### Health Check Failures

1. Check container status:
```bash
docker-compose -f docker-compose.dev.yml ps
```

2. Inspect health check logs:
```bash
docker inspect ai-sre-incident-service-dev | jq '.[0].State.Health'
```

3. Test health endpoint manually:
```bash
docker-compose -f docker-compose.dev.yml exec incident-service wget -O- http://localhost:8080/api/v1/health
```

### Build Failures

1. Clear Docker build cache:
```bash
docker builder prune -a
```

2. Remove all containers and volumes:
```bash
docker-compose -f docker-compose.dev.yml down -v
```

3. Rebuild from scratch:
```bash
docker-compose -f docker-compose.dev.yml build --no-cache
```

## Performance Optimization

### Production Builds

All production images are optimized:

- **Incident Service**: ~20MB (Alpine + Go binary)
- **Dashboard**: ~25MB (nginx + static files)
- **Demo App**: ~150MB (Node.js + dependencies)

### Resource Limits

Add resource limits in production:

```yaml
services:
  incident-service:
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### Volume Performance

For better I/O performance on macOS:

```yaml
volumes:
  - ./incident-service:/app:delegated
```

## Security Best Practices

1. **Never commit .env files** - Use `.env.example` as template
2. **Use secrets management** - Consider Docker secrets or external vaults
3. **Enable SSL in production** - Set `DATABASE_SSL_MODE=require`
4. **Rotate credentials regularly** - Update `ENCRYPTION_KEY` and tokens
5. **Scan images for vulnerabilities**:
```bash
docker scan ai-sre-platform-incident-service
```

## Kubernetes Deployment

For Kubernetes deployment, convert Docker Compose to Kubernetes manifests:

```bash
kompose convert -f docker-compose.prod.yml
```

Or use the provided Helm charts (coming soon).

## CI/CD Integration

### GitHub Actions

Example workflow for building and pushing images:

```yaml
- name: Build and push Docker images
  run: |
    docker-compose -f docker-compose.prod.yml build
    docker tag ai-sre-platform-incident-service ghcr.io/org/incident-service:${{ github.sha }}
    docker push ghcr.io/org/incident-service:${{ github.sha }}
```

### GitLab CI

```yaml
build:
  script:
    - docker-compose -f docker-compose.prod.yml build
    - docker-compose -f docker-compose.prod.yml push
```

## Monitoring

### Container Metrics

View resource usage:
```bash
docker stats
```

### Prometheus Integration

Metrics are exposed at:
- Incident Service: `http://localhost:9090/metrics`

Configure Prometheus to scrape:
```yaml
scrape_configs:
  - job_name: 'incident-service'
    static_configs:
      - targets: ['incident-service:9090']
```

## Backup and Restore

### Database Backup

```bash
docker exec ai-sre-postgres-prod pg_dump -U postgres ai_sre > backup.sql
```

### Database Restore

```bash
docker exec -i ai-sre-postgres-prod psql -U postgres ai_sre < backup.sql
```

### Volume Backup

```bash
docker run --rm -v ai-sre-platform-prod_postgres_data_prod:/data -v $(pwd):/backup alpine tar czf /backup/postgres-backup.tar.gz /data
```

## Support

For issues related to Docker deployment:
- Check logs: `docker-compose logs`
- Verify configuration: `docker-compose config`
- Review health checks: `docker ps`
- Consult [CONTRIBUTING.md](CONTRIBUTING.md) for development setup

## Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Multi-stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Health Checks](https://docs.docker.com/engine/reference/builder/#healthcheck)
