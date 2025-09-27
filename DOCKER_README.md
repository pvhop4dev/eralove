# EraLove Docker Setup Guide

This guide explains how to run the EraLove application using Docker and Docker Compose.

## Prerequisites

- Docker Desktop installed and running
- Docker Compose (included with Docker Desktop)
- Git (for cloning the repository)

## Quick Start

### Development Environment
```bash
# Start development environment with hot reload
make dev

# Or manually:
docker-compose -f docker-compose.dev.yml up -d
```

### Production Environment
```bash
# Start production environment
make prod

# Or manually:
docker-compose up -d
```

## Available Services

### Core Services
- **Frontend**: React + TypeScript + Vite application (Port 3000)
- **Backend**: Go + Fiber API server (Port 8080)
- **MongoDB**: Database (Port 27017)
- **Redis**: Cache (Port 6379)

### Optional Services
- **Nginx**: Reverse proxy for production (Port 80/443)

## Environment Variables

### Backend Configuration
The following environment variables are configured in docker-compose.yml:

```env
PORT=8080
ENVIRONMENT=production
MONGO_URI=mongodb://admin:password123@mongodb:27017/eralove?authSource=admin
DATABASE_NAME=eralove
REDIS_ADDR=redis:6379
REDIS_PASSWORD=password123
REDIS_DB=0
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRATION=24h
CORS_ORIGINS=http://localhost:3000,http://localhost:80,http://frontend
```

### Frontend Configuration
```env
REACT_APP_API_URL=http://localhost:8080/api/v1
```

## Available Make Commands

### Development Commands
```bash
make help              # Show all available commands
make frontend          # Start frontend development server
make backend           # Start backend development server
make install-deps      # Install all dependencies
make wire-gen          # Generate Wire dependency injection code
make swagger-gen       # Generate Swagger documentation
```

### Docker Commands - Production
```bash
make docker-up         # Start all services (Production)
make docker-down       # Stop all services
make docker-build      # Build all Docker images
make docker-rebuild    # Rebuild without cache
make docker-logs       # Show logs from all services
make docker-logs-backend   # Show backend logs only
make docker-logs-frontend  # Show frontend logs only
```

### Docker Commands - Development
```bash
make docker-dev-up     # Start development environment with hot reload
make docker-dev-down   # Stop development environment
make docker-dev-logs   # Show development logs
make docker-dev-rebuild # Rebuild development images
```

### Database Commands
```bash
make db-up             # Start only database services
make db-down           # Stop database services
make db-reset          # Reset database (WARNING: Deletes all data)
```

### Quick Commands
```bash
make dev               # Quick start development
make prod              # Quick start production
make stop              # Stop all environments
make restart           # Restart production
make restart-dev       # Restart development
make health            # Check service health
```

## Service URLs

### Development
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Swagger Documentation: http://localhost:8080/swagger/
- Health Check: http://localhost:8080/health
- MongoDB: localhost:27017
- Redis: localhost:6379

### Production
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- With Nginx (optional): http://localhost

## Docker Compose Files

### docker-compose.yml (Production)
- Optimized for production deployment
- Multi-stage builds for smaller images
- Health checks for all services
- Proper restart policies
- Volume persistence for data

### docker-compose.dev.yml (Development)
- Hot reload for backend using Air
- Development-friendly configuration
- Faster startup times
- Volume mounting for live code changes

## Volumes

### Persistent Data
- `mongodb_data`: MongoDB database files
- `redis_data`: Redis cache files
- `./backend/logs`: Application logs

### Development Volumes
- `./backend:/app`: Backend source code (hot reload)
- `./frontend:/app`: Frontend source code

## Networking

All services communicate through a custom Docker network:
- Production: `eralove-network`
- Development: `eralove-dev-network`

## Health Checks

All services include health checks:
- **Backend**: HTTP GET to `/health`
- **Frontend**: HTTP GET to `/health`
- **MongoDB**: MongoDB ping command
- **Redis**: Redis ping command

## Troubleshooting

### Common Issues

1. **Port conflicts**: Make sure ports 3000, 8080, 27017, and 6379 are not in use
2. **Permission issues**: Ensure Docker has proper permissions
3. **Memory issues**: Increase Docker memory allocation if needed

### Useful Commands

```bash
# Check running containers
docker ps

# Check logs for specific service
docker-compose logs -f [service-name]

# Restart specific service
docker-compose restart [service-name]

# Remove all containers and volumes
make clean-all

# Check service health
make health

# Access MongoDB shell
docker exec -it eralove-mongodb mongosh -u admin -p password123

# Access Redis CLI
docker exec -it eralove-redis redis-cli -a password123
```

### Reset Everything

If you need to completely reset the environment:

```bash
make clean-all
make prod  # or make dev
```

## Production Deployment

For production deployment:

1. Update environment variables in `docker-compose.yml`
2. Change default passwords
3. Configure proper SSL certificates
4. Use the nginx service with proper configuration
5. Set up proper backup strategies for volumes

```bash
# Start with nginx reverse proxy
docker-compose --profile production up -d
```

## Security Notes

⚠️ **Important**: Change default passwords before production deployment!

- MongoDB: `MONGO_INITDB_ROOT_PASSWORD`
- Redis: `REDIS_PASSWORD`
- JWT: `JWT_SECRET`

## API Documentation

Once the backend is running, you can access the Swagger API documentation at:
- http://localhost:8080/swagger/

This provides interactive documentation for all available API endpoints.
