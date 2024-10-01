#!/bin/sh

set -e

# echo "run db migration"

# # Load environment variables from app.env
# set -o allexport
# source app.env
# set +o allexport

# # Debugging: List files and print the contents of app.env
# ls
# cat app.env
# echo "$DB_SOURCE"

# # Run database migration
# /app/migrate -path db/migration -database "$DB_SOURCE" -verbose up

echo "start the app"

# Execute the main application
exec "$@"