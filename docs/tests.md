# Tests

During standard launch of tests:

    GO111MODULE=on go test ./... --tags=unit

VM will use default configuration for integration tests (with connection to VM),
and with unit tests (using Mock servers), standard configuration looks so:

```go
// Mocks
DefaultMockVMAddress        = "127.0.0.1:60051" // Default virtual machine address to connect from Cosmos SDK.
DefaultMockDataListen       = "127.0.0.1:60052" // Default data server address to listen for connections from VM.

// Integrations
DefaultVMAddress  = "127.0.0.1:50051" // Default virtual machine address to connect from Cosmos SDK.
DefaultDataListen = "127.0.0.1:50052" // Default data server address to listen for connections from VM.
```

To change these parameters during test launch, use next flags after test command:

* `--vm.mock.address` - Address of mock VM node, change only in case of conflicts with ports.
* `--ds.mock.listen` - Address to listen for data source server, change only in case of conflicts with ports.
* `--vm.address` - Address of VM node to connect during tests.
* `--ds.listen` - Address to listen for Data Source server during tests.

To launch tests **ONLY** related to VM:

     GO111MODULE=on go test dnode/x/vm/internal/keeper --tags=integ

## Integration tests

To launch tests covering basic logic run: 

    GO111MODULE=on go test ./... --tags=unit

There are two options to run integration tests (dnode <-> DVM integration):
1. Using Docker container.

    Requirements:
    * Docker installed;
    * DVM registry authorized;
    * DVM image pulled (`docker image pull ${REGISTRY}/dfinance/dvm:${TAG}`)

    Configuration:
    * `export DN_DVM_INTEG_TESTS_USE=docker` - using Docker for integration tests;
    * `export DN_DVM_INTEG_TESTS_DOCKER_REGISTRY=<docker_registry_path>` - Docker registry containing DVM image;
    * `export DN_DVM_INTEG_TESTS_DOCKER_TAG=master` - DVM Docker image tag;

2. Using prebuild binaries.

    Configuration:
    * `export DN_DVM_INTEG_TESTS_USE=binary` - using binary for integration tests;
    * `export DN_DVM_INTEG_TESTS_BINARY_PATH="/dvmDir"` - directory containing DVM binary (if not specified, file should be reachable within `$PATH`);

To launch VM integration tests run:

    GO111MODULE=on go test ./... --tags=integ
    
To launch REST API tests run:

    GO111MODULE=on go test ./... --tags=rest
    
To launch CLI tests (`dnode`, `dncli` binaries should be build and available within `$PATH`) run:

    GO111MODULE=on go test ./... --tags=cli
