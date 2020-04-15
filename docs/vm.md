# VM

DN blockchain currently supports smart-contracts via Move VM.

Both two types of Move transaction supported, like: deploy module/execute script.

To deploy module:

    dncli tx vm deploy-module [fileMV] --from <from> --fees <fees>
    
To execute script:

    dncli tx vm execute-script [fileMV] arg1:type1, arg2:type2, arg3:type3... --from <from> --fees <fees>
    
    # Or (as example with arguments):
    dncli tx vm execute-script [fileMV] true:Bool, 150:U64 --from <from> --fees <fees>
    
To get results of execution, gas spent, events, just query transaction:

    dncli query tx [transactionId]

Output will contains all events, collected from script execution/module deploy, also events have status, like for successful execution
(status keep):

```json
{
  "type": "keep"
}
```

And (status discard, when execution/deploy failed):

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

Also, events could contains event type **error** with similar fields, like discard, that could happen
together with even **keep**.

## Genesis compilation

First of all, to get DN correctly work, needs to compile standard DN smart modules libs,
and put result into genesis block. Result is WriteSet operations, that write compiled modules 
into storage.

So, first of all, go to VM folder, and run:

    cargo run --bin stdlib-builder stdlib/mvir mvir -po ../genesis-ws.json

After this, go into DN folder and run:

    dnode read-genesis-write-set [path to created file genesis-ws.json]

Now everything should be fine.

## Compilation

Launch compiler server, and DN.

Then use commands to compile modules/scripts:

    dncli query vm compile-script [mvirFile] [address] --to-file <script.mv> --compiler 127.0.0.1:50053
    dncli query vm compile-module [mvirFile] [address] --to-file <module.mv> --compiler 127.0.0.1:50053    

Where:
 * `mvirFile` - file contains MVir code.
 * `address` - address of account who will use compiled code.
 * `--to-file` - allows to output result to file, otherwise it will be printed in console.
 * `--compiler` - address of compiler, could be ignored, default is `127.0.0.1:50053`.

## Configuration

Default VM configuration file placed into `~/.dnode/config/vm.toml`, and will be 
initialized after `init` command.

As Move VM in case of DN connected using GRPC protocol (as alpha implementation,
later it will be changed for stability), `vm.toml` contains such default parameters:

```toml
# This is a TOML config file to configurate connection to VM.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# VM network address to connect.
vm_address = "127.0.0.1:50051"

# VM data server listen address.
vm_data_listen = "127.0.0.1:50052"

# VM deploy request timeout in milliseconds.
vm_deploy_timeout = 100

# VM execute contract request timeout in milliseconds.
vm_execute_timeout = 100
```

Where:

* `vm_address` - address of GRPC VM node contains Move VM, using to deploy/execute modules.
* `vm_data_listen` - address to listen for GRPC Data Source server (part of DN), using to share data between DN and VM.

The rest parameters are timeouts, don't recommend to change it.

## Get storage data

It possible to read storage data by path, e.g.:

    dncli query vm get-data [address] [path]

Where:
 * `address` - address of account contains data, could be bech32 or hex (libra).
 * `path` - path of resource, hex.

