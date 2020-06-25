# Governance

Governance module used to change chain parameters "on fly" without node(s) restart.
Changes are submitted via `x/gov` module using proposals.
Proposals submission differs from module to module (if module supports governance) so different CLI commands used.

## Proposal short overview

Proposal can be submitted by any user.

Proposal becomes active after two steps:
1. Deposit period.
    * Proposal has minimal deposit value which should be reached before the deposit period ends, otherwise proposal is rejected;
    * Proposal author could partially cover the deposit value by himself or leave it to other users;
    * Every user can transfer some amount of coins to help the proposal to reach its goal;
    * Deposit period is time limited;
    * Deposit parameters can be viewed via CLI commands;
    * If proposal is rejected, deposits are refunded to their respective depositor; 
2. Voting period.
    * Only "node" accounts (those who have staking voting power) can participate in voting;
    * If user votes "yes" his voting power is added to the overall "yes" counter;
    * Voting ends if 2/3 of users approve that proposal; 
    * Voting period is time limited;
    * Voting parameters can be viewed via CLI commands;

Use the following CLI commands to get all governance parameters and all available queries:

    dncli query gov params
    dncli query gov -h
    
More info could be found in the [Cosmos SDK gov specs](https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/spec/README.md).

## VM module proposals

`x/vm` module supports the following proposals.

### DVM stdlib update

Proposal is used to update DVM standard library code without the chain reboot.
Modules can be updated individually (adding new features for example) and in batch.

    dncli tx vm update-stdlib-proposal ./update.json 1000 http://github.com/repo 'fix for Foo module' --deposit 100dfi --from {accountAddress}

* `./update.json` - path to file containing modules bytecode (precompiled);
* `1000` - scheduled block height;
* `http://github.com/repo` - update source code for reference;
* `"fix for Foo module"` - update short description;
* `--deposit 100dfi` - deposit value (amount is transferred from the proposer);

Stdlib update is verified on proposal submission and scheduled to execute at the specified block height.
