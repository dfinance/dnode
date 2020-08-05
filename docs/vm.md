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
 * `address` - address of account containing data, could be bech32 or hex string (Libra);
 * `path` - resource path, hex string;

## Get storage data LCS (Libra Canonical Serialization) view

If is possible to get VM resource string representation (LCS view) using Move path.
This is similar to using `dncli query vm get-data` command, but VM path is build automatically:

    dncli query vm get-lcs-view [address] [moduleStructMovePath] [viewRequestPath]

Where:
* `address` - address of account containing data (or stdlib address), could be bech32 or hex string (Libra);
* `moduleStructMovePath` - Move resource path;
* `viewRequestPath` - path to file containing LCS view request in JSON format;

Here is an example reading stdlib `Block` resource data:

    dncli query vm get-lcs-view 0x0000000000000000000000000000000000000001 Block::BlockMetadata ./block.json

`block.json` file contains the following request:
```JSON
[
  {
    "name": "height",
    "type": "U64"
  }
]
```

The output in the example above would look like:
```JSON
{
  "Height": 1894
}
```

### LCS view request format

LCS representation doesn't include any additional fields meta data (like JSON/gRPC for instance).
LCS request is a struct schema description used to deserialize the resource data.

Request is the JSON array containing resource fields descriptions:
```JSON
[
  {                             // first resource field description
    "name": "my_vector_field",  // field name (any name)
    "type": "vector",           // field type (supported types)
    "inner_item": [             // nested struct schema used for "vector" and "struct" types (null for others)
      {                         // for "vector" type only one "inner_item" should exist (more for "struct" type)
        "name": "",             // not used for "vector" types, but must be non-empty for "struct" type
        "type": "U64"           // 0x1::Vector<u64>
      }
    ]
  }
]
```

Notes:
* fields order must match resource fields order;
* request must include all resource fields;

#### Supported types

* `U8` - unsigned int with 8 bits;
* `U64` - unsigned int with 64 bits;
* `U128` - unsigned int with 128 bits;
* `bool` - boolean;
* `address` - Libra address;
* `struct` - nested struct (`inner_item` must include nested struct fields schema);
* `vector` - `0x1::Vector` type (`inner_item` must include exactly one field schema);

#### Example

Let's assume we have `Foo` Move module with `Bar` resource :

```Move
address {module_address} {
	module Foo {
        use 0x1::Vector;

        struct Inner {
            a: u8,
            b: bool
        }

	    resource struct Bar {
	        u8Val:    u8,
            u64Val:   u64,
            u128Val:  u128,
            boolVal:  bool,
            addrVal:  address,
            vU8Val:   vector<u8>,
            vU64Val:  vector<u64>,
            inStruct: Inner,
            vComplex: vector<Inner>
        }
	}
}
```

The LCS viewer request containing resource schema would look like:
```JSON
[
    {
        "name": "u8Val",
        "type": "U8",
    },
    {
        "name": "u64Val",
        "type": "U64",
    },
    {
        "name": "u128Val",
        "type": "U128",
    },
    {
        "name": "boolVal",
        "type": "bool",
    },
    {
        "name": "addrVal",
        "type": "address",
    },
    {
        "name": "vectU8Val",
        "type": "vector",
        "inner_item": [
            {
                "type": "U8",
            }
        ]
    },
    {
        "name": "vectU64Val",
        "type": "vector",
        "inner_item": [
            {
                "type": "U64",
            }
        ]
    },
    {
        "name": "innerStruct",
        "type": "struct",
        "inner_item": [
            {
                "name": "a",
                "type": "U8",
            },
            {
                "name": "b",
                "type": "bool",
            }
        ]
    },
    {
        "name": "vectComplex",
        "type": "vector",
        "inner_item": [
            {
                "type": "struct",
                "inner_item": [
                    {
                        "name": "a",
                        "type": "U8",
                    },
                    {
                        "name": "b",
                        "type": "bool",
                    }
                ]
            }
        ]
    }
]
```

The output example:
```JSON
{
    "U8val": 100,
    "U64val": 10000,
    "U128val": 12345678910111213141516171819,
    "Boolval": true,
    "Addrval": [
        220,
        91,
        202,
        217,
        255,
        54,
        112,
        0,
        44,
        56,
        17,
        55,
        236,
        82,
        187,
        52,
        88,
        155,
        113,
        196
    ],
    "Vectu8val": "ZMg=",
    "Vectu64val": [
        1,
        2
    ],
    "Innerstruct": {
        "A": 128,
        "B": false
    },
    "Vectcomplex": [
        {
            "A": 1,
            "B": false
        },
        {
            "A": 2,
            "B": true
        }
    ]
}
```