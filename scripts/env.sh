#!/bin/bash

export IMAGE_NAME=${IMAGE_NAME:-appts}
export IMAGE_TAG=${IMAGE_TAG:-latest}

export HOST=${HOST:-0.0.0.0}
export PORT=${PORT:-8080}
export DB_HOST=${DB_HOST:-db}
export DB_PORT=${DB_PORT:-5432}
export DB_NAME=${DB_NAME:-appts}
export DB_USER=${DB_USER:-postgres}
export DB_PASS=${DB_PASS:-password}

export POSTGRES_DATA_DIR=${POSTGRES_DATA_DIR:-~/scr/postgres/data}

exec "$@"
