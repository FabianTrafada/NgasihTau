#!/bin/bash
# ===========================================
# PostgreSQL Initialization Script
# ===========================================
# This script runs automatically on first PostgreSQL startup
# Creates all databases and users for NgasihTau services

set -e

echo "[INFO] Starting database initialization..."

# Create databases and users using environment variables
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
    -- Create users with passwords from environment
    DO \$\$
    BEGIN
        -- User Service
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user_svc') THEN
            CREATE USER user_svc WITH PASSWORD '${USER_DB_PASSWORD}';
        END IF;
        
        -- Pod Service
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'pod_svc') THEN
            CREATE USER pod_svc WITH PASSWORD '${POD_DB_PASSWORD}';
        END IF;
        
        -- Material Service
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'material_svc') THEN
            CREATE USER material_svc WITH PASSWORD '${MATERIAL_DB_PASSWORD}';
        END IF;
        
        -- AI Service
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'ai_svc') THEN
            CREATE USER ai_svc WITH PASSWORD '${AI_DB_PASSWORD}';
        END IF;
        
        -- Notification Service
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'notification_svc') THEN
            CREATE USER notification_svc WITH PASSWORD '${NOTIFICATION_DB_PASSWORD}';
        END IF;
        
        -- Search Service
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'search_svc') THEN
            CREATE USER search_svc WITH PASSWORD '${SEARCH_DB_PASSWORD}';
        END IF;
    END
    \$\$;

    -- Create databases
    SELECT 'CREATE DATABASE ngasihtau_users OWNER user_svc' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ngasihtau_users')\gexec
    SELECT 'CREATE DATABASE ngasihtau_pods OWNER pod_svc' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ngasihtau_pods')\gexec
    SELECT 'CREATE DATABASE ngasihtau_materials OWNER material_svc' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ngasihtau_materials')\gexec
    SELECT 'CREATE DATABASE ngasihtau_ai OWNER ai_svc' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ngasihtau_ai')\gexec
    SELECT 'CREATE DATABASE ngasihtau_notifications OWNER notification_svc' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ngasihtau_notifications')\gexec
    SELECT 'CREATE DATABASE ngasihtau_search OWNER search_svc' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ngasihtau_search')\gexec
EOSQL

# Setup extensions and permissions for each database
for db in ngasihtau_users ngasihtau_pods ngasihtau_materials ngasihtau_ai ngasihtau_notifications ngasihtau_search; do
    echo "[INFO] Setting up extensions for $db"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$db" <<-EOSQL
        CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
        CREATE EXTENSION IF NOT EXISTS "pgcrypto";
EOSQL
done

# Grant permissions
declare -A DB_USERS=(
    ["ngasihtau_users"]="user_svc"
    ["ngasihtau_pods"]="pod_svc"
    ["ngasihtau_materials"]="material_svc"
    ["ngasihtau_ai"]="ai_svc"
    ["ngasihtau_notifications"]="notification_svc"
    ["ngasihtau_search"]="search_svc"
)

for db in "${!DB_USERS[@]}"; do
    user="${DB_USERS[$db]}"
    echo "[INFO] Granting permissions on $db to $user"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$db" <<-EOSQL
        GRANT USAGE ON SCHEMA public TO $user;
        GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO $user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO $user;
        GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO $user;
        ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO $user;
        GRANT CREATE ON SCHEMA public TO $user;
EOSQL
done

echo "[INFO] Database initialization completed!"
echo "[INFO] Created databases: ngasihtau_users, ngasihtau_pods, ngasihtau_materials, ngasihtau_ai, ngasihtau_notifications, ngasihtau_search"
