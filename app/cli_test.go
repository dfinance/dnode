package app

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	"github.com/dfinance/dnode/x/oracle"
)

func Test_CurrencyCLI(t *testing.T) {
	ct := NewCLITester(t)
	defer ct.Close()

	fmt.Println("start")

	ccSymbol, ccCurAmount, ccDecimals, ccRecipient := "testcc", sdk.NewInt(1000), int8(1), ct.Accounts["validator1"].Address
	nonExistingAddress := secp256k1.GenPrivKey().PubKey().Address()
	issueID := "issue1"

	// check issue currency multisig Tx
	{
		// submit & confirm call
		ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID).CheckSucceeded()
		ct.WaitForNextBlocks(1)
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
		ct.WaitForNextBlocks(1)
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
			// invalid number of args
			{
				q, _ := ct.QueryCurrenciesIssue(issueID)
				q.RemoveCmdArg(issueID)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
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

		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryCurrenciesCurrency(ccSymbol)
				q.RemoveCmdArg(ccSymbol)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
		}
	}

	// check destroy Query
	{
		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryCurrenciesCurrency(ccSymbol)
				q.RemoveCmdArg(ccSymbol)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
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
	assetCode := "eth_dfi"
	assetOracle1, assetOracle2, assetOracle3 := ct.Accounts["oracle1"].Address, ct.Accounts["oracle2"].Address, ct.Accounts["oracle3"].Address

	// check add asset Tx
	{
		ct.TxOracleAddAsset(nomineeAddr, assetCode, assetOracle1).CheckSucceeded()
		ct.WaitForNextBlocks(1)

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 1)
		require.Equal(t, assetCode, (*assets)[0].AssetCode)
		require.Len(t, (*assets)[0].Oracles, 1)
		require.Equal(t, assetOracle1, (*assets)[0].Oracles[0].Address.String())
		require.True(t, (*assets)[0].Active)

		// check incorrect inputs
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
	}

	// check set asset Tx
	{
		ct.TxOracleSetAsset(nomineeAddr, assetCode, assetOracle1, assetOracle2).CheckSucceeded()
		ct.WaitForNextBlocks(1)

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 1)
		require.Equal(t, assetCode, (*assets)[0].AssetCode)
		require.Len(t, (*assets)[0].Oracles, 2)
		require.Equal(t, assetOracle1, (*assets)[0].Oracles[0].Address.String())
		require.Equal(t, assetOracle2, (*assets)[0].Oracles[1].Address.String())
		require.True(t, (*assets)[0].Active)

		// check incorrect inputs
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

	// check post price Tx
	{
		now := time.Now().Truncate(1 * time.Second).UTC()
		postPrices := []struct {
			assetCode  string
			sender     string
			price      sdk.Int
			receivedAt time.Time
		}{
			{
				assetCode:  assetCode,
				sender:     assetOracle1,
				price:      sdk.NewInt(100),
				receivedAt: now,
			},
			{
				assetCode:  assetCode,
				sender:     assetOracle2,
				price:      sdk.NewInt(150),
				receivedAt: now.Add(1 * time.Second),
			},
		}

		startBlockHeight := ct.WaitForNextBlocks(1)
		for _, postPrice := range postPrices {
			tx := ct.TxOraclePostPrice(postPrice.sender, postPrice.assetCode, postPrice.price, postPrice.receivedAt)
			tx.CheckSucceeded()
		}
		endBlockHeight := ct.WaitForNextBlocks(1)

		// price could be posted in block height range [startBlockHeight:endBlockHeight], so we have to query all
		rawPricesRange := make([]oracle.PostedPrice, 0)
		for i := startBlockHeight; i <= endBlockHeight; i++ {
			q, rawPrices := ct.QueryOracleRawPrices(assetCode, i)
			q.CheckSucceeded()

			rawPricesRange = append(rawPricesRange, *rawPrices...)
		}

		require.Len(t, rawPricesRange, 2)
		for i, postPrice := range postPrices {
			rawPrice := rawPricesRange[i]
			require.Equal(t, postPrice.assetCode, rawPrice.AssetCode)
			require.Equal(t, postPrice.sender, rawPrice.OracleAddress.String())
			require.True(t, postPrice.price.Equal(rawPrice.Price))
			require.True(t, postPrice.receivedAt.Equal(rawPrice.ReceivedAt))
		}

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOraclePostPrice(assetOracle1, assetCode, sdk.OneInt(), time.Now())
				tx.RemoveCmdArg(nomineeAddr)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid price
			{
				tx := ct.TxOraclePostPrice(assetOracle1, assetCode, sdk.OneInt(), time.Now())
				tx.ChangeCmdArg(sdk.OneInt().String(), "not_int")
				tx.CheckFailedWithErrorSubstring("wrong value for price")
			}
			// invalid receivedAt
			{
				now := time.Now()
				tx := ct.TxOraclePostPrice(assetOracle1, assetCode, sdk.OneInt(), now)
				tx.ChangeCmdArg(strconv.FormatInt(now.Unix(), 10), "not_time.Time")
				tx.CheckFailedWithErrorSubstring("wrong value for time")
			}
			// MsgPostPrice ValidateBasic
			{
				tx := ct.TxOraclePostPrice(assetOracle1, assetCode, sdk.NewIntWithDecimal(1, 20), time.Now())
				tx.CheckFailedWithErrorSubstring("bytes limit")
			}
		}
	}

	// check add oracle Tx
	{
		ct.TxOracleAddOracle(nomineeAddr, assetCode, assetOracle3).CheckSucceeded()
		ct.WaitForNextBlocks(2)

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 1)
		require.Equal(t, assetCode, (*assets)[0].AssetCode)
		require.Len(t, (*assets)[0].Oracles, 3)
		require.Equal(t, assetOracle1, (*assets)[0].Oracles[0].Address.String())
		require.Equal(t, assetOracle2, (*assets)[0].Oracles[1].Address.String())
		require.Equal(t, assetOracle3, (*assets)[0].Oracles[2].Address.String())
		require.True(t, (*assets)[0].Active)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOracleAddOracle(nomineeAddr, assetCode, "invalid_address")
				tx.RemoveCmdArg(nomineeAddr)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid oracleAddress
			{
				tx := ct.TxOracleAddOracle(nomineeAddr, assetCode, "invalid_address")
				tx.CheckFailedWithErrorSubstring("oracle_address")
			}
		}
	}

	// check set oracle Tx
	{
		ct.TxOracleSetOracles(nomineeAddr, assetCode, assetOracle3, assetOracle2, assetOracle1).CheckSucceeded()
		ct.WaitForNextBlocks(2)

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 1)
		require.Equal(t, assetCode, (*assets)[0].AssetCode)
		require.Len(t, (*assets)[0].Oracles, 3)
		require.Equal(t, assetOracle3, (*assets)[0].Oracles[0].Address.String())
		require.Equal(t, assetOracle2, (*assets)[0].Oracles[1].Address.String())
		require.Equal(t, assetOracle1, (*assets)[0].Oracles[2].Address.String())
		require.True(t, (*assets)[0].Active)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOracleSetOracles(nomineeAddr, assetCode, assetOracle3, assetOracle2, assetOracle1)
				tx.RemoveCmdArg(nomineeAddr)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid oracles
			{
				tx := ct.TxOracleSetOracles(nomineeAddr, assetCode, "123")
				tx.CheckFailedWithErrorSubstring("")
			}
		}
	}

	// check rawPrices query with invalid arguments
	{
		// invalid number of args
		{
			q, _ := ct.QueryOracleRawPrices(assetCode, 1)
			q.RemoveCmdArg(assetCode)
			q.CheckFailedWithErrorSubstring("arg(s)")
		}
		// invalid blockHeight
		{
			q, _ := ct.QueryOracleRawPrices(assetCode, 1)
			q.ChangeCmdArg("1", "abc")
			q.CheckFailedWithErrorSubstring("blockHeight")
		}
		// blockHeight with no rawPrices
		{
			q, rawPrices := ct.QueryOracleRawPrices(assetCode, 1)
			q.CheckSucceeded()

			require.Empty(t, *rawPrices)
		}
	}

	// check price query
	{
		q, price := ct.QueryOraclePrice(assetCode)
		q.CheckSucceeded()

		require.Equal(t, assetCode, price.AssetCode)
		require.False(t, price.Price.IsZero())

		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryOraclePrice(assetCode)
				q.RemoveCmdArg(assetCode)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing assetCode
			{
				q, _ := ct.QueryOraclePrice("non_existing_assetCode")
				q.CheckFailedWithSDKError(sdk.ErrUnknownRequest(""))
			}
		}
	}
}
