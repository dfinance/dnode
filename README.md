# Wings Blockchain / Relay Part

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0)
[![Gitter](https://badges.gitter.im/WingsChat/community.svg)](https://gitter.im/WingsChat/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

**THIS IS VERY EARLY WORK IN PROGRESS, NOT FOR TESTNET/PRODUCTION USAGE**

Wings Blockchain Peg Zone implementation based on [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).

This is work in progress, yet it supports the following features:

* **Proof Of Authority** (PoA) validators mechanism
* **N/2+1** confirmations model
* **Multisignature** based on PoA validators
* Managing of validators state by PoA consensus
* Execution of messages (transactions) based on PoA consensus
* Issuing/destroying new coins based on PoA consensus
* **86400** blocks interval to confirm call execution under multisig

Motivation is allowing to moving tokens/currencies between different blockchains and Wings blockchain.

Additional information could be found in other repositories, that presents part of Wings Peg Zones.

Other repositories related to Peg Zones could be found:

* [Ethereum Peg Zone](https://github.com/WingsDao/eth-peg-zone)

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

First of all we need to create genesis configuration:

    wbd init --chain-id wings-testnet

Then let's create 4 accounts, one to store coins, the rest for PoA validators:

    wbcli keys add bank
    wbcli keys add validator1
    wbcli keys add validator2
    wbcli keys add validator3

Copy addresses and private keys from output, we will need them in the future.

As you see we create one account calling `bank` where we will be store all generated **WINGS** coins for start,
and then 3 accounts to make them PoA validators, we need at least 3 validators because by default it's a minimum amount of PoA validators to have.

Now let's add genesis account and initiate genesis PoA validators:

    wbd add-genesis-account [bank-address] 10000wings

    wbd add-genesis-poa-validator [validator-1-address] [validator-1-eth-address]
    wbd add-genesis-poa-validator [validator-2-address] [validator-2-eth-address]
    wbd add-genesis-poa-validator [validator-3-address] [validator-3-eth-address]

Replace expressions in brackets with correct addresses, include Ethereum addresses, configure chain by Cosmos SDK documentation:

    wbcli config chain-id wings-testnet
    wbcli config output json
    wbcli config indent true
    wbcli config trust-node true

Now we are ready to launch testnet:

    wbd start

Deposit validators accounts by sending them **WINGS**:

    wbcli tx send [validator-n] 10wings --from bank

## Add/remove/replace validator by multisignature

Before we start managing validators by PoA, let's remember that minimum amount of validators is 3, maximum is 11.

To add new validator use next command:

    wbcli tx poa ms-add-validator [validator-address] [eth-address] --validator-1

Where:

* **[validator-address]** - cosmos bench32 validator address
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

Where `[address]` is Bech32 Cosmos address.

## Confirm multisignature call

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

## Issuing new currency by multisig

To issue new currency:

    wbcli tx currencies ms-issue-currency [currencyId] [symbol] [amount] [decimals] [recipient] [issueID] [uniqueID]  --from validators1

Where:

| parameter | desc                                                                                                                        |
|----------------|-----------------------------------------------------------------------------------------------------------------------------|
| **currencyId** | Currency ID.                                                                                                                |
| **symbol**     | Currency symbol/denom to issue.                                                                                             |
| **amount**     | Amount to issue.                                                                                                            |
| **decimals**   | Currency decimals, maximum is 8.                                                                                            |
| **recipient**  | Cosmos address of account who's receiving coins.                                                                            |
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

# Rest API

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

# Docs

In progress.

# Tests

In progress.

# Contributors

Current project is under development and going to evolve together with other parts of Wings blockchain as
**Relay Layer** and Wings blockchain itself, anyway we have
planned things to:

* More Tests Coverage
* Refactoring
* Generate docs
* PoS government implementation instead of PoA

This project has the [following contributors](https://github.com/WingsDao/griffin-consensus-poc/graphs/contributors).

# License

Copyright © 2019 Wings Foundation

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the [GNU General Public License](https://github.com/WingsDAO/griffin-consensus-poc/tree/master/LICENSE) along with this program.  If not, see <http://www.gnu.org/licenses/>.
