DB_URL=postgresql://postgres:1952001@localhost:5431/simple_bank?sslmode=disable

postgres:
	docker run --name postgres12 -p 5431:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=1952001 -d postgres

createdb:
	docker exec -it postgres12 createdb --username=postgres --owner=postgres simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down


migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock: 
	mockgen --package mockdb  -destination db/mock/store.go github.com/nhat195/simple_bank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock migratedown1 migrateup1