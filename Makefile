postgres: 
	docker run --name postgres16 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root safe_as_houses

dropdb:
	docker exec -it postgres16 dropdb safe_as_houses

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/safe_as_houses?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/safe_as_houses?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/jtaylor-io/safe-as-houses/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock

