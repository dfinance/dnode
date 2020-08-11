package app

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
)

type ModuleSupplies map[string]sdk.Int

func (s ModuleSupplies) GetSupply(denom string) sdk.Int {
	value, ok := s[denom]
	if !ok {
		return sdk.ZeroInt()
	}

	return value
}

type CalculatedSupply struct {
	Supply   sdk.Int
	Accounts []AccountBalance
}

type AccountBalance struct {
	Name   string
	Amount sdk.Int
}

type CalculatedSupplies map[string]CalculatedSupply

func (s CalculatedSupplies) GetSupply(denom string) CalculatedSupply {
	value, ok := s[denom]
	if !ok {
		return CalculatedSupply{
			Supply:   sdk.ZeroInt(),
			Accounts: nil,
		}
	}

	return value
}

type AllSupplies struct {
	SupplyModule     ModuleSupplies
	CurrenciesModule ModuleSupplies
	Calculated       CalculatedSupplies
}

func (s AllSupplies) String() string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString("All supplies:\n")
	for denom, ccModSupply := range s.CurrenciesModule {
		sModSupply := s.SupplyModule.GetSupply(denom)
		calcSupply := s.Calculated.GetSupply(denom)

		strBuilder.WriteString(fmt.Sprintf("  [%s]:\n", denom))
		strBuilder.WriteString(fmt.Sprintf("    Supply mod amount:     %s\n", sModSupply.String()))
		strBuilder.WriteString(fmt.Sprintf("    Currencies mod amount: %s\n", ccModSupply.String()))
		strBuilder.WriteString(fmt.Sprintf("    Calculated amount:     %s\n", calcSupply.Supply.String()))
		for accIdx, acc := range calcSupply.Accounts {
			strBuilder.WriteString(fmt.Sprintf("    -%s: %s", acc.Name, acc.Amount.String()))
			if accIdx < len(calcSupply.Accounts)-1 {
				strBuilder.WriteString("\n")
			}
		}
	}

	return strBuilder.String()
}

func (s AllSupplies) AreEqual() error {
	var errStrs []string
	for denom, ccModSupply := range s.CurrenciesModule {
		sModSupply := s.SupplyModule.GetSupply(denom)
		calcSupply := s.Calculated.GetSupply(denom)

		if !sModSupply.Equal(calcSupply.Supply) {
			errStr := fmt.Sprintf("%s: supply mod/calculated diff: %s", denom, sModSupply.Sub(calcSupply.Supply).String())
			errStrs = append(errStrs, errStr)
		}

		if !sModSupply.Equal(ccModSupply) {
			errStr := fmt.Sprintf("%s: supply mod/currencies mod diff: %s", denom, sModSupply.Sub(ccModSupply).String())
			errStrs = append(errStrs, errStr)
		}
	}

	if len(errStrs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errStrs, ", "))
}

func (s AllSupplies) GetDiffString(s2 AllSupplies) []string {
	if len(s.CurrenciesModule) != len(s2.CurrenciesModule) {
		return []string{fmt.Sprintf("currencies mod length mismatch: %d / %d", len(s.CurrenciesModule), len(s2.CurrenciesModule))}
	}

	var diffStrs []string
	for denom, ccSupply1 := range s.CurrenciesModule {
		ccSupply2 := s2.CurrenciesModule[denom]
		sModSupply1, sModSupply2 := s.SupplyModule.GetSupply(denom), s2.SupplyModule.GetSupply(denom)
		calcSupply1, calcSupply2 := s.Calculated.GetSupply(denom), s2.Calculated.GetSupply(denom)

		if !sModSupply1.Equal(sModSupply2) {
			diffStrs = append(diffStrs, fmt.Sprintf("%s: supply mod diff: %s", denom, sModSupply2.Sub(sModSupply1).String()))
		}
		if !ccSupply1.Equal(ccSupply2) {
			diffStrs = append(diffStrs, fmt.Sprintf("%s: currencies mod diff: %s", denom, ccSupply2.Sub(ccSupply1).String()))
		}
		if !calcSupply1.Supply.Equal(calcSupply2.Supply) {
			diffStrs = append(diffStrs, fmt.Sprintf("%s: calculated mod diff: %s", denom, calcSupply2.Supply.Sub(calcSupply1.Supply).String()))
		}

		type accDiff struct {
			name   string
			before sdk.Int
			after  sdk.Int
		}
		accDiffs := make(map[string]accDiff)
		for _, acc := range calcSupply1.Accounts {
			diff, ok := accDiffs[acc.Name]
			if !ok {
				diff = accDiff{
					name:  acc.Name,
					after: sdk.ZeroInt(),
				}
			}
			diff.before = acc.Amount
			accDiffs[acc.Name] = diff
		}
		for _, acc := range calcSupply2.Accounts {
			diff, ok := accDiffs[acc.Name]
			if !ok {
				diff = accDiff{
					name:   acc.Name,
					before: sdk.ZeroInt(),
				}
			}
			diff.after = acc.Amount
			accDiffs[acc.Name] = diff
		}
		for _, diff := range accDiffs {
			if !diff.after.Equal(diff.before) {
				diffStrs = append(diffStrs, fmt.Sprintf("%s: account %q diff: %s", denom, diff.name, diff.after.Sub(diff.before).String()))
			}
		}
	}

	return diffStrs
}

// GetAllSupplies returns supply levels for each denom from supply, currencies and vmauth keepers.
func GetAllSupplies(t *testing.T, app *DnServiceApp, ctx sdk.Context) AllSupplies {
	getCalculatedSupplies := func() CalculatedSupplies {
		supplies := make(CalculatedSupplies)
		app.accountKeeper.IterateAccounts(ctx, func(acc authExported.Account) bool {
			accName := ""
			if modAcc, ok := acc.(*supply.ModuleAccount); ok {
				accName = modAcc.GetName()
			} else {
				accName = acc.GetAddress().String()
			}

			for _, coin := range acc.GetCoins() {
				denomSupply, ok := supplies[coin.Denom]
				if !ok {
					denomSupply = CalculatedSupply{
						Supply:   sdk.ZeroInt(),
						Accounts: make([]AccountBalance, 0),
					}
				}

				accBalance := AccountBalance{Name: accName, Amount: coin.Amount}

				denomSupply.Supply = denomSupply.Supply.Add(coin.Amount)
				denomSupply.Accounts = append(denomSupply.Accounts, accBalance)

				supplies[coin.Denom] = denomSupply
			}

			return false
		})
		return supplies
	}

	getModuleSupplies := func() ModuleSupplies {
		supplies := make(ModuleSupplies)
		for _, coin := range app.supplyKeeper.GetSupply(ctx).GetTotal() {
			supplies[coin.Denom] = coin.Amount
		}
		return supplies
	}

	getCCSupplies := func() ModuleSupplies {
		supplies := make(ModuleSupplies)
		for _, currency := range app.ccsKeeper.GetCurrencies(ctx) {
			supplies[currency.Denom] = currency.Supply
		}
		return supplies
	}

	return AllSupplies{
		SupplyModule:     getModuleSupplies(),
		CurrenciesModule: getCCSupplies(),
		Calculated:       getCalculatedSupplies(),
	}
}

// CreateCurrency creates currency with random VM paths.
func CreateCurrency(t *testing.T, app *DnServiceApp, ccDenom string, ccDecimals uint8) {
	params := ccstorage.CurrencyParams{
		Denom:    ccDenom,
		Decimals: ccDecimals,
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	err := app.ccKeeper.CreateCurrency(GetContext(app, false), params)
	require.NoError(t, err, "creating %q currency", ccDenom)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

// IssueCurrency creates currency issue multisig message and confirms it.
func IssueCurrency(t *testing.T, app *DnServiceApp,
	coin sdk.Coin, msgID, issueID string,
	recipientAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doCheck bool) (*sdk.Result, error) {

	issueMsg := currencies.NewMsgIssueCurrency(issueID, coin, accs[recipientAccIdx].Address)
	return MSMsgSubmitAndVote(t, app, msgID, issueMsg, recipientAccIdx, accs, privKeys, doCheck)
}

// WithdrawCurrency creates withdraw currency multisig message and confirms it.
func WithdrawCurrency(t *testing.T, app *DnServiceApp,
	chainID string, coin sdk.Coin,
	spenderAddr sdk.AccAddress, spenderPrivKey crypto.PrivKey, doCheck bool) (*sdk.Result, error) {

	spenderAcc := GetAccountCheckTx(app, spenderAddr)
	withdrawMsg := currencies.NewMsgWithdrawCurrency(coin, spenderAcc.GetAddress(), spenderAcc.GetAddress().String(), chainID)
	tx := GenTx([]sdk.Msg{withdrawMsg}, []uint64{spenderAcc.GetAccountNumber()}, []uint64{spenderAcc.GetSequence()}, spenderPrivKey)

	res, err := DeliverTx(app, tx)
	if doCheck {
		require.NoError(t, err)
	}

	return res, err
}

// CheckCurrencyExists checks currency exists.
func CheckCurrencyExists(t *testing.T, app *DnServiceApp, denom string, supply sdk.Int, decimals uint8) {
	currencyObj := ccstorage.Currency{}
	CheckRunQuery(t, app, currencies.CurrencyReq{Denom: denom}, queryCurrencyCurrencyPath, &currencyObj)

	require.Equal(t, denom, currencyObj.Denom, "denom")
	require.Equal(t, decimals, currencyObj.Decimals, "decimals")
	require.True(t, currencyObj.Supply.Equal(supply), "supply")
}

// CheckIssueExists checks issue exists.
func CheckIssueExists(t *testing.T, app *DnServiceApp, issueID string, coin sdk.Coin, payeeAddr sdk.AccAddress) {
	issue := currencies.Issue{}
	CheckRunQuery(t, app, currencies.IssueReq{ID: issueID}, queryCurrencyIssuePath, &issue)

	require.Equal(t, coin.Denom, issue.Coin.Denom, "coin.Denom")
	require.True(t, coin.Amount.Equal(issue.Coin.Amount), "coin.Amount")
	require.Equal(t, payeeAddr, issue.Payee)
}

// CheckWithdrawExists checks withdraw exists.
func CheckWithdrawExists(t *testing.T, app *DnServiceApp, id uint64, coin sdk.Coin, spenderAddr sdk.AccAddress, pzSpender string) {
	withdraw := currencies.Withdraw{}
	CheckRunQuery(t, app, currencies.WithdrawReq{ID: dnTypes.NewIDFromUint64(id)}, queryCurrencyWithdrawPath, &withdraw)

	require.Equal(t, id, withdraw.ID.UInt64())
	require.Equal(t, coin.Denom, withdraw.Coin.Denom)
	require.True(t, coin.Amount.Equal(withdraw.Coin.Amount))
	require.Equal(t, spenderAddr, withdraw.Spender)
	require.Equal(t, pzSpender, withdraw.PegZoneSpender)
	require.Equal(t, chainID, withdraw.PegZoneChainID)
}

// CheckRecipientCoins checks account balance.
func CheckRecipientCoins(t *testing.T, app *DnServiceApp, recipientAddr sdk.AccAddress, denom string, amount sdk.Int) {
	checkBalance := amount

	coins := app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
	actualBalance := coins.AmountOf(denom)

	require.True(t, actualBalance.Equal(checkBalance), " denom %q, checkBalance / actualBalance mismatch: %s / %s", denom, checkBalance.String(), actualBalance.String())

	balances, err := app.ccsKeeper.GetAccountBalanceResources(GetContext(app, true), recipientAddr)
	require.NoError(t, err, "denom %q: reading balance resources", denom)
	for _, balance := range balances {
		if balance.Denom == denom {
			require.Equal(t, amount.String(), balance.Resource.Value.String(), "denom %q: checking balance resource value", denom)
		}
	}
}
