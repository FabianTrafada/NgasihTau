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
#
# Requirements: 10.3 (Clean Architecture - database per service)

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

log_info "Database initialization completed successfully!"
log_info "Created databases and users:"
for database in "${!SERVICE_CONFIGS[@]}"; do
    IFS=':' read -r username password <<< "${SERVICE_CONFIGS[$database]}"
    echo "  - Database: $database | User: $username"
done
