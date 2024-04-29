#! /usr/bin/env bash
set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo 'Starting database...'

docker run \
  --name ucp-postgres \
  -e POSTGRES_PASSWORD=ucprulz \
  -p 5432:5432 \
  -v "$SCRIPT_DIR/init/init-db.sh:/docker-entrypoint-initdb.d/init-db.sh" \
  --rm \
  -d \
  postgres

echo "Database started successfully."
echo "Connect using: 'psql -h localhost -U ucp -P ucprulz -d ucp_state'
echo "Connection string: 'host=localhost user=postgres password=example port=5432 connect_timeout=10 database=dapr_test'