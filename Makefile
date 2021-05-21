.PHONY: swagger

export DATASTORE_EMULATOR_HOST=localhost:8081

appd:
	go run cmd/appd/appd.go

datastore:
	docker run -d --name gcloud-emulator-datastore -p 8081:8081 gcr.io/google.com/cloudsdktool/cloud-sdk:330.0.0-emulators gcloud beta emulators datastore start --no-store-on-disk --project=dev --host-port=0.0.0.0:8081

stop-datastore:
	docker stop gcloud-emulator-datastore

rm-datastore:
	docker rm gcloud-emulator-datastore

reset-datastore: stop-datastore rm-datastore datastore

postgres:
	docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres

stop-postgres:
	docker stop postgres

rm-postgres:
	docker rm postgres

reset-postgres: stop-postgres rm-postgres postgres

deploy:
	gcloud --project=api-ldej-nl app deploy --quiet

integration:
	go test -count=1 -tags=integration ./...

swagger:
	swag init -g ./cmd/appd/appd.go -o ./swagger

lint:
	golangci-lint run