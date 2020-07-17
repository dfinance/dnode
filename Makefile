include Makefile.ledger

git_tag=$(shell git describe --tags $(git rev-list --tags --max-count=1))
git_commit=$(shell git rev-list -1 HEAD)
tags = -X github.com/cosmos/cosmos-sdk/version.Name=dfinance \
	   -X github.com/cosmos/cosmos-sdk/version.ServerName=dnode \
	   -X github.com/cosmos/cosmos-sdk/version.ClientName=dncli \
	   -X github.com/cosmos/cosmos-sdk/version.Commit=$(git_commit) \
	   -X github.com/cosmos/cosmos-sdk/version.Version=${git_tag} \

build_dir=./.build
swagger_dir=$(build_dir)/swagger
cosmos_dir=$(swagger_dir)/cosmos-sdk
dnode = ./cmd/dnode
dncli =./cmd/dncli

cosmos_version=v0.38.4

all: install
install: go.sum install-dnode install-dncli
swagger-ui: swagger-ui-deps swagger-ui-build
tests: | test-unit test-cli test-rest test-integ

install-dnode:
	GO111MODULE=on go install -ldflags "$(tags)"  -tags "$(build_tags)" $(dnode)
install-dncli:
	GO111MODULE=on go install -ldflags "$(tags)"  -tags "$(build_tags)" $(dncli)

lint:
	@echo "--> Running Golang linter (unused variable / function warning are skipped)"
	golangci-lint run --exclude 'unused'

test-unit:
	@echo "--> Testing: UNIT tests"
	go test ./... -tags=unit -count=1
test-cli: install
	@echo "--> Testing: dncli CLI tests"
	go test ./... -tags=cli -count=1
test-rest: install
	@echo "--> Testing: dncli REST endpoints tests"
	go test ./... -tags=rest -count=1
test-integ: install
	@echo "--> Testing: dnode <-> dvm integration tests (using Binary)"
	DN_DVM_INTEG_TESTS_USE=binary go test ./... -v -tags=integ -count=1

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify

swagger-ui-deps:
	@echo "--> Preparing deps fro building Swagger API specificaion"

	@echo "-> Make tmp build folder"
	rm -rf $(swagger_dir)
	mkdir -p $(cosmos_dir)

	@echo "-> Cosmos-SDK $(cosmos_version) checkout"
	git -C $(swagger_dir) clone --branch $(cosmos_version) https://github.com/cosmos/cosmos-sdk.git

	@echo "-> Fetching Golang libraries: swag, statik"
	go get -u github.com/swaggo/swag/cmd/swag
	go get github.com/g3co/go-swagger-merger

swagger-ui-build:
	@echo "--> Building Swagger API specificaion, merging it to Cosmos SDK"

	@echo "-> Build swagger.yaml (that takes time)"
	swag init --dir . --output $(swagger_dir) --generalInfo ./cmd/dnode/main.go --parseDependency
	#swag init --dir . --output $(swagger_dir) --generalInfo ./cmd/dnode/main.go

	@echo "-> Merging swagger files"
	go-swagger-merger -o ./cmd/dncli/docs/swagger.yaml $(swagger_dir)/swagger.yaml $(cosmos_dir)/client/lcd/swagger-ui/swagger.yaml

	@echo "-> Building swagger.go file"
	echo "package docs\n\nconst Swagger = \`" > ./cmd/dncli/docs/swagger.go
	cat ./cmd/dncli/docs/swagger.yaml | sed "s/\`/'/g" >> ./cmd/dncli/docs/swagger.go
	echo "\`" >> ./cmd/dncli/docs/swagger.go

## binaries builds (xgo required: https://github.com/karalabe/xgo)
binaries: go.sum
	mkdir -p ./builds
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-darwin-amd64 ${dncli}
	GOOS=linux GOARCH=386 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-linux-386 ${dncli}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-linux-amd64 ${dncli}
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-windows-amd64.exe ${dncli}
	GOOS=windows GOARCH=386 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-windows-386.exe ${dncli}
