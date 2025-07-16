#!/bin/sh
set -e

KAFKA_BROKER_URL="kafka:29092"
REDIS_URL="redis2:6379"

KAFKA_HOST=$(echo $KAFKA_BROKER_URL | cut -d: -f1)
KAFKA_PORT=$(echo $KAFKA_BROKER_URL | cut -d: -f2)

REDIS_HOST=$(echo $REDIS_URL | cut -d: -f1)
REDIS_PORT=$(echo $REDIS_URL | cut -d: -f2)

echo "⏳ Waiting for Kafka at $KAFKA_BROKER_URL..."
until nc -z $KAFKA_HOST $KAFKA_PORT; do
  >&2 echo "Kafka is unavailable - sleeping for 2 seconds..."
  sleep 2
done
echo "✅ Kafka is ready."

echo "⏳ Waiting for Redis at $REDIS_URL..."
until nc -z $REDIS_HOST $REDIS_PORT; do
  >&2 echo "Redis is unavailable - sleeping for 1 second..."
  sleep 1
done
echo "✅ Redis is ready."

echo "🚀 Starting application..."
exec "$@"
