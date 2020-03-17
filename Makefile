include Makefile.ledger

git_tag=$(shell git describe --tags $(git rev-list --tags --max-count=1))
git_commit=$(shell git rev-list -1 HEAD)
tags = -X github.com/cosmos/cosmos-sdk/version.Name=dfinance \
	   -X github.com/cosmos/cosmos-sdk/version.ServerName=dnode \
	   -X github.com/cosmos/cosmos-sdk/version.ClientName=dncli \
	   -X github.com/cosmos/cosmos-sdk/version.Commit=$(git_commit) \
	   -X github.com/cosmos/cosmos-sdk/version.Version=${git_tag} \

all: install
install: go.sum install-dnode install-dncli install-oracleapp

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
	@echo "  >  Checking if there is any missing dependencies..."
	go get -u github.com/golang/protobuf/protoc-gen-go
