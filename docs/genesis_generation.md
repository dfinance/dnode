# Genesis Generation

First of all we need to create genesis configuration and set name for the node:

    dnode init <moniker> --chain-id dn-testnet

* **<moniker>** - node name;

Now configure CLI client:

    dncli config chain-id dn-testnet
    dncli config output json
    dncli config indent true
    dncli config trust-node true
    dncli config compiler tcp://127.0.0.1:50051
    dncli config node http://127.0.0.1:26657

If you want to keep your keys in file based storage, instead of OS keystorage, configure it:

    dncli config keyring-backend file

Then let's create accounts:

    dncli keys add pos
    dncli keys add bank
    dncli keys add nominee
    dncli keys add validator1
    dncli keys add validator2
    dncli keys add validator3
    dncli keys add orders
    dncli keys add gov

Copy addresses and private keys from output, they would be used later.

* **pos** - account used by `Proof of Stake` validator;
* **bank** - account used to store all generated **dfi** coins;
* **validator1..3**` - accounts used as `Proof of Authority` validators for PegZone integration (3 account created as by default this is the minimum number);
* **nominee** - administrator account for oracles system;
* **orders** - module account used for DEX system to lock coins on order posting;
* **gov** - module account used by Governance system to lock / refund deposits;

Now let's add genesis accounts and initiate genesis PoA validators.

    dnode add-genesis-account [pos-address]  1000000000000000000000000dfi
    dnode add-genesis-account [bank-address] 95000000000000000000000000dfi,100000000000000btc,10000000000000usdt
    dnode add-genesis-account [nominee]      1000000000000000000000000dfi
    dnode add-genesis-account [validator-1-address] 1000000000000000000000000dfi
    dnode add-genesis-account [validator-2-address] 1000000000000000000000000dfi
    dnode add-genesis-account [validator-3-address] 1000000000000000000000000dfi
    dnode add-genesis-account [orders-address] 1000000000000000000000000dfi --module-name orders
    dnode add-genesis-account [gov-address] 1000000000000000000000000dfi --module-name gov

    dnode add-genesis-poa-validator [validator-1-address] [validator-1-eth-address]
    dnode add-genesis-poa-validator [validator-2-address] [validator-2-eth-address]
    dnode add-genesis-poa-validator [validator-3-address] [validator-3-eth-address]

Replace expressions in brackets with correct addresses, include Ethereum addresses.

For VM to work correctly, we need to deploy standard library write sets.
It should be done before the next commands, refer to the tutorial **[how to initialize genesis for VM](/docs/vm.md#genesis-compilation)**.

The following commands might be omitted as [`dfi`, `eth`, `usdt`, `btc`] currencies already exist in the default generated genesis above.
Currencies can be added to the chain later using `gov` proposals.

    dnode set-currency-info dfi  18 01608540feb9c6bd277405cfdc0e9140c1431f236f7d97865575e830af3dd67e7e 01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d
    dnode set-currency-info eth  18 0138f4f2895881c804de0e57ced1d44f02e976f9c6561c889f7b7eef8e660d2c9a 012a00668b5325f832c28a24eb83dffa8295170c80345fbfbf99a5263f962c76f4
    dnode set-currency-info usdt 6  01a04b6467f35792e0fda5638a509cc807b3b289a4e0ea10794c7db5dc1a63d481 01d058943a984bc02bc4a8547e7c0d780c59334e9aa415b90c87e70d140b2137b8
    dnode set-currency-info btc  8  019a2b233aea4cab2e5b6701280f8302be41ea5731af93858fd96e038499eda072 019fdf92aeba5356ec5455b1246c2e1b71d5c7192c6e5a1b50444dafaedc1c40c9

We can also add DEX markets to genesis (markets can be added later via non-genesis Tx command as well):

    dnode add-market-gen eth dfi
    dnode add-market-gen btc dfi
    dnode add-market-gen usdt dfi

Time to prepare `pos` account (if you're using custom keyring-backend, add `--keyring-backend file` flag):

    dnode gentx --name pos --amount 1000000000000000000000000dfi

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
