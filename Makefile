include Makefile.ledger

git_tag=$(shell git describe --tags $(git rev-list --tags --max-count=1))
git_commit=$(shell git rev-list -1 HEAD)
tags = -X github.com/cosmos/cosmos-sdk/version.Name=dfinance \
	   -X github.com/cosmos/cosmos-sdk/version.ServerName=dnode \
	   -X github.com/cosmos/cosmos-sdk/version.ClientName=dncli \
	   -X github.com/cosmos/cosmos-sdk/version.Commit=$(git_commit) \
	   -X github.com/cosmos/cosmos-sdk/version.Version=${git_tag} \

dncli =./cmd/dncli

all: install
install: go.sum install-dnode install-dncli install-oracleapp

install-dnode:
		GO111MODULE=on go install --ldflags "$(tags)"  -tags "$(build_tags)" ./cmd/dnode
install-dncli:
		GO111MODULE=on go install  --ldflags "$(tags)"  -tags "$(build_tags)" ${dncli}
install-oracleapp:
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/oracle-app

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

## binaries builds (xgo required: https://github.com/karalabe/xgo)
binaries: go.sum
	mkdir -p ./builds
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-darwin-amd64 ${dncli}
	GOOS=linux GOARCH=386 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-linux-386 ${dncli}
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-linux-amd64 ${dncli}
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-windows-amd64.exe ${dncli}
	GOOS=windows GOARCH=386 CGO_ENABLED=0 GO111MODULE=on go build --ldflags "$(tags)"  -tags "$(build_tags)" -o ./builds/dncli-${git_tag}-windows-386.exe ${dncli}