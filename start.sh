#!/bin/sh

set -e
 
echo "run db migration"
source app.env
ls 
cat app.env
echo "$DB_SOURCE"
/app/migrate -path db/migration -database "$DB_SOURCE" -verbose up
echo "start the app"
exec "$@"
