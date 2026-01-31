#!/bin/bash
# ===========================================
# PostgreSQL Multiple Database Initialization
# ===========================================
# This script creates multiple databases and dedicated users for each microservice
# It runs automatically when the PostgreSQL container starts for the first time
#
# Each service gets:
# - Its own database
# - A dedicated user with limited permissions
# - Full access only to its own database

set -e
set -u

# ===========================================
# Configuration
# ===========================================
# Service database and user mappings
# Format: database_name:username:password
declare -A SERVICE_CONFIGS=(
    ["ngasihtau_users"]="user_svc:user_svc_password"
    ["ngasihtau_pods"]="pod_svc:pod_svc_password"
    ["ngasihtau_materials"]="material_svc:material_svc_password"
    ["ngasihtau_ai"]="ai_svc:ai_svc_password"
    ["ngasihtau_notifications"]="notification_svc:notification_svc_password"
    ["ngasihtau_search"]="search_svc:search_svc_password"
)

# Migration folder mappings (database -> migration folder name)
# Note: Multiple folders can target the same database (e.g., material + offline)
declare -A MIGRATION_FOLDERS=(
    ["ngasihtau_users"]="user"
    ["ngasihtau_pods"]="pod"
    ["ngasihtau_materials"]="material"
    ["ngasihtau_ai"]="ai"
    ["ngasihtau_notifications"]="notification"
)

# Additional migration folders for the same database
# Format: database -> space-separated list of additional folders
declare -A ADDITIONAL_MIGRATIONS=(
    ["ngasihtau_materials"]="offline"
)

# Base path for migrations (relative to where script runs or absolute)
MIGRATIONS_BASE_PATH="${MIGRATIONS_BASE_PATH:-/migrations}"

# ===========================================
# Functions
# ===========================================

function log_info() {
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

function log_error() {
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $1" >&2
}

function create_user() {
    local username=$1
    local password=$2
    
    log_info "Creating user: $username"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
        DO \$\$
        BEGIN
            IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$username') THEN
                CREATE USER $username WITH PASSWORD '$password';
                RAISE NOTICE 'User $username created successfully';
            ELSE
                RAISE NOTICE 'User $username already exists, skipping creation';
            END IF;
        END
        \$\$;
EOSQL
}

function create_database() {
    local database=$1
    local owner=$2
    
    log_info "Creating database: $database with owner: $owner"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
        SELECT 'CREATE DATABASE $database OWNER $owner'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$database')\gexec
EOSQL
}

function grant_permissions() {
    local database=$1
    local username=$2
    
    log_info "Granting permissions on $database to $username"
    
    # Grant connection privilege
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
        GRANT CONNECT ON DATABASE $database TO $username;
EOSQL
    
    # Grant schema and table permissions within the database
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$database" <<-EOSQL
        -- Grant usage on public schema
        GRANT USAGE ON SCHEMA public TO $username;
        
        -- Grant all privileges on all tables (current and future)
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $username;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO $username;
        
        -- Grant all privileges on all sequences (current and future)
        GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $username;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO $username;
        
        -- Grant execute on all functions (current and future)
        GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO $username;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO $username;
        
        -- Allow creating objects in public schema
        GRANT CREATE ON SCHEMA public TO $username;
EOSQL
}

function revoke_public_access() {
    local database=$1
    
    log_info "Revoking public access on $database"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
        -- Revoke default public access to the database
        REVOKE ALL ON DATABASE $database FROM PUBLIC;
EOSQL
}

function setup_extensions() {
    local database=$1
    
    log_info "Setting up extensions for $database"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$database" <<-EOSQL
        -- Enable UUID generation extension
        CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
        
        -- Enable pgcrypto for encryption functions
        CREATE EXTENSION IF NOT EXISTS "pgcrypto";
EOSQL
}

function run_migrations() {
    local database=$1
    local migration_folder=$2
    
    local migration_path="${MIGRATIONS_BASE_PATH}/${migration_folder}"
    
    # Check if migration folder exists
    if [ ! -d "$migration_path" ]; then
        log_info "No migrations found for $database at $migration_path, skipping..."
        return 0
    fi
    
    # Check if there are any migration files
    if [ -z "$(ls -A "$migration_path"/*.sql 2>/dev/null)" ]; then
        log_info "No SQL migration files in $migration_path, skipping..."
        return 0
    fi
    
    log_info "Running migrations for $database from $migration_path"
    
    # Build database URL
    local db_url="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:5432/${database}?sslmode=disable"
    
    # Check if migrate tool is available
    if command -v migrate &> /dev/null; then
        migrate -path "$migration_path" -database "$db_url" up
        if [ $? -eq 0 ]; then
            log_info "Migrations completed successfully for $database"
        else
            log_error "Migration failed for $database"
            return 1
        fi
    else
        log_info "migrate tool not found, running SQL files directly..."
        # Fallback: run .up.sql files in order
        for sql_file in $(ls "$migration_path"/*.up.sql 2>/dev/null | sort); do
            log_info "Applying migration: $(basename "$sql_file")"
            psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$database" -f "$sql_file"
            if [ $? -ne 0 ]; then
                log_error "Failed to apply migration: $sql_file"
                return 1
            fi
        done
        log_info "All SQL migrations applied for $database"
    fi
}

# ===========================================
# Main Execution
# ===========================================

log_info "Starting database initialization..."

# Process each service configuration
for database in "${!SERVICE_CONFIGS[@]}"; do
    IFS=':' read -r username password <<< "${SERVICE_CONFIGS[$database]}"
    
    log_info "Processing service: $database"
    
    # Create the service user
    create_user "$username" "$password"
    
    # Create the database with the service user as owner
    create_database "$database" "$username"
    
    # Revoke public access for security
    revoke_public_access "$database"
    
    # Grant appropriate permissions
    grant_permissions "$database" "$username"
    
    # Setup common extensions
    setup_extensions "$database"
    
    log_info "Completed setup for: $database"
    echo "---"
done

# Also grant superuser access to all databases (for migrations and admin tasks)
log_info "Granting superuser access to all service databases..."
for database in "${!SERVICE_CONFIGS[@]}"; do
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
        GRANT ALL PRIVILEGES ON DATABASE $database TO $POSTGRES_USER;
EOSQL
done

# ===========================================
# Run Migrations
# ===========================================
log_info "Running database migrations..."
for database in "${!MIGRATION_FOLDERS[@]}"; do
    migration_folder="${MIGRATION_FOLDERS[$database]}"
    run_migrations "$database" "$migration_folder"
done

# Run additional migrations (e.g., offline module on materials database)
log_info "Running additional migrations..."
for database in "${!ADDITIONAL_MIGRATIONS[@]}"; do
    for migration_folder in ${ADDITIONAL_MIGRATIONS[$database]}; do
        run_migrations "$database" "$migration_folder"
    done
done

log_info "Database initialization completed successfully!"
log_info "Created databases and users:"
for database in "${!SERVICE_CONFIGS[@]}"; do
    IFS=':' read -r username password <<< "${SERVICE_CONFIGS[$database]}"
    echo "  - Database: $database | User: $username"
done
