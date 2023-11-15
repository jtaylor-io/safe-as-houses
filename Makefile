postgres: 
	docker run --name postgres15 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15-alpine

createdb:
	docker exec -it postgres15 createdb --username=root --owner=root safe_as_houses

dropdb:
	docker exec -it postgres15 dropdb safe_as_houses

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/safe_as_houses?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/safe_as_houses?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/safe_as_houses?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/safe_as_houses?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/jtaylor-io/safe-as-houses/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown smigrateup1 migratedown1 sqlc test server mock

