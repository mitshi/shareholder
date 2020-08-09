.PHONY: run
run:
	export $(cat .env | xargs) && go run *.go

.PHONY: generate
generate:
	sqlc generate

.PHONY: newmigration
newmigration:
	gomigrate create -ext sql -dir internal/db/migrations -seq $(name)

.PHONY: migrate
migrate:
	POSTGRESQL_URL='postgres://ox@localhost:5432/physicalshare?sslmode=disable'
	gomigrate -database ${POSTGRESQL_URL} -path internal/db/migrations up

.PHONY: sync
sync:
	rsync -a . ark:/srv/ox/shareholder
