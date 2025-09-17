#!/bin/bash

# Script to run Linier Channel with Docker using local MySQL and Redis

echo "🚀 Starting Linier Channel with Docker (Local MySQL & Redis)..."

# Check if MySQL is running locally
echo "📋 Checking local MySQL connection..."
if ! nc -z localhost 3306; then
    echo "❌ MySQL is not running on localhost:3306"
    echo "Please start MySQL locally first"
    exit 1
fi

# Check if Redis is running locally
echo "📋 Checking local Redis connection..."
if ! nc -z localhost 6379; then
    echo "❌ Redis is not running on localhost:6379"
    echo "Please start Redis locally first"
    exit 1
fi

# Create necessary directories (using paths from config.local.env)
echo "📁 Creating necessary directories..."
mkdir -p "/Users/herihandoko/Documents/Transcode/input"
mkdir -p "/Users/herihandoko/Documents/Transcode/output"
mkdir -p "/Users/herihandoko/Documents/Transcode/archive"

# Set proper permissions
chmod 755 "/Users/herihandoko/Documents/Transcode/input"
chmod 755 "/Users/herihandoko/Documents/Transcode/output"
chmod 755 "/Users/herihandoko/Documents/Transcode/archive"

# Build and start containers
echo "🔨 Building and starting containers..."
docker-compose -f docker-compose.local.yml up --build -d

# Wait for app to be ready
echo "⏳ Waiting for application to be ready..."
sleep 10

# Check if app is running
echo "🔍 Checking application health..."
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Application is running successfully!"
    echo ""
    echo "🌐 Application URL: http://localhost:8080"
    echo "📊 Health Check: http://localhost:8080/health"
    echo "📁 API Endpoints: http://localhost:8080/api/v1"
    echo ""
    echo "📝 Logs: docker-compose -f docker-compose.local.yml logs -f app"
    echo "🛑 Stop: docker-compose -f docker-compose.local.yml down"
else
    echo "❌ Application failed to start. Check logs:"
    echo "docker-compose -f docker-compose.local.yml logs app"
    exit 1
fi
