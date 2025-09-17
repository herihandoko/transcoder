#!/bin/bash

# Script to run Linier Channel RC with Docker

echo "🚀 Starting Linier Channel RC Environment..."

# Check if external MySQL is running (optional)
echo "📋 Checking external MySQL connection (optional)..."
if nc -z localhost 3306; then
    echo "✅ External MySQL detected on localhost:3306"
    USE_EXTERNAL_MYSQL=true
else
    echo "ℹ️  No external MySQL found, will use containerized MySQL if needed"
    USE_EXTERNAL_MYSQL=false
fi

# Check if external Redis is running (optional)
echo "📋 Checking external Redis connection (optional)..."
if nc -z localhost 6379; then
    echo "✅ External Redis detected on localhost:6379"
    USE_EXTERNAL_REDIS=true
else
    echo "ℹ️  No external Redis found, will use containerized Redis if needed"
    USE_EXTERNAL_REDIS=false
fi

# Create necessary directories
echo "📁 Creating RC storage directories..."
mkdir -p "./rc-storage/input"
mkdir -p "./rc-storage/output"
mkdir -p "./rc-storage/archive"

# Set proper permissions
chmod 755 "./rc-storage/input"
chmod 755 "./rc-storage/output"
chmod 755 "./rc-storage/archive"

# Determine which services to start
COMPOSE_PROFILES=""
if [ "$USE_EXTERNAL_MYSQL" = false ]; then
    COMPOSE_PROFILES="$COMPOSE_PROFILES,with-mysql"
fi
if [ "$USE_EXTERNAL_REDIS" = false ]; then
    COMPOSE_PROFILES="$COMPOSE_PROFILES,with-redis"
fi

# Remove leading comma if exists
COMPOSE_PROFILES=$(echo $COMPOSE_PROFILES | sed 's/^,//')

# Build and start containers
echo "🔨 Building and starting RC containers..."
if [ -n "$COMPOSE_PROFILES" ]; then
    echo "📦 Using profiles: $COMPOSE_PROFILES"
    docker-compose -f docker-compose.rc.yml --profile $COMPOSE_PROFILES up --build -d
else
    echo "📦 Using external MySQL and Redis"
    docker-compose -f docker-compose.rc.yml up --build -d
fi

# Wait for app to be ready
echo "⏳ Waiting for application to be ready..."
sleep 15

# Check if app is running
echo "🔍 Checking application health..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ RC Application is running successfully!"
    echo ""
    echo "🌐 Application URL: http://localhost:8080"
    echo "📊 Health Check: http://localhost:8080/health"
    echo "📁 API Endpoints: http://localhost:8080/api/v1"
    echo "📁 HLS Files: http://localhost:80/transcoded/"
    echo ""
    echo "📝 Logs: docker-compose -f docker-compose.rc.yml logs -f app"
    echo "🛑 Stop: docker-compose -f docker-compose.rc.yml down"
    echo ""
    echo "📂 RC Storage:"
    echo "   - Input: ./rc-storage/input/"
    echo "   - Output: ./rc-storage/output/"
    echo "   - Archive: ./rc-storage/archive/"
else
    echo "❌ RC Application failed to start. Check logs:"
    echo "docker-compose -f docker-compose.rc.yml logs app"
    exit 1
fi
