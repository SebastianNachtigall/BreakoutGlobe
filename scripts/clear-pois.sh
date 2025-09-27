#!/bin/bash

# Script to clear all POIs from the database
# Make sure Docker Compose is running first

echo "🗑️  Clearing all POIs from the database..."

# Check if postgres container is running
if ! docker compose ps postgres | grep -q "Up"; then
    echo "❌ PostgreSQL container is not running. Please start it with:"
    echo "   docker compose up -d postgres"
    exit 1
fi

# Connect to the database and run the SQL commands
echo "🔍 Checking current POI count..."
CURRENT_COUNT=$(docker compose exec -T postgres psql -U postgres -d breakoutglobe -t -c "SELECT COUNT(*) FROM pois;" 2>/dev/null | tr -d ' ')

if [ $? -ne 0 ]; then
    echo "❌ Could not connect to database or 'pois' table doesn't exist."
    echo "Let's check what tables exist:"
    docker compose exec -T postgres psql -U postgres -d breakoutglobe -c "\dt"
    exit 1
fi

echo "📊 Current POI count: $CURRENT_COUNT"

if [ "$CURRENT_COUNT" -eq 0 ]; then
    echo "✅ No POIs to delete!"
    exit 0
fi

# Ask for confirmation
read -p "❓ Are you sure you want to delete all $CURRENT_COUNT POIs? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Operation cancelled."
    exit 1
fi

# Delete all POIs
echo "🗑️  Deleting all POIs..."
docker compose exec -T postgres psql -U postgres -d breakoutglobe -c "DELETE FROM pois;"

# Check the result
NEW_COUNT=$(docker compose exec -T postgres psql -U postgres -d breakoutglobe -t -c "SELECT COUNT(*) FROM pois;" | tr -d ' ')
echo "✅ Deleted $(($CURRENT_COUNT - $NEW_COUNT)) POIs. Remaining: $NEW_COUNT"

echo "🎉 Database cleanup complete!"