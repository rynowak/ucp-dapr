#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER ucp;
    CREATE DATABASE ucp_state;
    CREATE DATABASE ucp_metadata;
	GRANT ALL PRIVILEGES ON DATABASE ucp_state TO ucp;
    GRANT ALL PRIVILEGES ON DATABASE ucp_metadata TO ucp;
EOSQL