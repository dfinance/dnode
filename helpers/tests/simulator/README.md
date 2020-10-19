# Simulation parameters

Simulation:
* `ID` - simulation id;
* `SimDuration` - simulation duration;
* `BlockTimeMin` - minimum block duration;
* `BlockTimeMax` - maximum block duration;

Account balances:
* `MainTokens` - initial account balance for main tokens (`xfi`);
* `StakingTokens` - initial account balance for staking tokens (`sxfi`);
* `LPTokens` - initial account balance for LP tokens (`lpt`);

Validators:
* `PoA validator` - number of PegZone validators, ATM has no influence on simulation;
* `TM validators (total)` - total number of PoS validators (some of them might become `unbonded`);
* `TM validators (active)` - number of active PoS validators (they would be `bonded`);

`DelegateBondingOp` / `DelegateLPOp` operation:
* `Delegate bonding/LP tokens every` - period of operation in simulated time;
* `Delegate amount ratio (of acc balance)` - % of account balance for delegation;
* `Max limit ratio (staked ratio)` - operation limit: % of staked tokens to total tokens supply;

`RedelegateBondingOp` / `RedelegateLPOp` operation:
* `Redelegate bonding/LP tokens every` - period of operation in simulated time;
* `Redelegate amount ratio (of del shares)` - % of delegation shares for redelegation;

`UndelegateBondingOp` / `UndelegateLPOp` operation:
* `Undelegate bonding/LP tokens every` - period of operation in simulated time;
* `Undelegate amount ratio (of del shares)` - % of delegation shares for undelegation;

`ValidatorRewardOp` operation:
* `Withdraw all validators comissions every` - period of operation in simulated time;

`DelegatorRewardOp` operation:
* `Withdraw all delegators rewards every` - period of operation in simulated time;

`LockValidatorRewardsOp` operation:
* `Lock rewards every` - period of operation in simulated time;
* `Ratio of all validators` - operation limit: % of all validators to lock rewards;

# Operations

## `DelegateBondingOp` / `DelegateLPOp` operation

Picks a validator and searches for an account to delegate bonding tokens.
* SelfStake increment is allowed;
* Delegation amount = current account balance * {delegateRatio};
* Delegation is allowed if ratio (current staking bonding pools supply / total bonding tokens supply) < {maxBondingRatio};

Op priorities:
- validator:
  - bonded;
  - lowest bonding tokens amount;
- account:
  - highest bonding tokens balance;
  - enough coins;

## `RedelegateBondingOp` / `RedelegateLPOp` operation

* Picks a validator and redelegate bonding tokens to an other validator;
* Redelegation amount = current account delegation amount * {redelegateRatio};

Op priorities:
- dstValidator:
  - bonded;
  - lowest bonding tokens amount;
- srcValidator - highest account delegation bonding shares;
- account:
  - random;
  - has no active redelegations with srcValidator and dstValidator;
  - has enough bonding coins;
  - not a dstValidator owner;

## `UndelegateBondingOp` / `UndelegateLPOp` operation

Picks a validator and undelegates bonding tokens.
* Undelegation amount = current account delegation amount * {undelegateRatio}.

Op priorities:
- validator - highest bonding tokens amount (all statuses);
- account:
  - random;
  - has a validators bonding delegation;
  - not a validator owner;

## `ValidatorRewardOp` operation

Takes all validators commissions rewards.

## `DelegatorRewardOp` operation

Takes all delegators rewards (excluding locked ones).

## `LockValidatorRewardsOp` operation

Takes validator commissions rewards.

Op priority:
- validator - random;

# CSV report

Report item is generated every simulated day.

* `BlockHeight` - block height;
* `BlockTime` - block time;
* `SimDuration` - simulation time (real world time);
* `Validators: Bonded` - number of `bonded` PoS validator;
* `Validators: Unbonding` - number of `unbonding` PoS validator;;
* `Validators: Unbonded` - number of `unbonded` PoS validator;;
* `Staking: Bonded` - amount of bonded staking tokens (`bonded` validators);
* `Staking: NotBonded` - amount of not-bonded staking tokens (`unbonding`/`unbonded` validators);
* `Staking: LPs` - amount of staked LP tokens;
* `Staking: ActiveRedelegations` - number of current redelegations;
* `Mint: MinInflation` - minimum inflation rate;
* `Mint: MaxInflation` - maximum inflation rate;
* `Mint: AnnualProvision` - annual minted tokens estimation (tokens per year);
* `Mint: BlocksPerYear` - number of blocks per year estimation;
* `Dist: FoundationPool` - FoundationPool supply (decimals);
* `Dist: PTreasuryPool` - PublicTreasuryPool supply (decimals);
* `Dist: LiquidityPPool` - LiquidityProvidersPool supply (decimals);
* `Dist: HARP` - HARP supply (decimals);
* `Dist: MAccBalance [main]` - rewards balance keeped by the distribution module (mail tokens);
* `Dist: MAccBalance [staking]` - rewards balance keeped by the distribution module (staking tokens);
* `Dist: BankBalance [main]` - rewards balance keeped by the distribution bank (mail tokens);
* `Dist: BankBalance [staking]` - rewards balance keeped by the distribution bank (staking tokens);
* `Dist: LockedRatio` - rate of bonded delegated tokens for locked validators to all bonded delegated tokens;
* `Supply: Total [main]` - total tokens supply (main tokens);
* `Supply: Total [staking]` - total tokens supply (staking tokens);
* `Supply: Total [LP]` - total tokens supply (LP tokens);
* `Stats: Staked/TotalSupply [staking]` - rate of (bonded + not-bonded tokens) to total supply of staking tokens;
* `Stats: Staked/TotalSupply [LPs]` - rate of staked LP tokens to total supply;
* `Accounts: TotalBalance [main]` - sum of all accounts balances (main tokens);
* `Accounts: TotalBalance [staking]` - sum of all accounts balances (staking tokens);
* `Counters: Bonding: Delegations` - number of bonding delegation operations;
* `Counters: Bonding: Redelegations` - number of bonding redelegation operations;
* `Counters: Bonding: Undelegations` - number of bonding undelegation operations;
* `Counters: LP: Delegations` - number of LP delegation operations;
* `Counters: LP: Redelegations` - number of LP redelegation operations;
* `Counters: LP: Undelegations` - number of LP undelegation operations;
* `Counters: RewardWithdraws` - number of delegators rewards withdraw operations;
* `Counters: RewardsCollected [main]` - accumulated amount of delegators rewards collected (main tokens);
* `Counters: RewardsCollected [staking]` - accumulated amount of delegators rewards collected (staking tokens);
* `Counters: CommissionWithdraws` - number of validators commission rewards operations;
* `Counters: CommissionsCollected [main]` - accumulated amount of validators commission rewards collected (main tokens);
* `Counters: CommissionsCollected [staking]` - accumulated amount of validators commission rewards collected (staking tokens);
* `Counters: LockedRewards` - number of validators rewards lock operations;
