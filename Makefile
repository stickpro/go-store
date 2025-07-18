.PHONY:
.SILENT:
.DEFAULT_GOAL := run

MIGRATIONS_DIR = ./sql/postgres/migrations/
SEEDS_DIR = ./sql/postgres/seeds/

VERSION ?= $(strip $(shell ./scripts/version.sh))
VERSION_NUMBER := $(strip $(shell ./scripts/version.sh number))
COMMIT_HASH := $(shell git rev-parse --short HEAD)

OUT_BIN ?= ./.bin/go-store
GO_LDFLAGS ?=
GO_OPT_BASE := -ldflags "-X main.version=$(VERSION) $(GO_LDFLAGS) -X main.commitHash=$(COMMIT_HASH)"

BUILD_ENV := CGO_ENABLED=0
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S), Linux)
	BUILD_ENV += GOOS=linux
endif
ifeq ($(UNAME_S), Darwin)
	BUILD_ENV += GOOS=darwin
endif

UNAME_P := $(shell uname -p)
ifeq ($(UNAME_P),x86_64)
	BUILD_ENV += GOARCH=amd64
endif
ifneq ($(filter arm%,$(UNAME_P)),)
	BUILD_ENV += GOARCH=arm64
endif

#build
build:
	go mod download && $(BUILD_ENV) && go build $(GO_OPT_BASE) -o $(OUT_BIN) ./cmd/app

run: build
	$(OUT_BIN) $(filter-out $@,$(MAKECMDGOALS))

#database

db-create-migration:
	migrate create -ext sql -dir "$(MIGRATIONS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

db-create-seed:
	migrate create -ext sql -dir "$(SEEDS_DIR)" $(filter-out $@,$(MAKECMDGOALS))

#liner
lint:
	golangci-lint run --show-stats

fmt:
	gofumpt -l -w .
#gewnerator
gen-envs:
	go run ./cmd/app config genenvs

gen-sql:
	cd sql && pgxgen -pgxgen-config=pgxgen-postgres.yaml -sqlc-config=sqlc-postgres.yaml crud
	cd sql && pgxgen -pgxgen-config=pgxgen-postgres.yaml -sqlc-config=sqlc-postgres.yaml sqlc generate

gen-swag:
	swag fmt
	swag init --parseDependency --parseInternal -g ./cmd/app/main.go

download-geodb:
	curl -L -o ./storage/geo/GeoLite2-City.mmdb https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb

download-geo-cities:
	curl -L -o ./storage/geo/city.csv https://github.com/hflabs/city/raw/refs/heads/master/city.csv


# Empty goals trap
%:
	@: