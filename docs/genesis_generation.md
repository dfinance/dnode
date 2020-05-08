# Genesis Generation

First of all we need to create genesis configuration and name of node:

    dnode init <moniker> --chain-id dn-testnet

Where `<moniker>` must be your node name.

Now configure cli:

    dncli config chain-id dn-testnet
    dncli config output json
    dncli config indent true
    dncli config trust-node true
    dncli config compiler 127.0.0.1:50053
    dncli config node http://127.0.0.1:26657

If you want to keep your keys in file, instead of keystorage of your os, configure it:

    dncli config keyring-backend file

Then let's create 4 accounts, one to store coins, the rest for PoA validators:

    dncli keys add pos
    dncli keys add bank
    dncli keys add nominee
    dncli keys add validator1
    dncli keys add validator2
    dncli keys add validator3

Copy addresses and private keys from output, we will need them in the future.

First of all we create `pos` account, this account will be used later as `Proof of Stake` validator.

As you see we create one account calling `bank` where we will be store all generated **dfi** coins for start,
and then 3 accounts to make them PoA validators, we need at least 3 validators because by default it's a minimum amount of PoA validators to have.

`nominee` is account administrator of oracles system.

Now let's add genesis account and initiate genesis PoA validators and PoS account.

Also to have VM correct work, needs to deploy standard lib write operations.

It should be done before next commands, so see tutorial **[how to initialize genesis for VM](/docs/vm.md#genesis-compilation)**.

    dnode add-genesis-account [pos-address]  1000000000000000000000000dfi
    dnode add-genesis-account [bank-address] 95000000000000000000000000dfi,100000000000000btc,10000000000000usdt
    dnode add-genesis-account [nominee]      1000000000000000000000000dfi
    dnode add-genesis-account [validator-1-address] 1000000000000000000000000dfi
    dnode add-genesis-account [validator-2-address] 1000000000000000000000000dfi
    dnode add-genesis-account [validator-3-address] 1000000000000000000000000dfi

    dnode add-genesis-poa-validator [validator-1-address] [validator-1-eth-address]
    dnode add-genesis-poa-validator [validator-2-address] [validator-2-eth-address]
    dnode add-genesis-poa-validator [validator-3-address] [validator-3-eth-address]

Replace expressions in brackets with correct addresses, include Ethereum addresses.

Now let's register information about added coins in `genesis.json`:

    dnode add-currency-info dfi  18 100000000000000000000000000  018bfc024222e94fbed60ff0c9c1cf48c5b2809d83c82f513b2c385e21ba8a2d35
    dnode add-currency-info eth  18 100000000000000000000000000 01f8799f504905a182aff8d5fc102da1d73b8bec199147bb5512af6e99006baeb6
    dnode add-currency-info btc  8  100000000000000 01fe7c965b1c008c5974c7750959fa10189e803225d5057207563553922a09f906
    dnode add-currency-info usdt 6  10000000000000 0136cb3312422fa6991412077ee93dd9db6cb5b3fcf55750fe2cc739d1d399673b

Time to change denom in PoS configuration.
So open `~/.dnode/config/genesis.json` and find this stake settings:

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

    "bond_denom": "dfi"

By changing this we determine "dfi" as staking currency.

Time to prepare `pos` account (if you're using custom keyring-backend, add `--keyring-backend file` flag):

    dnode gentx --name pos --amount 1000000000000000000000000dfi

After run this command you will see output like:

    Genesis transaction written to "~/.dnode/config/gentx/gentx-<hash>.json"

After you have generated a genesis transaction, you will have to input the genTx into the genesis file, so that DN chain is aware of the validators. To do so, run:

    dnode collect-gentxs

If you want to change VM settings, look at [VM section](#configuration).

Also, you can setup an initial oracles, using next command:

    dnode add-oracle-asset-gen [denom] [oracles]

Where `[denom]` is currency pair, like 'eth_usdt' or 'btc_eth', etc.
And `[oracles]` could be oracles accounts or nominee account, separated by comma.

To make sure that genesis file is correct:

    dnode validate-genesis

Now we are ready to launch testnet:

    dnode start
