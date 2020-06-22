# VM

DN blockchain currently supports smart-contracts via Move VM.

Two types of Move transaction are supported: publish module / execute script.

To publish a module:

    dncli tx vm publish [fileMV] --from <from> --fees <fees>
    
To execute a script:

    dncli tx vm execute [fileMV] arg1:type1, arg2:type2, arg3:type3... --from <from> --fees <fees>
    
    # Or (as an example with arguments):
    dncli tx vm execute [fileMV] true:Bool, 150:U64 --from <from> --fees <fees>
    
To get execution results (gas spent, events) just query the transaction:

    dncli query tx [transactionId]

To get detailed explanation about VM error (if the transaction contains the error), just query the transaction with `vm` module:

    dncli query vm tx [transactionId]

Output will contain all events collected from script execution / module deploy.

Events have status:
* successful execution (status `keep`):

    ```json
    {
      "type": "keep"
    }
    ```
  
* execution/deploy failed (status `discard`):

    ```json
    {
      "type": "discard",
      "attributes": [
        {
          "key": "major_status",
          "value": "0"
        },
        {
          "key": "sub_status",
          "value": "0"
        },
        {
          "key": "message",
          "value": "error message"
        }
      ]
    }
    ```
  
* error state (status `error`): event fields are similar to `keep` and `discard` statuses.

## Genesis compilation

First of all, to get DN work correctly, we need to compile standard DN smart module libs
and put the result into the genesis block.
Results are WriteSet operations, that write compiled modules into the storage.

1. Go to VM folder and run:

        cargo run --bin stdlib-builder lang/stdlib -po ./genesis-ws.json

2. Go to DN folder and run:

        dnode read-genesis-write-set [path to created file genesis-ws.json]

Everything should be fine now.

## Compilation

Launch the DVM server (compiler & runtime) and DN.

Then use commands to compile modules/scripts:

    dncli query vm compile-script [moveFile] [address] --to-file <script.move.json>
    dncli query vm compile-module [moveFile] [address] --to-file <module.move.json>  

Where:
 * `moveFile` - file that contains Move code;
 * `address` - address of account who will use the compiled code;
 * `--to-file` - allows to output the result to a file, otherwise it will be printed to console;
 * `--compiler` - address of the compiler server (optional, default is `tcp://127.0.0.1:50051`);

Refer to [DVM readme](https://github.com/dfinance/dvm/blob/master/README.md) on how to install and start the compiler
server and the VM runtime server.

## Configuration

Default VM configuration file can be found at `~/.dnode/config/vm.toml`.
Configuration file is initialized to defaults on `init` command.

```toml
# This is a TOML config file to configurate connection to VM.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# VM network address to connect.
vm_address = "tcp://127.0.0.1:50051"

# VM data server listen address.
vm_data_listen = "tpc://127.0.0.1:50052

# VM retry settings.

## Retry max attempts.
## Default is 0 - infinity attempts, -1 - to desable.
vm_retry_max_attempts = 0

## Initial backoff in ms.
## Default is 100ms.
vm_retry_initial_backoff = 100

## Max backoff in ms.
## Default is 150ms.
vm_retry_max_backoff = 150

## Backoff multiplier.
## Default 
vm_retry_backoff_multiplier = 0.1
```

Where:

* `vm_address` - address of Move VM runtime server (used to deploy/execute modules);
* `vm_data_listen` - address of the Data Source listen server (part of DN) which is used to share data between DN and VM;

The rest are timeout and retry mechanism parameters, we don't recommend to change them.

Supported protocol schemes for DN <-> VM communication are:
* `tcp` - using gRPC over network (example: `tcp://127.0.0.1:50051`);
* `unix` - using gRPC over Unix sockets (example: `unix:///socket_file.sock` for file at `/socket_file.sock` path);

Protocol notes:
*  refer to [DVM readme](https://github.com/dfinance/dvm/blob/master/README.md) to find a corresponding protocol scheme
used to configure a VM server (as a reference: DN `unix:///file` -> VM `ipc://file`,
DN `tcp://127.0.0.1:50051` -> `http://127.0.0.1:50051`);
* compiler address (used by `dncli` application) also supports `tcp ` and `unix` schemes and
its value can be found at `~/dncli/config` file, the `compiler` field;

## Get storage data

It possible to read storage data by path, e.g.:

    dncli query vm get-data [address] [path]

Where:
 * `address` - address of account containing data, could be bech32 or hex string (libra);
 * `path` - resource path, hex string;
