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

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/jtaylor-io/safe-as-houses/db/sqlc Store

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=safe-as-houses.json \
	proto/*.proto

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7.2-alpine

redis_ping:
	docker exec -it redis redis-cli ping

.PHONY: postgres createdb dropdb migrateup migratedown smigrateup1 migratedown1 sqlc test server mock proto redis redis_ping

