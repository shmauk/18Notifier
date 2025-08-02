#!/bin/sh

echo "Waiting for DGraph to be ready..."

# Wait for DGraph to be healthy
while ! curl -f http://${DGRAPH_ENDPOINT:-http://dgraph:8080}/health > /dev/null 2>&1; do
  echo "DGraph not ready yet, waiting..."
  sleep 5
done

echo "DGraph is ready! Initializing schema..."

# Send the schema to DGraph
if curl -X POST ${DGRAPH_ENDPOINT:-http://dgraph:8080}/admin/schema \
  -H "Content-Type: application/graphql" \
  --data-binary @/app/schema.sdl; then
  echo "Schema initialized successfully!"
else
  echo "Failed to initialize schema"
  exit 1
fi 