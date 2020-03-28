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
cosmos_version=v0.37.4

all: install
install: go.sum install-dnode install-dncli install-oracleapp
swagger-ui: swagger-ui-deps swagger-ui-build

install-dnode:
		GO111MODULE=on go install --ldflags "$(tags)"  -tags "$(build_tags)" ./cmd/dnode
install-dncli:
		GO111MODULE=on go install  --ldflags "$(tags)"  -tags "$(build_tags)" ./cmd/dncli
install-oracleapp:
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/oracle-app

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

## deps: Install missing dependencies. Runs `go get` internally. e.
deps:
	@echo "-->  Checking if there is any missing dependencies..."
	go get -u github.com/golang/protobuf/protoc-gen-go

swagger-ui-deps:
	@echo "--> Preparing deps fro building Swagger API specificaion"

	@echo "-> Make tmp build folder"
	rm -rf $(swagger_dir)
	mkdir -p $(cosmos_dir)

	@echo "-> Cosmos-SDK $(cosmos_version) checkout"
	git -C $(swagger_dir) clone --branch $(cosmos_version) https://github.com/cosmos/cosmos-sdk.git
	cp $(cosmos_dir)/client/lcd/swagger-ui/swagger.yaml ./cmd/dncli/docs/swagger-ui/sdk-swagger.yaml

	@echo "-> Fetching Golang libraries: swag, statik"
	go get -u github.com/swaggo/swag/cmd/swag
	go get github.com/rakyll/statik

swagger-ui-build:
	@echo "--> Building Swagger API specificaion, merging it to Cosmos SDK"

	@echo "-> Build swagger.yaml (that takes time)"
	swag init --dir . --output $(swagger_dir) --generalInfo ./cmd/dnode/main.go --parseDependency
	cp $(swagger_dir)/swagger.yaml ./cmd/dncli/docs/swagger-ui/dn-swagger.yaml

	@echo "-> Build statik FS"
	rm -rf ./cmd/dncli/docs/statik
	statik -src=./cmd/dncli/docs/swagger-ui -dest=./cmd/dncli/docs
