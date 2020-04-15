# Tests

During standard launch of tests:

    GO111MODULE=on go test ./... --tags=unit

VM will use default configuration for integration tests (with connection to VM),
and with unit tests (using Mock servers), standard configuration looks so:

```go
// Mocks
DefaultMockVMAddress        = "127.0.0.1:60051" // Default virtual machine address to connect from Cosmos SDK.
DefaultMockDataListen       = "127.0.0.1:60052" // Default data server address to listen for connections from VM.
DefaultMockVMTimeoutDeploy  = 100               // Default timeout for deploy module request.
DefaultMockVMTimeoutExecute = 100               // Default timeout for execute request.

// Integrations
DefaultVMAddress        = "127.0.0.1:50051" // Default virtual machine address to connect from Cosmos SDK.
DefaultDataListen       = "127.0.0.1:50052" // Default data server address to listen for connections from VM.
DefaultVMTimeoutDeploy  = 100               // Default timeout for deploy module request.
DefaultVMTimeoutExecute = 100               // Default timeout for execute request.
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
    
To launch VM integration tests (required docker installed, dvm registry authorized and image pulled) run:

    export TAG=master # needed tag (master by default)
    export REGISTRY=  # replace of registry contains dvm
    
    docker image pull ${REGISTRY}/dfinance/dvm:${TAG}
    
    GO111MODULE=on go test ./... --tags=integ
    
To launch REST API tests run:

    GO111MODULE=on go test ./... --tags=rest
    
To launch CLI tests (`dnode`, `dncli` binaries should be build and available within `$PATH`) run:

    GO111MODULE=on go test ./... --tags=cli
