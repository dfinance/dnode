package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

type (
	// Operations order:
	//   1: addAccountOps
	//   2: accountBalanceOps
	SquashOptions struct {
		// Add account operations
		addAccountOps []addAccountOperation
		// Account balance modification operations
		accountBalanceOps []accountBalanceOperation
	}

	addAccountOperation struct {
		// Account address
		Address sdk.AccAddress
		// Account balance
		Coins sdk.Coins
	}

	accountBalanceOperation struct {
		// Coin denom
		Denom string
		// Remove coin balance
		// 1st priority
		Remove bool
		// Rename coin / move balance (empty - no renaming)
		// 2nd priority
		RenameTo string
	}
)

func (opts *SquashOptions) SetAddAccountOp(addressRaw, coinsRaw string) error {
	op := addAccountOperation{}

	addr, err := sdk.AccAddressFromBech32(addressRaw)
	if err != nil {
		return fmt.Errorf("address (%s): invalid AccAddress: %w", addressRaw, err)
	}
	op.Address = addr

	coins, err := sdk.ParseCoins(coinsRaw)
	if err != nil {
		return fmt.Errorf("coins (%s): sdk.Coins parsing failed: %w", coinsRaw, err)
	}
	op.Coins = coins

	opts.addAccountOps = append(opts.addAccountOps, op)

	return nil
}

func (opts *SquashOptions) SetAccountBalanceOp(denomRaw string, remove bool, renameToRaw string) error {
	op := accountBalanceOperation{}
	op.Remove = remove

	if remove && renameToRaw != "" {
		return fmt.Errorf("remove op can not coexist with rename op")
	}

	if err := sdk.ValidateDenom(denomRaw); err != nil {
		return fmt.Errorf("denom (%s): invalid: %w", denomRaw, err)
	}
	op.Denom = denomRaw

	if renameToRaw != "" {
		if err := sdk.ValidateDenom(renameToRaw); err != nil {
			return fmt.Errorf("renameTo denom (%s): invalid: %w", renameToRaw, err)
		}
		op.RenameTo = renameToRaw
	}

	opts.accountBalanceOps = append(opts.accountBalanceOps, op)

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		addAccountOps:     nil,
		accountBalanceOps: nil,
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k VMAccountKeeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	// addAccountOps
	for i, accOpt := range opts.addAccountOps {
		acc := k.NewAccountWithAddress(ctx, accOpt.Address)
		if err := acc.SetCoins(accOpt.Coins); err != nil {
			return fmt.Errorf("addAccountOps[%d]: SetCoins: %w", i, err)
		}
		k.SetAccount(ctx, acc)
	}

	// accountBalanceOps
	{
		// remove ops
		for i, op := range opts.accountBalanceOps {
			if !op.Remove {
				continue
			}

			var opErr error
			k.IterateAccounts(ctx, func(acc authexported.Account) (stop bool) {
				coins := acc.GetCoins()
				coinToDel := sdk.NewCoin(op.Denom, sdk.ZeroInt())
				for _, coin := range coins {
					if coin.Denom != op.Denom {
						continue
					}
					coinToDel.Amount = coin.Amount
					break
				}

				coins = coins.Sub(sdk.NewCoins(coinToDel))
				if err := acc.SetCoins(coins); err != nil {
					opErr = fmt.Errorf("accountBalanceOps[%d] (%s) remove: SetCoins: %w", i, acc.GetAddress(), err)
					return true
				}
				k.SetAccount(ctx, acc)

				return false
			})
			if opErr != nil {
				return opErr
			}
		}

		// rename ops
		for i, op := range opts.accountBalanceOps {
			if op.RenameTo == "" {
				continue
			}

			var opErr error
			k.IterateAccounts(ctx, func(acc authexported.Account) (stop bool) {
				coins := acc.GetCoins()
				oldCoin := sdk.NewCoin(op.Denom, sdk.ZeroInt())
				for _, coin := range coins {
					if coin.Denom != op.Denom {
						continue
					}
					oldCoin.Amount = coin.Amount
					break
				}
				newCoin := sdk.NewCoin(op.RenameTo, oldCoin.Amount)

				coins = coins.Sub(sdk.NewCoins(oldCoin))
				coins = coins.Add(newCoin)
				if err := acc.SetCoins(coins); err != nil {
					opErr = fmt.Errorf("accountBalanceOps[%d] (%s) rename: SetCoins: %w", i, acc.GetAddress(), err)
					return true
				}
				k.SetAccount(ctx, acc)

				return false
			})
			if opErr != nil {
				return opErr
			}
		}
	}

	return nil
}
