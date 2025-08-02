# Deployment Guide

This guide explains how to deploy the 18xxNotifier service using Docker containers.

## Prerequisites

- Docker and Docker Compose installed
- Discord Bot Token and Application ID

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Discord Bot Configuration
DISCORD_TOKEN=your_discord_bot_token_here
DISCORD_APP_ID=your_discord_application_id_here

# DGraph Configuration
DGRAPH_ENDPOINT=http://dgraph:8080

# Application Configuration
PORT=3000
```

## Running the Application

1. **Start all services:**
   ```bash
   cd src
   docker-compose up -d
   ```

2. **View logs:**
   ```bash
   cd src
   docker-compose logs -f
   ```

3. **Stop all services:**
   ```bash
   cd src
   docker-compose down
   ```

## Services

### DGraph Database
- **Port:** 8080 (GraphQL), 8000 (HTTP), 9080 (gRPC)
- **Health Check:** Available at `http://localhost:8080/health`
- **Data Persistence:** Stored in Docker volume `dgraph_data`

### Schema Initialization
- **Purpose:** Waits for DGraph to be ready and initializes the schema
- **Dependencies:** DGraph service must be healthy
- **Schema File:** Uses `docs/schema.sdl`

### 18xxNotifier Application
- **Port:** 3000
- **Health Check:** Available at `http://localhost:3000/health`
- **Dependencies:** DGraph and Schema Init services

## Development

To rebuild the application after code changes:

```bash
cd src
docker-compose build app
docker-compose up -d
```

## Troubleshooting

1. **Check service status:**
   ```bash
   cd src
   docker-compose ps
   ```

2. **View specific service logs:**
   ```bash
   cd src
   docker-compose logs schema-init
   docker-compose logs app
   ```

3. **Reset everything:**
   ```bash
   cd src
   docker-compose down -v
   docker-compose up -d
   ``` 