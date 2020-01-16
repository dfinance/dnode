include Makefile.ledger

PROTO_IN_DIR=./vm-proto/protos/
PROTO_OUT_DIR=./x/core/protos/
PROTOBUF_FILES=./vm-proto/protos/vm.proto

all: protos install
install: protos go.sum
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/wbd
		GO111MODULE=on go install -tags "$(build_tags)" ./cmd/wbcli
go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify
protos:
	mkdir -p  ${PROTO_OUT_DIR}
	protoc -I ${PROTO_IN_DIR} --go_out=plugins=grpc:$(PROTO_OUT_DIR) $(PROTOBUF_FILES)
