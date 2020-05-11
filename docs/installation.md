# Installation

Before we start you should have a correct 'GOPATH', 'GOROOT' environment variables.

Required:

    * golang 1.13.8 or later.

## Install as binary

To install both cli and daemon as binaries you can use Makefile:

    make install

So after this command both `dnode` and `dncli` will be available from console

    dnode version --long
    dncli version --long

If you want to install specific application (not everything), you always can do:

    make install-dnode
    make install-dncli
    make install-oracleapp

## Build from go

And let's build both daemon and cli:

    GO111MODULE=on go build -o dnode ./cmd/dnode 
    GO111MODULE=on go build -o dncli ./cmd/dncli

And then you can try:

    ./dnode --help
    ./dncli --help

## Run from go

Both commands must execute fine, after it you can run both daemon and cli:

    GO111MODULE=on go run ./cmd/dnode --help
    GO111MODULE=on go run ./cmd/dncli --help
 