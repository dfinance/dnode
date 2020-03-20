package app

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	ccTypes "github.com/dfinance/dnode/x/currencies/types"
)

func Test_CurrencyCLI(t *testing.T) {
	ct := NewCLITester(t)
	defer ct.Close()

	ccSymbol, ccCurAmount, ccDecimals, ccRecipient := "testcc", sdk.NewInt(1000), int8(1), ct.Accounts["validator1"].Address
	nonExistingAddress := secp256k1.GenPrivKey().PubKey().Address()
	issueID := "issue1"

	// check issue currency multisig Tx
	{
		// submit & confirm call
		ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID).CheckSucceeded()
		ct.WaitForNextNBLocks(1)
		ct.ConfirmCall(issueID)
		// check currency issued
		q, issue := ct.QueryCurrenciesIssue(issueID)
		q.CheckSucceeded()
		require.Equal(t, ccSymbol, issue.Symbol)
		require.True(t, ccCurAmount.Equal(issue.Amount))
		require.Equal(t, ccRecipient, issue.Recipient.String())

		// check incorrect inputs
		{
			// wrong number of args
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID)
				tx.RemoveCmdArg(issueID)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// from non-existing account
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, nonExistingAddress.String(), ccSymbol, ccCurAmount, ccDecimals, issueID)
				tx.CheckFailedWithErrorSubstring("not found")
			}
			// invalid amount
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID)
				tx.ChangeCmdArg(ccCurAmount.String(), "invalid_amount")
				tx.CheckFailedWithErrorSubstring("not a number")
			}
			// invalid decimals
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID)
				tx.ChangeCmdArg(strconv.Itoa(int(ccDecimals)), "invalid_decimals")
				tx.CheckFailedWithErrorSubstring("not a number")
			}
			// invalid recipient
			{
				tx := ct.TxCurrenciesIssue("invalid_addr", ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID)
				tx.CheckFailedWithErrorSubstring("decoding bech32 failed")
			}
			// MsgIssueCurrency ValidateBasic
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, sdk.ZeroInt(), ccDecimals, issueID)
				tx.CheckFailedWithErrorSubstring("wrong amount")
			}
		}
	}

	// check destroy currency Tx
	{
		// reduce amount
		destroyAmount := sdk.NewInt(100)
		ct.TxCurrenciesDestroy(ccRecipient, ccRecipient, ccSymbol, destroyAmount).CheckSucceeded()
		ct.WaitForNextNBLocks(1)
		ccCurAmount = ccCurAmount.Sub(destroyAmount)
		// check destroy
		q, destroy := ct.QueryCurrenciesDestroy(sdk.ZeroInt())
		q.CheckSucceeded()
		require.True(ct.t, sdk.ZeroInt().Equal(destroy.ID))
		require.Equal(ct.t, ccSymbol, destroy.Symbol)
		require.Equal(ct.t, ct.ChainID, destroy.ChainID)
		require.Equal(ct.t, ccRecipient, destroy.Recipient)
		require.Equal(ct.t, ccRecipient, destroy.Spender.String())
		require.True(ct.t, destroyAmount.Equal(destroy.Amount))

		// check incorrect inputs
		{
			// wrong number of args
			{
				tx := ct.TxCurrenciesDestroy(ccRecipient, ccRecipient, ccSymbol, sdk.OneInt())
				tx.RemoveCmdArg(ccSymbol)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// from non-existing account
			{
				tx := ct.TxCurrenciesDestroy(ccRecipient, nonExistingAddress.String(), ccSymbol, sdk.OneInt())
				tx.CheckFailedWithErrorSubstring("not found")
			}
			// invalid amount
			{
				tx := ct.TxCurrenciesDestroy(ccRecipient, ccRecipient, ccSymbol, ccCurAmount)
				tx.ChangeCmdArg(ccCurAmount.String(), "invalid_amount")
				tx.CheckFailedWithErrorSubstring("amount")
			}
			// MsgIssueCurrency ValidateBasic
			{
				tx := ct.TxCurrenciesDestroy(ccRecipient, ccRecipient, ccSymbol, sdk.ZeroInt())
				tx.CheckFailedWithErrorSubstring("wrong amount")
			}
		}
	}

	// check balance
	{
		q, acc := ct.QueryAccount(ccRecipient)
		q.CheckSucceeded()
		require.Len(t, acc.Coins, 2)
		for _, coin := range acc.Coins {
			if coin.Denom != ccSymbol {
				continue
			}

			require.True(t, ccCurAmount.Equal(coin.Amount))
		}
	}

	// check issue Query
	{
		// check incorrect inputs
		{
			// non-existing issueID
			{
				q, _ := ct.QueryCurrenciesIssue("non_existing")
				q.CheckFailedWithSDKError(ccTypes.ErrWrongIssueID(""))
			}
		}
	}

	// check currency Query
	{
		q, currency := ct.QueryCurrenciesCurrency(ccSymbol)
		q.CheckSucceeded()

		require.True(ct.t, currency.CurrencyId.IsZero())
		require.Equal(ct.t, ccSymbol, currency.Symbol)
		require.True(ct.t, ccCurAmount.Equal(currency.Supply))
		require.Equal(ct.t, ccDecimals, currency.Decimals)
	}

	// check destroy Query
	{
		// check incorrect inputs
		{
			// non-existing destroyID
			{
				q, _ := ct.QueryCurrenciesDestroy(sdk.OneInt())
				q.ChangeCmdArg("1", "non_int")
				q.CheckFailedWithErrorSubstring("")
			}
		}
	}

	// check destroys Query
	{
		q, destroys := ct.QueryCurrenciesDestroys(1, 10)
		q.CheckSucceeded()
		require.Len(t, *destroys, 1)

		// check incorrect inputs
		{
			// wrong number of args
			{
				q, _ := ct.QueryCurrenciesDestroys(1, 10)
				q.RemoveCmdArg("10")
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// page / limit
			{
				q, _ := ct.QueryCurrenciesDestroys(-1, 10)
				q.CheckFailedWithErrorSubstring("")
				q, _ = ct.QueryCurrenciesDestroys(1, -1)
				q.CheckFailedWithErrorSubstring("")
			}
		}
	}
}

func Test_OracleCLI(t *testing.T) {
	ct := NewCLITester(t)
	defer ct.Close()

	nomineeAddr := ct.Accounts["oracle1"].Address
	assetCode, assetOracle1, assetOracle2 := "eth_dfi", ct.Accounts["oracle1"].Address, ct.Accounts["oracle2"].Address

	// check add asset Tx
	{
		ct.TxOracleAddAsset(nomineeAddr, assetCode, assetOracle1).CheckSucceeded()
		ct.WaitForNextNBLocks(1)

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 1)
		require.Equal(t, assetCode, (*assets)[0].AssetCode)
		require.Len(t, (*assets)[0].Oracles, 1)
		require.Equal(t, assetOracle1, (*assets)[0].Oracles[0].Address.String())
		require.True(t, (*assets)[0].Active)
	}

	// check add asset with incorrect inputs
	{
		// invalid number of args
		{
			tx := ct.TxOracleAddAsset(nomineeAddr, assetCode, assetOracle1)
			tx.RemoveCmdArg(assetCode)
			tx.CheckFailedWithErrorSubstring("arg(s)")
		}
		// invalid denom
		{
			tx := ct.TxOracleAddAsset(nomineeAddr, "WRONG_ASSET", assetOracle1)
			tx.CheckFailedWithErrorSubstring("non lower case symbol")
		}
		// invalid oracles
		{
			tx := ct.TxOracleAddAsset(nomineeAddr, assetCode, "123")
			tx.CheckFailedWithErrorSubstring("")
		}
		// empty denom
		{
			tx := ct.TxOracleAddAsset(nomineeAddr, "", assetOracle1)
			tx.CheckFailedWithErrorSubstring("denom argument")
		}
		// empty oracles
		{
			tx := ct.TxOracleAddAsset(nomineeAddr, assetCode)
			tx.CheckFailedWithErrorSubstring("oracles argument")
		}
	}

	// check set asset Tx
	{
		ct.TxOracleSetAsset(nomineeAddr, assetCode, assetOracle1, assetOracle2).CheckSucceeded()
		ct.WaitForNextNBLocks(1)

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 1)
		require.Equal(t, assetCode, (*assets)[0].AssetCode)
		require.Len(t, (*assets)[0].Oracles, 2)
		require.Equal(t, assetOracle1, (*assets)[0].Oracles[0].Address.String())
		require.Equal(t, assetOracle2, (*assets)[0].Oracles[1].Address.String())
		require.True(t, (*assets)[0].Active)
	}

	// check set asset with incorrect inputs
	{
		// invalid number of args
		{
			tx := ct.TxOracleSetAsset(nomineeAddr, assetCode, assetOracle1)
			tx.RemoveCmdArg(nomineeAddr)
			tx.CheckFailedWithErrorSubstring("arg(s)")
		}
		// invalid denom
		{
			tx := ct.TxOracleSetAsset(nomineeAddr, "WRONG_ASSET", assetOracle1)
			tx.CheckFailedWithErrorSubstring("non lower case symbol")
		}
		// invalid oracles
		{
			tx := ct.TxOracleSetAsset(nomineeAddr, assetCode, "123")
			tx.CheckFailedWithErrorSubstring("")
		}
		// empty denom
		{
			tx := ct.TxOracleSetAsset(nomineeAddr, "", assetOracle1)
			tx.CheckFailedWithErrorSubstring("denom argument")
		}
		// empty oracles
		{
			tx := ct.TxOracleSetAsset(nomineeAddr, assetCode)
			tx.CheckFailedWithErrorSubstring("oracles argument")
		}
	}
}
