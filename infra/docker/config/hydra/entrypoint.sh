#!/bin/sh
set -eu  # ‘-e’ para salir en error, ‘-u’ para variables no definidas

export POSTGRES_PASSWORD=$(cat "$POSTGRES_PASSWORD_FILE")
export SECRETS_SYSTEM=$(cat "$SECRETS_SYSTEM_FILE")
export DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable&options=-c%20search_path%3Dhydra"

exec hydra serve all