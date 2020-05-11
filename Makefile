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
cosmos_version=v0.38.2
dncli =./cmd/dncli

all: install
install: go.sum install-dnode install-dncli install-oracleapp
swagger-ui: swagger-ui-deps swagger-ui-build

install-dnode:
		GO111MODULE=on go install --ldflags "$(tags)"  -tags "$(build_tags)" ./cmd/dnode
install-dncli:
		GO111MODULE=on go install  --ldflags "$(tags)"  -tags "$(build_tags)" ${dncli}
install-oracleapp:
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/oracle-app

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
	go get github.com/rakyll/statik

	@echo "-> Modify SDK's swagger-ui"
	mv $(cosmos_dir)/client/lcd/swagger-ui/swagger.yaml $(cosmos_dir)/client/lcd/swagger-ui/sdk-swagger.yaml
	sed -i.bak -e 's/url:.*/urls: [{url: \".\/dn-swagger.yaml\", name: \"Dfinance API\"},{url: \"\.\/sdk-swagger\.yaml\", name: \"Cosmos SDK API\"}],/' $(cosmos_dir)/client/lcd/swagger-ui/index.html

swagger-ui-build:
	@echo "--> Building Swagger API specificaion, merging it to Cosmos SDK"

	@echo "-> Build swagger.yaml (that takes time)"
	#swag init --dir . --output $(swagger_dir) --generalInfo ./cmd/dnode/main.go --parseDependency
	swag init --dir . --output $(swagger_dir) --generalInfo ./cmd/dnode/main.go
	cp $(swagger_dir)/swagger.yaml $(cosmos_dir)/client/lcd/swagger-ui/dn-swagger.yaml

	@echo "-> Build statik FS"
	rm -rf ./cmd/dncli/docs/statik
	statik -src=$(cosmos_dir)/client/lcd/swagger-ui -dest=./cmd/dncli/docs

## binaries builds (xgo required: https://github.com/karalabe/xgo)
binaries: go.sum
	mkdir -p ./builds
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-darwin-amd64 ${dncli}
	GOOS=linux GOARCH=386 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-linux-386 ${dncli}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-linux-amd64 ${dncli}
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-windows-amd64.exe ${dncli}
	GOOS=windows GOARCH=386 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-windows-386.exe ${dncli}
