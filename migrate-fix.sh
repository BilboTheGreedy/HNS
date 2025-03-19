#!/bin/bash
# Script to fix a dirty database migration

# Check if a migration tool is specified
if [ -z "$1" ]; then
  echo "Usage: $0 <migration-tool> [database-url]"
  echo "Example: $0 golang-migrate postgresql://postgres:postgres@localhost:5432/hns"
  echo "Example: $0 flyway jdbc:postgresql://localhost:5432/dbname"
  exit 1
fi

MIGRATION_TOOL=$1
DB_URL=$2

case $MIGRATION_TOOL in
  "golang-migrate"|"migrate")
    echo "Using golang-migrate to fix dirty database..."
    if [ -z "$DB_URL" ]; then
      echo "Error: Database URL is required for golang-migrate"
      exit 1
    fi
    
    # Check current migration status
    echo "Current migration status:"
    migrate -database "$DB_URL" -path ./migrations version
    
    # Force the version to 2 (the version mentioned in your error)
    echo "Forcing database version to 2..."
    migrate -database "$DB_URL" -path ./migrations force 2
    
    # Verify the fix
    echo "New migration status:"
    migrate -database "$DB_URL" -path ./migrations version
    ;;
    
  "flyway")
    echo "Using Flyway to fix dirty database..."
    if [ -z "$DB_URL" ]; then
      echo "Error: Database URL is required for Flyway"
      exit 1
    fi
    
    # Check current migration status
    echo "Current migration status:"
    flyway -url="$DB_URL" info
    
    # Repair the database to fix the dirty flag
    echo "Repairing database..."
    flyway -url="$DB_URL" repair
    
    # Verify the fix
    echo "New migration status:"
    flyway -url="$DB_URL" info
    ;;
    
  "django")
    echo "Using Django to fix dirty database..."
    
    # Show migrations status
    echo "Current migration status:"
    python manage.py showmigrations
    
    # Fake the migration to mark it as completed
    echo "Marking migration as completed..."
    python manage.py migrate --fake
    
    # Verify the fix
    echo "New migration status:"
    python manage.py showmigrations
    ;;
    
  "knex")
    echo "Using Knex.js to fix dirty database..."
    
    # Check current migration status
    echo "Current migration status:"
    npx knex migrate:status
    
    # Force the specified migration version
    echo "Forcing migration version..."
    npx knex migrate:up 002_migration_name.js --force
    
    # Verify the fix
    echo "New migration status:"
    npx knex migrate:status
    ;;
    
  *)
    echo "Unknown migration tool: $MIGRATION_TOOL"
    echo "Supported tools: golang-migrate, flyway, django, knex"
    exit 1
    ;;
esac

echo "Migration fix completed. Please check the output to ensure it was successful."
echo "You can now run your regular migration command to continue from this point."