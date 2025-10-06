#!/bin/sh

# Wait for MinIO to be ready
echo "Waiting for MinIO to be ready..."
until curl -f http://localhost:9000/minio/health/live > /dev/null 2>&1; do
  echo "MinIO is unavailable - sleeping"
  sleep 2
done

echo "MinIO is up - configuring..."

# Configure mc (MinIO Client)
mc alias set myminio http://localhost:9000 minioadmin minioadmin123

# Create bucket if it doesn't exist
mc mb myminio/eralove-uploads --ignore-existing

# Set bucket policy to public read (optional, for public file access)
mc anonymous set download myminio/eralove-uploads

echo "MinIO configuration complete!"
echo "Bucket 'eralove-uploads' is ready"
