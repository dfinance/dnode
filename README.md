# Wings Blockchain / Relay Part

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0)
[![Gitter](https://badges.gitter.im/WingsChat/community.svg)](https://gitter.im/WingsChat/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

**THIS IS VERY EARLY WORK IN PROGRESS, NOT FOR TESTNET/PRODUCTION USAGE**

Wings Blockchain (WB) is based on [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).

This is work in progress, yet it supports the following features:

* **Proof Of Authority** (PoA) validators mechanism.
* **N/2+1** confirmations model.
* **Multisignature** based on PoA validators.
* Managing of validators state by PoA consensus.
* Execution of messages (transactions) based on PoA consensus.
* Issuing/destroying new coins based on PoA consensus.
* **86400** blocks interval to confirm call execution under multisig.
* **Support PoS**: staking, delegation, slashing, supply, etc.
* **Supports Smart Contracts**: Move Virtual Machine developed by Libra (Facebook).

Motivation is allowing to implement DeFi products without headache.

Additional information could be found in other repositories, that presents part of WB.

WB (Wings Blockchain) is technical name and will be changed in future.

Other repositories related to Peg Zones could be found at [project page](https://github.com/WingsDao).

# Installation

Before we start you should have a correct 'GOPATH', 'GOROOT' environment variables.

To install fetch this repository:

    git clone git@github.com:WingsDao/blockchain-relay-layer.git

And let's build both daemon and cli:

    GO111MODULE=on go build cmd/wbd/main.go
    GO111MODULE=on go build cmd/wbcli/main.go

Both commands must execute fine, after it you can run both daemon and cli:

    GO111MODULE=on go run cmd/wbd/main.go
    GO111MODULE=on go run cmd/wbcli/main.go

## Install as binary

To install both cli and daemon as binaries you can use Makefile:

    make install

So after this command both `wbd` and `wbcli` will be available from console.

# Usage

First of all we need to create genesis configuration and name of node:

    wbd init <moniker> --chain-id wings-testnet

Where `<moniker>` must be your node name.

Then let's create 4 accounts, one to store coins, the rest for PoA validators:

    wbcli keys add pos
    wbcli keys add bank
    wbcli keys add validator1
    wbcli keys add validator2
    wbcli keys add validator3

Copy addresses and private keys from output, we will need them in the future.

First of all we create `pos` account, this account will be used later as `Proof of Stake` validator.

As you see we create one account calling `bank` where we will be store all generated **WINGS** coins for start,
and then 3 accounts to make them PoA validators, we need at least 3 validators because by default it's a minimum amount of PoA validators to have.

Now let's add genesis account and initiate genesis PoA validators and PoS account.

Also to have VM correct work, needs to deploy standard lib write operations.

It should be done before next commands, so see tutorial **[how to initialize genesis for VM](#genesis-compilation)**.

    wbd add-genesis-account [pos-address]  5000000000000wings
    wbd add-genesis-account [bank-address] 90000000000000000000000000wings
    wbd add-genesis-account [validator-1-address]  5000000000000wings
    wbd add-genesis-account [validator-2-address]  5000000000000wings
    wbd add-genesis-account [validator-3-address]  5000000000000wings

    wbd add-genesis-poa-validator [validator-1-address] [validator-1-eth-address]
    wbd add-genesis-poa-validator [validator-2-address] [validator-2-eth-address]
    wbd add-genesis-poa-validator [validator-3-address] [validator-3-eth-address]

Replace expressions in brackets with correct addresses, include Ethereum addresses.

Now configure cli:

    wbcli config chain-id wings-testnet
    wbcli config output json
    wbcli config indent true
    wbcli config trust-node true
    wbcli config compiler 127.0.0.1:50053

Time to change denom in PoS configuration.
So open `~/.wbd/config/genesis.json` and find this stake settings:

```json
"staking": {
  "params": {
    "unbonding_time": "1814400000000000",
    "max_validators": 100,
    "max_entries": 7,
    "bond_denom": "stake"
  },
  "last_total_power": "0",
  "last_validator_powers": null,
  "validators": null,
  "delegations": null,
  "unbonding_delegations": null,
  "redelegations": null,
  "exported": false
}
```

Change line:

    "bond_denom": "stake"
To:

    "bond_denom": "wings"

By changing this we determine "wings" as staking currency.

Time to prepare `pos` account:

    wbd gentx --name pos --amount 5000000000000wings

After run this command you will see output like:

    Genesis transaction written to "~/.wbd/config/gentx/gentx-<hash>.json"

After you have generated a genesis transaction, you will have to input the genTx into the genesis file, so that WB chain is aware of the validators. To do so, run:

    wbd collect-gentxs
   
To make sure that genesis file is correct:

    wbd validate-genesis

If you want to change VM settings, look at [VM section](#configuration).

Now we are ready to launch testnet:

    wbd start

# Docs

## Peg Zone

### Add/remove/replace validator by multisignature

Before we start managing validators by PoA, let's remember that minimum amount of validators is 3, maximum is 11.

To add new validator use next command:

    wbcli tx poa ms-add-validator [validator-address] [eth-address] --validator-1

Where:

* **[validator-address]** - WB bench32 validator address
* **[eth-address]** - validator ethereum address

To remove:

    wbcli tx poa ms-remove-validator [validator-address] --from validator1

To replace:

    wbcli tx poa ms-replace-validator [old-address] [new-address] [eth-address] --from validator-1

To get validators list (include their amount and required confirmations amount to execute call):

    wbcli query poa validators

To get minimum/maximum validators amount:

    wbcli query poa minmax

To get validator:

    wbcli query poa validator [address]

Where `[address]` is Bech32 WB address.

### Confirm multisignature call

To confirm multisignature call you need to extract call id from transaction execution output and confirm this call
by other validators:

    wbcli tx multisig confirm-call [call-id]

Once call submited under multisignature, there is **86400** blocks interval to confirm it by other validators, if call
not confirmed by that time, it will be marked as rejected.

To revoke confirmation from call:

    wbcli tx multisig revoke-confirm [call-id]

Once call reaches **N/2+1** amount of confirmations, message inside call will be executed.

To get call information:

    wbcli query multisig call [call-id]

To get calls amount:

    wbcli query multisig lastId

### Issuing new currency by multisig

To issue new currency:

    wbcli tx currencies ms-issue-currency [currencyId] [symbol] [amount] [decimals] [recipient] [issueID] [uniqueID]  --from validators1

Where:

| parameter | desc                                                                                                                        |
|----------------|-----------------------------------------------------------------------------------------------------------------------------|
| **currencyId** | Currency ID.                                                                                                                |
| **symbol**     | Currency symbol/denom to issue.                                                                                             |
| **amount**     | Amount to issue.                                                                                                            |
| **decimals**   | Currency decimals, maximum is 8.                                                                                            |
| **recipient**  | WB address of account who's receiving coins.                                                                            |
| **issueID**    | Any issue id, usually transaction id.                                                                                       |
| **uniqueID**   | Call unique id, required to prevent double spend on issuing new currencies, usually it's sha256(chainId + symbol + txHash), serialized to hex. |

To destroy currency from any account call:

    wbcli tx currencies destroy-currency [symbol] [amount] [recipient] --from account

To get issued currencies demons/symbols:

    wbcli query currencies currency [symbol]

To get specific issue info:

    wbcli query currencies issue [issueID]

To get destroys list:

    wbcli query currencies destroys [page] [limit]

Where:

* **[page]** - page number
* **[limit]** - limit of destroys per page

To get destroy by ID:

    wbcli query currencies destroy [destroyID]

Where:

* **[destroyID]** - destroy ID, usually just from 0 to N.

### Rest API

Launch REST API:

    wbcli rest-server --chain-id wings-testnet --trust-node

All REST API returns JSON.

Multisig:

* `/multisig/call/{id}`  - get call by id.
* `/multisig/calls`      - get array of active calls (that waiting for confirmations)
* `/multisig/unique/{unique}` - get call by unique id.

Currencies:

* `/currencies/issue/{issueID}` - Get issue operation by issue id.
* `/currencies/currency/{symbol}` - Get currency info by symbol.
* `/currencies/destroy/{destroyID}` - Get destroy info by id.
* `/currencies/destroys/{page}?limit={limit}` - Get destroys list, limit parameter is optional, equal 100 by default.

PoA:

* `/poa/validators` - PoA validators list.


## Fees

Currently WB supports transactions only with non-zero fees in wings cryptocurrency, so it means each transaction
must contains at least **1wings**.

## VM

WB blockchain currently supports smart-contracts via Move VM.

Both two types of Move transaction supported, like: deploy module/execute script.

To deploy module:

    wbcli tx vm deploy-module [fileMV] --from <from> --fees <fees>
    
To execute script:

    wbcli tx vm execute-script [fileMV] arg1:type1, arg2:type2, arg3:type3... --from <from> --fees <fees>
    
    # Or (as example with arguments):
    wbcli tx vm execute-script [fileMV] true:Bool, 150:U64 --from <from> --fees <fees>
    
To get results of execution, gas spent, events, just query transaction:

    wbcli query tx [transactionId]

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

### Genesis compilation

First of all, to get WB correctly work, needs to compile standard WB smart modules libs,
and put result into genesis block. Result is WriteSet operations, that write compiled modules 
into storage.

So, first of all, go to VM folder, and run:

    cargo run --bin stdlib-builder stdlib/mvir mvir -po ../genesis-ws.json

After this, go into WB folder and run:

    wbd read-genesis-write-set [path to created file genesis-ws.json]

Now everything should be fine.

### Compilation

Launch compiler server, and WB.

Then use commands to compile modules/scripts:

    wbcli query vm compile-script [mvirFile] [address] --to-file <script.mv> --compiler 127.0.0.1:50053
    wbcli query vm compile-module [mvirFile] [address] --to-file <module.mv> --compiler 127.0.0.1:50053    

Where:
 * `mvirFile` - file contains MVir code.
 * `address` - address of account who will use compiled code.
 * `--to-file` - allows to output result to file, otherwise it will be printed in console.
 * `--compiler` - address of compiler, could be ignored, default is `127.0.0.1:50053`.

### Configuration

Default VM configuration file placed into `~/.wbd/config/vm.toml`, and will be 
initialized after `init` command.

As Move VM in case of WB connected using GRPC protocol (as alpha implementation,
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
* `vm_data_listen` - address to listen for GRPC Data Source server (part of WB), using to share data between WB and VM.

The rest parameters are timeouts, don't recommend to change it.

### Tests

During standard launch of tests:

    GO111MODULE=on go test ./...

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

     GO111MODULE=on go test wings-blockchain/x/vm/internal/keeper

# Get storage data

It possible to read storage data by path, e.g.:

    wbcli query vm get-data [address] [path]

Where:
 * `address` - address of account contains data, could be bech32 or hex (libra).
 * `path` - path of resource, hex.

# Tests

To launch tests run: 

    GO111MODULE=on go test ./...
    
    And with integration tests:
    GO111MODULE=on go test ./... --tags integ

# Contributors

This project has the [following contributors](https://github.com/WingsDao/wings-blockchain/graphs/contributors).

To help project you always can open [issue](https://github.com/WingsDao/wings-blockchain/pulls) or fork, do changes in your own fork and open [pull request](https://github.com/WingsDao/wings-blockchain/pulls).

# License

Copyright © 2019 Wings Foundation

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the [GNU General Public License](https://github.com/WingsDAO/griffin-consensus-poc/tree/master/LICENSE) along with this program.  If not, see <http://www.gnu.org/licenses/>.
