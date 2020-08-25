// +build cli

package app

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmCoreTypes "github.com/tendermint/tendermint/rpc/core/types"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/poa"
)

const (
	NotFoundErrSubString = "The specified item could not be found in the keyring"
)

func TestCurrencies_CLI(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	wsStop, wsChs := ct.CheckWSsSubscribed(false, "TestCurrencies_CLI", []string{"message.module='currencies'"}, 10)
	defer wsStop()
	go cliTester.PrintEvents(t, wsChs, "currencies", "multisig")

	ccDenom := "btc"
	ccDecimals := ct.Currencies[ccDenom].Decimals
	ccCurAmount, ccRecipient := sdk.NewInt(1000), ct.Accounts["validator1"].Address
	nonExistingAddress := secp256k1.GenPrivKey().PubKey().Address()
	issueID := "issue1"

	// check issue currency multisig Tx, issue Query, currency Query
	{
		// submit & confirm call
		{
			ct.TxCurrenciesIssue(ccRecipient, ccRecipient, issueID, ccDenom, ccCurAmount).CheckSucceeded()
			ct.ConfirmCall(issueID)
		}

		// check issue appeared
		{
			q, currency := ct.QueryCurrenciesCurrency(ccDenom)
			q.CheckSucceeded()

			require.Equal(t, ccDenom, currency.Denom)
			require.True(t, ccCurAmount.Equal(currency.Supply))
			require.Equal(t, ccDecimals, currency.Decimals)
		}

		// check currency issued
		{
			q, issue := ct.QueryCurrenciesIssue(issueID)
			q.CheckSucceeded()

			require.Equal(t, ccDenom, issue.Coin.Denom)
			require.True(t, ccCurAmount.Equal(issue.Coin.Amount))
			require.Equal(t, ccRecipient, issue.Payee.String())
		}

		// incorrect inputs
		{
			// wrong number of args
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, issueID, ccDenom, ccCurAmount)
				tx.RemoveCmdArg(issueID)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// from non-existing account
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, nonExistingAddress.String(), issueID, ccDenom, ccCurAmount)
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid amount
			{
				coin := sdk.NewCoin(ccDenom, ccCurAmount)
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, issueID, ccDenom, ccCurAmount)
				tx.ChangeCmdArg(coin.String(), "invalid_amount"+ccDenom)
				tx.CheckFailedWithErrorSubstring("parsing coin")
			}
			// invalid recipient
			{
				tx := ct.TxCurrenciesIssue("invalid_addr", ccRecipient, issueID, ccDenom, ccCurAmount)
				tx.CheckFailedWithErrorSubstring("Bech32 / HEX")
			}
			// MsgIssueCurrency ValidateBasic
			{
				tx := ct.TxCurrenciesIssue(ccRecipient, ccRecipient, issueID, ccDenom, sdk.ZeroInt())
				tx.CheckFailedWithErrorSubstring("wrong amount")
			}
		}
	}

	// check withdraw currency Tx
	{
		// reduce amount
		withdrawAmount := sdk.NewInt(100)
		{
			ct.TxCurrenciesWithdraw(ccRecipient, ccRecipient, ccDenom, withdrawAmount).CheckSucceeded()
			ccCurAmount = ccCurAmount.Sub(withdrawAmount)
		}

		// check withdraw appeared
		{
			id := dnTypes.NewIDFromUint64(0)
			q, withdraw := ct.QueryCurrenciesWithdraw(id)
			q.CheckSucceeded()

			require.True(t, withdraw.ID.Equal(id))
			require.Equal(t, ccDenom, withdraw.Coin.Denom)
			require.Equal(t, ct.IDs.ChainID, withdraw.PegZoneChainID)
			require.Equal(t, ccRecipient, withdraw.PegZoneSpender)
			require.Equal(t, ccRecipient, withdraw.Spender.String())
			require.True(t, withdrawAmount.Equal(withdraw.Coin.Amount))
		}

		// incorrect inputs
		{
			// wrong number of args
			{
				coin := sdk.NewCoin(ccDenom, sdk.OneInt())
				tx := ct.TxCurrenciesWithdraw(ccRecipient, ccRecipient, ccDenom, sdk.OneInt())
				tx.RemoveCmdArg(coin.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// from non-existing account
			{
				tx := ct.TxCurrenciesWithdraw(ccRecipient, nonExistingAddress.String(), ccDenom, sdk.OneInt())
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid amount
			{
				coin := sdk.NewCoin(ccDenom, ccCurAmount)
				tx := ct.TxCurrenciesWithdraw(ccRecipient, ccRecipient, ccDenom, ccCurAmount)
				tx.ChangeCmdArg(coin.String(), "invalid_amount"+ccDenom)
				tx.CheckFailedWithErrorSubstring("parsing coin")
			}
			// MsgWithdrawCurrency ValidateBasic
			{
				tx := ct.TxCurrenciesWithdraw(ccRecipient, ccRecipient, ccDenom, sdk.ZeroInt())
				tx.CheckFailedWithErrorSubstring("wrong amount")
			}
		}
	}

	// check issue Query
	{
		// incorrect inputs
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
				q.CheckFailedWithSDKError(currencies.ErrWrongIssueID)
			}
		}
	}

	// check currency Query
	{
		// incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryCurrenciesCurrency(ccDenom)
				q.RemoveCmdArg(ccDenom)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
		}
	}

	// check withdraw Query
	{
		// incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryCurrenciesCurrency(ccDenom)
				q.RemoveCmdArg(ccDenom)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing withdrawID
			{
				q, _ := ct.QueryCurrenciesWithdraw(dnTypes.NewIDFromUint64(1))
				q.ChangeCmdArg("1", "non_int")
				q.CheckFailedWithErrorSubstring("")
			}
		}
	}

	// check withdraws Query
	{
		q, withdraws := ct.QueryCurrenciesWithdraws(1, 10)
		q.CheckSucceeded()
		require.Len(t, *withdraws, 1)

		// incorrect inputs
		{
			// page / limit
			{
				q, _ := ct.QueryCurrenciesWithdraws(1, 10)
				q.ChangeCmdArg("--page=1", "--page=-1")
				q.CheckFailedWithErrorSubstring("")

				q, _ = ct.QueryCurrenciesWithdraws(1, 10)
				q.ChangeCmdArg("--limit=10", "--limit=abc")
				q.CheckFailedWithErrorSubstring("")
			}
		}
	}
}

// Check that distribution commands in CLI disabled.
func TestDisableRewards_CLI(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	code, stdOut, _ := ct.TxDistributionWithoutParams().Send()

	require.Equal(t, 0, code)
	require.NotContains(t, string(stdOut), "distribution")
}

func TestOracle_CLI(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	wsStop, wsChs := ct.CheckWSsSubscribed(false, "TestOracle_CLI", []string{"message.module='oracle'"}, 10)
	defer wsStop()
	go cliTester.PrintEvents(t, wsChs, "oracle")

	nomineeAddr := ct.Accounts["nominee"].Address
	assetCode := dnTypes.AssetCode("eth_xfi")
	assetOracle1, assetOracle2, assetOracle3 := ct.Accounts["oracle1"].Address, ct.Accounts["oracle2"].Address, ct.Accounts["oracle3"].Address

	// check add asset Tx
	{
		ct.TxOracleAddAsset(nomineeAddr, assetCode, assetOracle1).CheckSucceeded()

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 2)
		asset := (*assets)[1]
		require.Equal(t, ct.DefAssetCode, (*assets)[0].AssetCode)
		require.Equal(t, assetCode, asset.AssetCode)
		require.Len(t, asset.Oracles, 1)
		require.Equal(t, assetOracle1, asset.Oracles[0].Address.String())
		require.True(t, asset.Active)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOracleAddAsset(nomineeAddr, assetCode, assetOracle1)
				tx.RemoveCmdArg(assetCode.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid assetCode
			{
				tx := ct.TxOracleAddAsset(nomineeAddr, "wrongasset", assetOracle1)
				tx.CheckFailedWithErrorSubstring("assetCode argument")
			}
			// invalid oracles
			{
				tx := ct.TxOracleAddAsset(nomineeAddr, assetCode, "123")
				tx.CheckFailedWithErrorSubstring("")
			}
			// empty assetCode
			{
				tx := ct.TxOracleAddAsset(nomineeAddr, "", assetOracle1)
				tx.CheckFailedWithErrorSubstring("assetCode argument")
			}
			// empty oracles
			{
				tx := ct.TxOracleAddAsset(nomineeAddr, assetCode)
				tx.CheckFailedWithErrorSubstring("oracleAddresses argument")
			}
		}
	}

	// check set asset Tx
	{
		ct.TxOracleSetAsset(nomineeAddr, assetCode, assetOracle1, assetOracle2).CheckSucceeded()

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 2)
		asset := (*assets)[1]
		require.Equal(t, assetCode, asset.AssetCode)
		require.Len(t, asset.Oracles, 2)
		require.Equal(t, assetOracle1, asset.Oracles[0].Address.String())
		require.Equal(t, assetOracle2, asset.Oracles[1].Address.String())
		require.True(t, asset.Active)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOracleSetAsset(nomineeAddr, assetCode, assetOracle1)
				tx.RemoveCmdArg(assetCode.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid assetCode
			{
				tx := ct.TxOracleSetAsset(nomineeAddr, "wrongasset", assetOracle1)
				tx.CheckFailedWithErrorSubstring("assetCode argument")
			}
			// invalid oracles
			{
				tx := ct.TxOracleSetAsset(nomineeAddr, assetCode, "123")
				tx.CheckFailedWithErrorSubstring("")
			}
			// empty assetCode
			{
				tx := ct.TxOracleSetAsset(nomineeAddr, "", assetOracle1)
				tx.CheckFailedWithErrorSubstring("assetCode argument")
			}
			// empty oracles
			{
				tx := ct.TxOracleSetAsset(nomineeAddr, assetCode)
				tx.CheckFailedWithErrorSubstring("oracleAddresses argument")
			}
		}
	}

	// check post price Tx
	{
		now := time.Now().Truncate(1 * time.Second).UTC()
		postPrices := []struct {
			assetCode  dnTypes.AssetCode
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
				tx.RemoveCmdArg(assetCode.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid price
			{
				tx := ct.TxOraclePostPrice(assetOracle1, assetCode, sdk.OneInt(), time.Now())
				tx.ChangeCmdArg(sdk.OneInt().String(), "not_int")
				tx.CheckFailedWithErrorSubstring("parsing Int")
			}
			// invalid receivedAt
			{
				now := time.Now()
				tx := ct.TxOraclePostPrice(assetOracle1, assetCode, sdk.OneInt(), now)
				tx.ChangeCmdArg(strconv.FormatInt(now.Unix(), 10), "not_time.Time")
				tx.CheckFailedWithErrorSubstring("parsing Int")
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

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 2)
		asset := (*assets)[1]
		require.Equal(t, assetCode, asset.AssetCode)
		require.Len(t, asset.Oracles, 3)
		require.Equal(t, assetOracle1, asset.Oracles[0].Address.String())
		require.Equal(t, assetOracle2, asset.Oracles[1].Address.String())
		require.Equal(t, assetOracle3, asset.Oracles[2].Address.String())
		require.True(t, asset.Active)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOracleAddOracle(nomineeAddr, assetCode, "invalid_address")
				tx.RemoveCmdArg(assetCode.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid oracleAddress
			{
				tx := ct.TxOracleAddOracle(nomineeAddr, assetCode, "invalid_address")
				tx.CheckFailedWithErrorSubstring("oracleAddress argument")
			}
		}
	}

	// check set oracle Tx
	{
		ct.TxOracleSetOracles(nomineeAddr, assetCode, assetOracle3, assetOracle2, assetOracle1).CheckSucceeded()

		q, assets := ct.QueryOracleAssets()
		q.CheckSucceeded()
		require.Len(t, *assets, 2)
		asset := (*assets)[1]
		require.Equal(t, assetCode, asset.AssetCode)
		require.Len(t, asset.Oracles, 3)
		require.Equal(t, assetOracle3, asset.Oracles[0].Address.String())
		require.Equal(t, assetOracle2, asset.Oracles[1].Address.String())
		require.Equal(t, assetOracle1, asset.Oracles[2].Address.String())
		require.True(t, asset.Active)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxOracleSetOracles(nomineeAddr, assetCode, assetOracle3, assetOracle2, assetOracle1)
				tx.RemoveCmdArg(assetCode.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid oracles
			{
				tx := ct.TxOracleSetOracles(nomineeAddr, assetCode, "123")
				tx.CheckFailedWithErrorSubstring("oracleAddresses argument")
			}
		}
	}

	// check rawPrices query with invalid arguments
	{
		// invalid number of args
		{
			q, _ := ct.QueryOracleRawPrices(assetCode, 1)
			q.RemoveCmdArg(assetCode.String())
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
				q.RemoveCmdArg(assetCode.String())
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing assetCode
			{
				q, _ := ct.QueryOraclePrice("nonexisting_asset")
				q.CheckFailedWithSDKError(sdkErrors.ErrUnknownRequest)
			}
		}
	}
}

func TestPOA_CLI(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	wsStop, wsChs := ct.CheckWSsSubscribed(false, "TestPOA_CLI", []string{"message.module='poa'"}, 10)
	defer wsStop()
	go cliTester.PrintEvents(t, wsChs, "poa")

	curValidators := make([]poa.Validator, 0)
	addValidator := func(address, ethAddress string) {
		sdkAddr, err := sdk.AccAddressFromBech32(address)
		require.NoError(t, err, "converting account address")
		curValidators = append(curValidators, poa.Validator{
			Address:    sdkAddr,
			EthAddress: ethAddress,
		})
	}
	removeValidator := func(address string) {
		idx := -1
		for i, v := range curValidators {
			if v.Address.String() == address {
				idx = i
			}
		}

		require.GreaterOrEqual(t, idx, 0, "not found")
		curValidators = append(curValidators[:idx], curValidators[idx+1:]...)
	}

	for _, acc := range ct.Accounts {
		if acc.IsPOAValidator {
			addValidator(acc.Address, acc.EthAddress)
		}
	}

	senderAddr := ct.Accounts["validator1"].Address
	newValidatorAccName := "plain"
	newValidatorAcc := ct.Accounts[newValidatorAccName]
	nonExistingAddress := secp256k1.GenPrivKey().PubKey().Address()

	// check add validator Tx
	{
		require.LessOrEqual(t, len(curValidators), len(cliTester.EthAddresses), "not enough predefined ethAddresses")
		newEthAddress, issueID := cliTester.EthAddresses[len(curValidators)], "newValidator"

		ct.TxPoaAddValidator(senderAddr, newValidatorAcc.Address, newEthAddress, issueID).CheckSucceeded()
		ct.ConfirmCall(issueID)

		// update account
		addValidator(newValidatorAcc.Address, newEthAddress)
		ct.Accounts[newValidatorAccName].IsPOAValidator = true

		// check validator added
		newValidator := curValidators[len(curValidators)-1]
		q, validators := ct.QueryPoaValidators()
		q.CheckSucceeded()
		q, rcvV := ct.QueryPoaValidator(newValidator.Address.String())
		q.CheckSucceeded()

		require.Len(t, (*validators).Validators, len(curValidators))
		require.True(t, newValidator.Address.Equals(rcvV.Address))
		require.Equal(t, newValidator.EthAddress, rcvV.EthAddress)

		// check incorrect inputs
		{
			// wrong number of args
			{
				tx := ct.TxPoaAddValidator(senderAddr, newValidatorAcc.Address, newEthAddress, issueID)
				tx.RemoveCmdArg(issueID)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing fromAddress
			{
				tx := ct.TxPoaAddValidator(nonExistingAddress.String(), newValidatorAcc.Address, newEthAddress, issueID)
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid validator address
			{
				tx := ct.TxPoaAddValidator(senderAddr, "invalid_address", newEthAddress, issueID)
				tx.CheckFailedWithErrorSubstring("address")
			}
			// MsgAddValidator ValidateBasic
			{
				tx := ct.TxPoaAddValidator(senderAddr, newValidatorAcc.Address, "invalid_eth_address", issueID)
				tx.CheckFailedWithErrorSubstring("invalid_eth_address")
			}
		}
	}

	// check remove validator Tx
	{
		issueID := "rmValidator"
		ct.TxPoaRemoveValidator(senderAddr, newValidatorAcc.Address, issueID).CheckSucceeded()
		ct.ConfirmCall(issueID)

		// update account
		removeValidator(newValidatorAcc.Address)
		ct.Accounts[newValidatorAccName].IsPOAValidator = false

		// check validator removed
		qValidators, validators := ct.QueryPoaValidators()
		qValidators.CheckSucceeded()
		qValidator, _ := ct.QueryPoaValidator(newValidatorAcc.Address)
		qValidator.CheckFailedWithErrorSubstring("not found")

		require.Len(t, (*validators).Validators, len(curValidators))

		// check incorrect inputs
		{
			// wrong number of args
			{
				tx := ct.TxPoaRemoveValidator(senderAddr, newValidatorAcc.Address, issueID)
				tx.RemoveCmdArg(issueID)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing fromAddress
			{
				tx := ct.TxPoaRemoveValidator(nonExistingAddress.String(), newValidatorAcc.Address, issueID)
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid validator address
			{
				tx := ct.TxPoaRemoveValidator(senderAddr, "invalid_address", issueID)
				tx.CheckFailedWithErrorSubstring("address")
			}
		}
	}

	// check replace validator Tx
	{
		require.LessOrEqual(t, len(curValidators), len(cliTester.EthAddresses), "not enough predefined ethAddresses")
		newEthAddress := cliTester.EthAddresses[len(curValidators)]

		targetValidatorName := "validator2"
		targetValidatorAcc := ct.Accounts[targetValidatorName]
		issueID := "ReplaceValidator"

		tx := ct.TxPoaReplaceValidator(senderAddr, targetValidatorAcc.Address, newValidatorAcc.Address, newEthAddress, issueID)
		tx.CheckSucceeded()
		ct.ConfirmCall(issueID)

		// update accounts
		removeValidator(targetValidatorAcc.Address)
		addValidator(newValidatorAcc.Address, newEthAddress)
		ct.Accounts[targetValidatorName].IsPOAValidator = false
		ct.Accounts[newValidatorAccName].IsPOAValidator = true

		// check validator replaced
		newValidator := curValidators[len(curValidators)-1]
		q, validators := ct.QueryPoaValidators()
		q.CheckSucceeded()
		q, rcvV := ct.QueryPoaValidator(newValidatorAcc.Address)
		q.CheckSucceeded()

		require.Len(t, (*validators).Validators, len(curValidators))
		require.True(t, newValidator.Address.Equals(rcvV.Address))
		require.Equal(t, newValidator.EthAddress, rcvV.EthAddress)

		// check incorrect inputs
		{
			// wrong number of args
			{
				tx := ct.TxPoaReplaceValidator(senderAddr, targetValidatorAcc.Address, newValidatorAcc.Address, newEthAddress, issueID)
				tx.RemoveCmdArg(issueID)
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing fromAddress
			{
				tx := ct.TxPoaReplaceValidator(nonExistingAddress.String(), targetValidatorAcc.Address, newValidatorAcc.Address, newEthAddress, issueID)
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid old validator address
			{
				tx := ct.TxPoaReplaceValidator(senderAddr, "invalid_address", newValidatorAcc.Address, newEthAddress, issueID)
				tx.CheckFailedWithErrorSubstring("oldValidator")
			}
			// invalid new validator address
			{
				tx := ct.TxPoaReplaceValidator(senderAddr, targetValidatorAcc.Address, "invalid_address", newEthAddress, issueID)
				tx.CheckFailedWithErrorSubstring("newValidator")
			}
			// invalid ethAddress
			{
				tx := ct.TxPoaReplaceValidator(senderAddr, targetValidatorAcc.Address, newValidatorAcc.Address, "invalid_eth_address", issueID)
				tx.CheckFailedWithErrorSubstring("invalid_eth_address")
			}
		}
	}

	// check validators query
	{
		q, validators := ct.QueryPoaValidators()
		q.CheckSucceeded()

		require.Equal(t, len(curValidators)/2+1, int(validators.Confirmations))
		require.Len(t, validators.Validators, len(curValidators))
		for _, rcvV := range validators.Validators {
			found := false
			for _, curV := range curValidators {
				if rcvV.EthAddress == curV.EthAddress && curV.Address.Equals(rcvV.Address) {
					found = true
				}
			}
			require.True(t, found, "validator %s: not found", rcvV.String())
		}
	}

	// check minMax params query
	{
		q, params := ct.QueryPoaMinMax()
		q.CheckSucceeded()

		poaGenesis := poa.GenesisState{}
		require.NoError(t, ct.Cdc.UnmarshalJSON(ct.GenesisState()[poa.ModuleName], &poaGenesis))
		require.Equal(t, poaGenesis.Parameters.MaxValidators, params.MaxValidators)
		require.Equal(t, poaGenesis.Parameters.MinValidators, params.MinValidators)
	}

	// check validator query
	{
		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryPoaValidator(curValidators[0].Address.String())
				q.RemoveCmdArg(curValidators[0].Address.String())
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid address
			{
				q, _ := ct.QueryPoaValidator("invalid_address")
				q.CheckFailedWithErrorSubstring("address")
			}
			// non-existing validator
			{
				addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
				q, _ := ct.QueryPoaValidator(addr.String())
				q.CheckFailedWithErrorSubstring("not found")
			}
		}
	}
}

func TestMS_CLI(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	wsStop, wsChs := ct.CheckWSsSubscribed(false, "TestMS_CLI", []string{"message.module='multisig'"}, 10)
	defer wsStop()
	go cliTester.PrintEvents(t, wsChs, "multisig")

	ccDenom1, ccDenom2 := "btc", "eth"
	ccCurAmount := sdk.NewInt(1000)
	callUniqueId1, callUniqueId2 := "issue1", "issue2"
	nonExistingAddress := secp256k1.GenPrivKey().PubKey().Address()

	// get all validators
	ccRecipients := make([]string, 0)
	for _, acc := range ct.Accounts {
		if acc.IsPOAValidator {
			ccRecipients = append(ccRecipients, acc.Address)
		}
	}

	// create calls
	ct.TxCurrenciesIssue(ccRecipients[0], ccRecipients[0], callUniqueId1, ccDenom1, ccCurAmount).CheckSucceeded()
	ct.TxCurrenciesIssue(ccRecipients[1], ccRecipients[1], callUniqueId2, ccDenom2, ccCurAmount).CheckSucceeded()

	checkCall := func(call multisig.CallResp, approved bool, callID dnTypes.ID, uniqueID, creatorAddr string, votesAddr ...string) {
		require.Len(t, call.Votes, len(votesAddr))
		for i := range call.Votes {
			require.Equal(t, call.Votes[i].String(), votesAddr[i])
		}
		require.Equal(t, approved, call.Call.Approved)
		require.Equal(t, approved, call.Call.Executed)
		require.False(t, call.Call.Failed)
		require.False(t, call.Call.Rejected)
		require.GreaterOrEqual(t, call.Call.Height, int64(0))
		require.Empty(t, call.Call.Error)
		require.Equal(t, creatorAddr, call.Call.Creator.String())
		require.NotNil(t, call.Call.Msg)
		require.Equal(t, callID.String(), call.Call.ID.String())
		require.Equal(t, uniqueID, call.Call.UniqueID)
		require.NotEmpty(t, call.Call.MsgRoute)
		require.NotEmpty(t, call.Call.MsgType)
	}

	// check calls query
	{
		q, calls := ct.QueryMultiSigCalls()
		q.CheckSucceeded()

		require.Len(t, *calls, 2)
		checkCall((*calls)[0], false, dnTypes.NewIDFromUint64(0), callUniqueId1, ccRecipients[0], ccRecipients[0])
		checkCall((*calls)[1], false, dnTypes.NewIDFromUint64(1), callUniqueId2, ccRecipients[1], ccRecipients[1])
	}

	// check call query
	{
		q, call := ct.QueryMultiSigCall(dnTypes.NewIDFromUint64(0))
		q.CheckSucceeded()

		checkCall(*call, false, dnTypes.NewIDFromUint64(0), callUniqueId1, ccRecipients[0], ccRecipients[0])

		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryMultiSigCall(dnTypes.NewIDFromUint64(0))
				q.RemoveCmdArg("0")
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid callID
			{
				q, _ := ct.QueryMultiSigCall(dnTypes.NewIDFromUint64(0))
				q.ChangeCmdArg("0", "abc")
				q.CheckFailedWithErrorSubstring("abc")
			}
			// non-existing callID
			{
				q, _ := ct.QueryMultiSigCall(dnTypes.NewIDFromUint64(2))
				q.CheckFailedWithSDKError(multisig.ErrWrongCallId)
			}
		}
	}

	// check uniqueCall query
	{
		q, call := ct.QueryMultiSigUnique(callUniqueId1)
		q.CheckSucceeded()

		checkCall(*call, false, dnTypes.NewIDFromUint64(0), callUniqueId1, ccRecipients[0], ccRecipients[0])

		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryMultiSigUnique(callUniqueId1)
				q.RemoveCmdArg(callUniqueId1)
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing uniqueID
			{
				q, _ := ct.QueryMultiSigUnique("non_existing_uniqueID")
				q.CheckFailedWithSDKError(multisig.ErrWrongCallUniqueId)
			}
		}
	}

	// check lastId query
	{
		q, lastId := ct.QueryMultiLastId()
		q.CheckSucceeded()

		require.EqualValues(t, 1, lastId.LastID.UInt64())
	}

	// check confirm call Tx
	{
		// add votes for existing call from other senders
		callID, callUniqueID := dnTypes.NewIDFromUint64(0), callUniqueId1
		votes := []string{ccRecipients[0]}
		for i := 1; i < len(ccRecipients)/2+1; i++ {
			ct.TxMultiSigConfirmCall(ccRecipients[i], callID).CheckSucceeded()
			votes = append(votes, ccRecipients[i])
		}

		// check call approved
		q, call := ct.QueryMultiSigCall(callID)
		q.CheckSucceeded()

		checkCall(*call, true, callID, callUniqueID, ccRecipients[0], votes...)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxMultiSigConfirmCall(ccRecipients[0], callID)
				tx.RemoveCmdArg(callID.String())
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing fromAddress
			{
				tx := ct.TxMultiSigConfirmCall(nonExistingAddress.String(), callID)
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid callID
			{
				tx := ct.TxMultiSigConfirmCall(ccRecipients[0], callID)
				tx.ChangeCmdArg(callID.String(), "not_int")
				tx.CheckFailedWithErrorSubstring("not_int")
			}
		}
	}

	// check revoke confirm Tx
	{
		ct.TxMultiSigRevokeConfirm(ccRecipients[1], dnTypes.NewIDFromUint64(1)).CheckSucceeded()

		// check call removed
		q, resp := ct.QueryMultiSigCall(dnTypes.NewIDFromUint64(1))
		q.CheckSucceeded()

		require.Len(t, resp.Votes, 0)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxMultiSigRevokeConfirm(ccRecipients[0], dnTypes.NewIDFromUint64(0))
				tx.RemoveCmdArg(strconv.FormatUint(0, 10))
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing fromAddress
			{
				tx := ct.TxMultiSigRevokeConfirm(nonExistingAddress.String(), dnTypes.NewIDFromUint64(0))
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid callID
			{
				tx := ct.TxMultiSigRevokeConfirm(ccRecipients[0], dnTypes.NewIDFromUint64(0))
				tx.ChangeCmdArg(strconv.FormatUint(0, 10), "not_int")
				tx.CheckFailedWithErrorSubstring("not_int")
			}
		}
	}
}

func TestMarkets_CLI(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	ownerAddr := ct.Accounts["validator1"].Address

	// add markets
	ct.TxMarketsAdd(ownerAddr, cliTester.DenomBTC, cliTester.DenomXFI).CheckSucceeded()
	ct.TxMarketsAdd(ownerAddr, cliTester.DenomETH, cliTester.DenomXFI).CheckSucceeded()

	// check addMarket Tx
	{
		// invalid owner
		{
			tx := ct.TxMarketsAdd("invalid_address", cliTester.DenomBTC, "atom")
			tx.CheckFailedWithErrorSubstring("keyring")
		}

		// non-existing currency
		{
			tx := ct.TxMarketsAdd(ownerAddr, cliTester.DenomBTC, "atom")
			tx.CheckFailedWithSDKError(markets.ErrWrongAssetDenom)
		}

		// already existing market
		{
			tx := ct.TxMarketsAdd(ownerAddr, cliTester.DenomBTC, cliTester.DenomXFI)
			tx.CheckFailedWithSDKError(markets.ErrMarketExists)
		}
	}

	// check market query
	{
		// non-existing marketID
		{
			q, _ := ct.QueryMarketsMarket(dnTypes.NewIDFromUint64(10))
			q.CheckFailedWithErrorSubstring("wrong ID")
		}

		// existing marketID (btc-xfi)
		{
			q, market := ct.QueryMarketsMarket(dnTypes.NewIDFromUint64(0))
			q.CheckSucceeded()

			require.Equal(t, market.ID.UInt64(), uint64(0))
			require.Equal(t, market.BaseAssetDenom, cliTester.DenomBTC)
			require.Equal(t, market.QuoteAssetDenom, cliTester.DenomXFI)
		}
	}

	// check list query
	{
		// all markets
		{
			q, markets := ct.QueryMarketsList(1, 10, nil, nil)
			q.CheckSucceeded()

			require.Len(t, *markets, 2)
			require.Equal(t, (*markets)[0].ID.UInt64(), uint64(0))
			require.Equal(t, (*markets)[0].BaseAssetDenom, cliTester.DenomBTC)
			require.Equal(t, (*markets)[1].ID.UInt64(), uint64(1))
			require.Equal(t, (*markets)[1].BaseAssetDenom, cliTester.DenomETH)
		}

		// check page / limit parameters
		{
			// page 1, limit 1
			qP1L1, marketsP1L1 := ct.QueryMarketsList(1, 1, nil, nil)
			qP1L1.CheckSucceeded()

			require.Len(t, *marketsP1L1, 1)

			// page 2, limit 1
			qP2L1, marketsP2L1 := ct.QueryMarketsList(1, 1, nil, nil)
			qP2L1.CheckSucceeded()

			require.Len(t, *marketsP2L1, 1)

			// page 2, limit 10 (no markets)
			qP2L10, marketsP2L10 := ct.QueryMarketsList(2, 10, nil, nil)
			qP2L10.CheckSucceeded()

			require.Empty(t, *marketsP2L10)
		}

		// check baseDenom filter
		{
			baseDenom := cliTester.DenomETH
			q, markets := ct.QueryMarketsList(-1, -1, &baseDenom, nil)
			q.CheckSucceeded()

			require.Len(t, *markets, 1)
			require.Equal(t, (*markets)[0].BaseAssetDenom, baseDenom)
		}

		// check quoteDenom filter
		{
			quoteDenom := cliTester.DenomXFI
			q, markets := ct.QueryMarketsList(-1, -1, nil, &quoteDenom)
			q.CheckSucceeded()

			require.Len(t, *markets, 2)
			require.Equal(t, (*markets)[0].QuoteAssetDenom, quoteDenom)
			require.Equal(t, (*markets)[1].QuoteAssetDenom, quoteDenom)
		}

		// check multiple filters
		{
			baseDeno := cliTester.DenomBTC
			quoteDenom := cliTester.DenomXFI
			q, markets := ct.QueryMarketsList(-1, -1, &baseDeno, &quoteDenom)
			q.CheckSucceeded()

			require.Len(t, *markets, 1)
		}
	}
}

func TestOrders_CLI(t *testing.T) {
	t.Parallel()

	const (
		DecimalsXFI = "1000000000000000000"
		DecimalsETH = "1000000000000000000"
		DecimalsBTC = "100000000"
	)

	oneXfi := sdk.NewUintFromString(DecimalsXFI)
	oneBtc := sdk.NewUintFromString(DecimalsBTC)
	oneEth := sdk.NewUintFromString(DecimalsETH)
	accountBalances := []cliTester.StringPair{
		{
			Key:   cliTester.DenomBTC,
			Value: sdk.NewUint(10000).Mul(oneBtc).String(),
		},
		{
			Key:   cliTester.DenomETH,
			Value: sdk.NewUint(100000000).Mul(oneEth).String(),
		},
		{
			Key:   cliTester.DenomXFI,
			Value: sdk.NewUint(100000000).Mul(oneXfi).String(),
		},
	}
	accountOpts := []cliTester.AccountOption{
		{Name: "client1", Balances: accountBalances},
		{Name: "client2", Balances: accountBalances},
	}

	ct := cliTester.New(
		t,
		false,
		cliTester.AccountsOption(accountOpts...),
	)
	defer ct.Close()

	ownerAddr1 := ct.Accounts[accountOpts[0].Name].Address
	ownerAddr2 := ct.Accounts[accountOpts[1].Name].Address
	marketID0, marketID1 := dnTypes.NewIDFromUint64(0), dnTypes.NewIDFromUint64(1)
	assetCode0, assetCode1 := dnTypes.AssetCode("btc_xfi"), dnTypes.AssetCode("eth_xfi")

	wsStop, wsChs := ct.CheckWSsSubscribed(false, "TestOrders_CLI", []string{"message.module='orders'"}, 10)
	defer wsStop()
	go cliTester.PrintEvents(t, wsChs, "orders")

	// add market
	ct.TxMarketsAdd(ownerAddr1, cliTester.DenomBTC, cliTester.DenomXFI).CheckSucceeded()
	ct.TxMarketsAdd(ownerAddr1, cliTester.DenomETH, cliTester.DenomXFI).CheckSucceeded()

	// check AddOrder Tx
	{
		// invalid owner
		{
			tx := ct.TxOrdersPost("invalid_address", assetCode0, orders.AskDirection, sdk.OneUint(), sdk.OneUint(), 60)
			tx.CheckFailedWithErrorSubstring("keyring")
		}

		// invalid marketID
		{
			tx := ct.TxOrdersPost(ownerAddr1, dnTypes.AssetCode("wrong_code"), orders.AskDirection, sdk.OneUint(), sdk.OneUint(), 60)
			tx.CheckFailedWithSDKError(orders.ErrWrongAssetCode)
		}

		// invalid direction
		{
			tx := ct.TxOrdersPost(ownerAddr1, assetCode0, orders.AskDirection, sdk.OneUint(), sdk.OneUint(), 60)
			tx.ChangeCmdArg("ask", "invalid")
			tx.CheckFailedWithErrorSubstring("direction")
		}

		// invalid price
		{
			tx := ct.TxOrdersPost(ownerAddr1, assetCode0, orders.AskDirection, sdk.ZeroUint(), sdk.OneUint(), 60)
			tx.ChangeCmdArg("0", "invalid")
			tx.CheckFailedWithErrorSubstring("price")
		}

		// invalid quantity
		{
			tx := ct.TxOrdersPost(ownerAddr1, assetCode0, orders.AskDirection, sdk.OneUint(), sdk.ZeroUint(), 60)
			tx.ChangeCmdArg("0", "invalid")
			tx.CheckFailedWithErrorSubstring("quantity")
		}
	}

	// add orders
	inputOrders := []struct {
		MarketID     dnTypes.ID
		AssetCode    dnTypes.AssetCode
		OwnerAddress string
		Direction    orders.Direction
		Price        sdk.Uint
		Quantity     sdk.Uint
		TtlInSec     int
	}{
		{
			MarketID:     marketID0,
			AssetCode:    assetCode0,
			OwnerAddress: ownerAddr1,
			Direction:    orders.BidDirection,
			Price:        sdk.NewUintFromString("10000000000000000000"),
			Quantity:     sdk.NewUintFromString("100000000"),
			TtlInSec:     60,
		},
		{
			MarketID:     marketID0,
			AssetCode:    assetCode0,
			OwnerAddress: ownerAddr2,
			Direction:    orders.BidDirection,
			Price:        sdk.NewUintFromString("20000000000000000000"),
			Quantity:     sdk.NewUintFromString("200000000"),
			TtlInSec:     90,
		},
		{
			MarketID:     marketID0,
			AssetCode:    assetCode0,
			OwnerAddress: ownerAddr1,
			Direction:    orders.AskDirection,
			Price:        sdk.NewUintFromString("50000000000000000000"),
			Quantity:     sdk.NewUintFromString("500000000"),
			TtlInSec:     60,
		},
		{
			MarketID:     marketID0,
			AssetCode:    assetCode0,
			OwnerAddress: ownerAddr2,
			Direction:    orders.AskDirection,
			Price:        sdk.NewUintFromString("60000000000000000000"),
			Quantity:     sdk.NewUintFromString("600000000"),
			TtlInSec:     90,
		},
		{
			MarketID:     marketID1,
			AssetCode:    assetCode1,
			OwnerAddress: ownerAddr1,
			Direction:    orders.AskDirection,
			Price:        sdk.NewUintFromString("10000000000000000000"),
			Quantity:     sdk.NewUintFromString("100000000"),
			TtlInSec:     30,
		},
	}
	for _, input := range inputOrders {
		ct.TxOrdersPost(input.OwnerAddress, input.AssetCode, input.Direction, input.Price, input.Quantity, input.TtlInSec).CheckSucceeded()
	}

	// check orders added
	{
		for i, input := range inputOrders {
			orderID := dnTypes.NewIDFromUint64(uint64(i))
			q, order := ct.QueryOrdersOrder(orderID)
			q.CheckSucceeded()

			require.True(t, order.ID.Equal(orderID), "order %d: ID", i)
			require.True(t, order.Market.ID.Equal(input.MarketID), "order %d: MarketID", i)
			require.Equal(t, order.Owner.String(), input.OwnerAddress, "order %d: Owner", i)
			require.True(t, order.Direction.Equal(input.Direction), "order %d: Direction", i)
			require.True(t, order.Price.Equal(input.Price), "order %d: Price", i)
			require.True(t, order.Quantity.Equal(input.Quantity), "order %d: Quantity", i)
			require.Equal(t, order.Ttl, time.Duration(input.TtlInSec)*time.Second, "order %d: Ttl", i)
		}
	}

	// check list query
	{
		// request all
		{
			q, orders := ct.QueryOrdersList(-1, -1, nil, nil, nil)
			q.CheckSucceeded()

			require.Len(t, *orders, len(inputOrders))
		}

		// check page / limit parameters
		{
			// page 1, limit 1
			qP1L1, ordersP1L1 := ct.QueryOrdersList(1, 1, nil, nil, nil)
			qP1L1.CheckSucceeded()

			require.Len(t, *ordersP1L1, 1)

			// page 2, limit 1
			qP2L1, ordersP2L1 := ct.QueryOrdersList(1, 1, nil, nil, nil)
			qP2L1.CheckSucceeded()

			require.Len(t, *ordersP2L1, 1)

			// page 2, limit 10 (no orders)
			qP2L10, ordersP2L10 := ct.QueryOrdersList(2, 10, nil, nil, nil)
			qP2L10.CheckSucceeded()

			require.Empty(t, *ordersP2L10)
		}

		// check marketID filter
		{
			market0Count, market1Count := 0, 0
			for _, input := range inputOrders {
				if input.MarketID.UInt64() == 0 {
					market0Count++
				}
				if input.MarketID.UInt64() == 1 {
					market1Count++
				}
			}

			q0, orders0 := ct.QueryOrdersList(-1, -1, &marketID0, nil, nil)
			q0.CheckSucceeded()

			require.Len(t, *orders0, market0Count)

			q1, orders1 := ct.QueryOrdersList(-1, -1, &marketID1, nil, nil)
			q1.CheckSucceeded()

			require.Len(t, *orders1, market1Count)
		}

		// check direction filter
		{
			askCount, bidCount := 0, 0
			for _, input := range inputOrders {
				if input.Direction.Equal(orders.AskDirection) {
					askCount++
				}
				if input.Direction.Equal(orders.BidDirection) {
					bidCount++
				}
			}

			askDirection := orders.AskDirection
			qAsk, ordersAsk := ct.QueryOrdersList(-1, -1, nil, &askDirection, nil)
			qAsk.CheckSucceeded()

			require.Len(t, *ordersAsk, askCount)

			bidDirection := orders.BidDirection
			qBid, ordersBid := ct.QueryOrdersList(-1, -1, nil, &bidDirection, nil)
			qBid.CheckSucceeded()

			require.Len(t, *ordersBid, bidCount)
		}

		// check owner filter
		{
			client1Count, client2Count := 0, 0
			for _, input := range inputOrders {
				if input.OwnerAddress == ownerAddr1 {
					client1Count++
				}
				if input.OwnerAddress == ownerAddr2 {
					client2Count++
				}
			}

			q1, orders1 := ct.QueryOrdersList(-1, -1, nil, nil, &ownerAddr1)
			q1.CheckSucceeded()

			require.Len(t, *orders1, client1Count)

			q2, orders2 := ct.QueryOrdersList(-1, -1, nil, nil, &ownerAddr2)
			q2.CheckSucceeded()

			require.Len(t, *orders2, client2Count)
		}

		// check multiple filters
		{
			marketID := marketID0
			owner := ownerAddr1
			direction := orders.AskDirection
			count := 0
			for _, input := range inputOrders {
				if input.MarketID.Equal(marketID) && input.OwnerAddress == owner && input.Direction == direction {
					count++
				}
			}

			q, orders := ct.QueryOrdersList(-1, -1, &marketID, &direction, &owner)
			q.CheckSucceeded()

			require.Len(t, *orders, count)
		}
	}

	// revoke order
	{
		orderIdx := len(inputOrders) - 1
		orderID := dnTypes.NewIDFromUint64(uint64(orderIdx))
		inputOrder := inputOrders[orderIdx]
		ct.TxOrdersRevoke(inputOrder.OwnerAddress, orderID).CheckSucceeded()

		q, _ := ct.QueryOrdersOrder(orderID)
		q.CheckFailedWithSDKError(orders.ErrWrongOrderID)
		inputOrders = inputOrders[:len(inputOrders)-2]
	}

	// check RevokeOrder Tx
	{
		// invalid from address
		{
			tx := ct.TxOrdersRevoke("invalid_address", dnTypes.NewIDFromUint64(0))
			tx.CheckFailedWithErrorSubstring("keyring")
		}

		// non-existing orderID
		{
			tx := ct.TxOrdersRevoke(ownerAddr1, dnTypes.NewIDFromUint64(10))
			tx.CheckFailedWithSDKError(orders.ErrWrongOrderID)
		}

		// wrong owner (not an order owner)
		{
			tx := ct.TxOrdersRevoke(ct.Accounts["validator1"].Address, dnTypes.NewIDFromUint64(0))
			tx.CheckFailedWithSDKError(orders.ErrWrongOwner)
		}
	}
}

func Test_RestServer(t *testing.T) {
	t.Parallel()

	ct := cliTester.New(t, false)
	defer ct.Close()

	restUrl := ct.StartRestServer(false)

	// check server is running
	{
		resp, err := http.Get(restUrl + "/blocks/latest")
		require.NoError(t, err, "Get request")
		require.NotNil(t, resp, "response")
		require.NotNil(t, resp.Body, "response body")
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err, "body read")

		resultBlock := tmCoreTypes.ResultBlock{}
		require.NoError(t, ct.Cdc.UnmarshalJSON(body, &resultBlock), "body unmarshal")

		require.NotNil(t, resultBlock.Block, "result block")
		require.Equal(t, ct.IDs.ChainID, resultBlock.Block.ChainID)
		require.GreaterOrEqual(t, resultBlock.Block.Height, int64(1))
	}
}
