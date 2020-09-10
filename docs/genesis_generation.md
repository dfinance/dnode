# Genesis Generation

First of all we need to create genesis configuration and set name for the node:

    dnode init <moniker> --chain-id dn-testnet

* **<moniker>** - node name;

Now configure CLI client:

    dncli config chain-id dn-testnet
    dncli config output json --home damir/.dncli 
    dncli config indent true --home damir/.dncli
    dncli config trust-node true --home damir/.dncli
    dncli config compiler tcp://127.0.0.1:50051 --home damir/.dncli
    dncli config node http://127.0.0.1:26657 --home damir/.dncli

If you want to keep your keys in file based storage, instead of OS keystorage, configure it:

    dncli config keyring-backend file

Then let's create accounts:

    dncli keys add pos
    dncli keys add bank
    dncli keys add nominee
    dncli keys add validator1
    dncli keys add orders
    dncli keys add gov

Copy addresses and private keys from output, they would be used later.

* **pos** - account used by `Proof of Stake` validator;
* **bank** - account used to store all generated **xfi** coins;
* **validator1..3**` - accounts used as `Proof of Authority` validators for PegZone integration (3 account created as by default this is the minimum number);
* **nominee** - administrator account for oracles system;
* **orders** - module account used for DEX system to lock coins on order posting;
* **gov** - module account used by Governance system to lock / refund deposits;

Now let's add genesis accounts and initiate genesis PoA validators.

    dnode add-genesis-account [pos-address]  1000000000000000000000000sxfi
    dnode add-genesis-account [bank-address] 95000000000000000000000000xfi,100000000000000btc,10000000000000usdt
    dnode add-genesis-account [nominee]      1000000000000000000000000xfi
    dnode add-genesis-account [validator-1-address] 1000000000000000000000000xfi
    dnode add-genesis-account [orders-address] 1000000000000000000000000xfi --module-name orders
    dnode add-genesis-account [gov-address] 1000000000000000000000000xfi --module-name gov

    dnode add-genesis-poa-validator [validator-1-address] [validator-1-eth-address]

Replace expressions in brackets with correct addresses, include Ethereum addresses.

For VM to work correctly, we need to deploy standard library write sets.
It should be done before the next commands, refer to the tutorial **[how to initialize genesis for VM](/docs/vm.md#genesis-compilation)**.

The following commands might be omitted as [`xfi`, `eth`, `usdt`, `btc`] currencies already exist in the default generated genesis above.
Currencies can be added to the chain later using `gov` proposals.

    dnode set-currency sxfi  18  --home damir/.dnode
    dnode set-currency xfi  18  --home damir/.dnode
    dnode set-currency eth  18  --home damir/.dnode
    dnode set-currency usdt 6  --home damir/.dnode
    dnode set-currency btc  8  --home damir/.dnode

We can also add DEX markets to genesis (markets can be added later via non-genesis Tx command as well):

    dnode add-market-gen eth xfi
    dnode add-market-gen btc xfi
    dnode add-market-gen usdt xfi

Time to prepare `pos` account (if you're using custom keyring-backend, add `--keyring-backend file` flag):

    dnode gentx --name pos --amount 1000000000000000000000000sxfi --min-self-delegation 1000000000000000000000000 --home damir/.dnode --keyring-backend file --home-client damir/.dncli

The output like:

    Genesis transaction written to "~/.dnode/config/gentx/gentx-<hash>.json"

After you have generated a genesis transaction, you will have to input the genTx into the genesis file, so that DN chain is aware of the validators:

    dnode collect-gentxs

If you want to change VM settings, refer to [VM section](#configuration).

You could also setup an initial oracles, using the next command:

    dnode add-oracle-asset-gen [denom] [oracles]

* **denom**` - currency pair code(`eth_usdt`, `btc_eth`, etc);
* **oracles**` - oracle accounts or nominee account, separated by a comma;

To make sure genesis file is correct:

    dnode validate-genesis

Now we are ready to launch the testnet node:

    dnode start
