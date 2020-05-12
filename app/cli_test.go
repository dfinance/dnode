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
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

const (
	NotFoundErrSubString = "The specified item could not be found in the keyring"
)

func Test_CurrencyCLI(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()

	ccSymbol, ccCurAmount, ccDecimals, ccRecipient := "btc", sdk.NewInt(1000), int8(1), ct.Accounts["validator1"].Address
	nonExistingAddress := secp256k1.GenPrivKey().PubKey().Address()
	issueID := "issue1"

	// check issue currency multisig Tx
	{
		// submit & confirm call
		ct.TxCurrenciesIssue(ccRecipient, ccRecipient, ccSymbol, ccCurAmount, ccDecimals, issueID).CheckSucceeded()
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
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
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
		ccCurAmount = ccCurAmount.Sub(destroyAmount)
		// check destroy
		q, destroy := ct.QueryCurrenciesDestroy(sdk.ZeroInt())
		q.CheckSucceeded()
		require.True(t, sdk.ZeroInt().Equal(destroy.ID))
		require.Equal(t, ccSymbol, destroy.Symbol)
		require.Equal(t, ct.ChainID, destroy.ChainID)
		require.Equal(t, ccRecipient, destroy.Recipient)
		require.Equal(t, ccRecipient, destroy.Spender.String())
		require.True(t, destroyAmount.Equal(destroy.Amount))

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
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
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
				q.CheckFailedWithSDKError(ccTypes.ErrWrongIssueID)
			}
		}
	}

	// check currency Query
	{
		q, currency := ct.QueryCurrenciesCurrency(ccSymbol)
		q.CheckSucceeded()

		require.True(t, currency.CurrencyId.IsZero())
		require.Equal(t, ccSymbol, currency.Symbol)
		require.True(t, ccCurAmount.Equal(currency.Supply))
		require.Equal(t, ccDecimals, currency.Decimals)

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
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()

	nomineeAddr := ct.Accounts["nominee"].Address
	assetCode := "eth_dfi"
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
				tx.RemoveCmdArg(assetOracle1)
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
				q.CheckFailedWithSDKError(sdkErrors.ErrUnknownRequest)
			}
		}
	}
}

func Test_PoaCLI(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()

	curValidators := make([]poaTypes.Validator, 0)
	addValidator := func(address, ethAddress string) {
		sdkAddr, err := sdk.AccAddressFromBech32(address)
		require.NoError(t, err, "converting account address")
		curValidators = append(curValidators, poaTypes.Validator{
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
		q, validators := ct.QueryPoaValidators()
		q.CheckSucceeded()
		q, rcvV := ct.QueryPoaValidator(newValidatorAcc.Address)
		q.CheckSucceeded()

		require.Len(t, (*validators).Validators, len(curValidators))
		require.True(t, rcvV.Address.Empty())
		require.Empty(t, rcvV.EthAddress)

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
		issueID := "replaceValidator"

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

		poaGenesis := poaTypes.GenesisState{}
		require.NoError(t, ct.Cdc.UnmarshalJSON(ct.GenesisState()[poaTypes.ModuleName], &poaGenesis))
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
				// non-existing assetCode
				{
					q, _ := ct.QueryPoaValidator("invalid_address")
					q.CheckFailedWithErrorSubstring("address")
				}
			}
			// non-existing validator
			{
				addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
				q, rcvV := ct.QueryPoaValidator(addr.String())
				q.CheckSucceeded()

				require.Empty(t, rcvV.EthAddress)
				require.True(t, rcvV.Address.Empty())
			}
		}
	}
}

func Test_MultiSigCLI(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()

	ccSymbol1, ccSymbol2 := "btc", "usdt"
	ccCurAmount, ccDecimals := sdk.NewInt(1000), int8(1)
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
	ct.TxCurrenciesIssue(ccRecipients[0], ccRecipients[0], ccSymbol1, ccCurAmount, ccDecimals, callUniqueId1).CheckSucceeded()
	ct.TxCurrenciesIssue(ccRecipients[1], ccRecipients[1], ccSymbol2, ccCurAmount, ccDecimals, callUniqueId2).SetGas(300000).CheckSucceeded()

	checkCall := func(call msTypes.CallResp, approved bool, callID uint64, uniqueID, creatorAddr string, votesAddr ...string) {
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
		require.Equal(t, callID, call.Call.MsgID)
		require.Equal(t, uniqueID, call.Call.UniqueID)
		require.NotEmpty(t, call.Call.MsgRoute)
		require.NotEmpty(t, call.Call.MsgType)
	}

	// check calls query
	{
		q, calls := ct.QueryMultiSigCalls()
		q.CheckSucceeded()

		require.Len(t, *calls, 2)
		checkCall((*calls)[0], false, 0, callUniqueId1, ccRecipients[0], ccRecipients[0])
		checkCall((*calls)[1], false, 1, callUniqueId2, ccRecipients[1], ccRecipients[1])
	}

	// check call query
	{
		q, call := ct.QueryMultiSigCall(0)
		q.CheckSucceeded()

		checkCall(*call, false, 0, callUniqueId1, ccRecipients[0], ccRecipients[0])

		// check incorrect inputs
		{
			// invalid number of args
			{
				q, _ := ct.QueryMultiSigCall(0)
				q.RemoveCmdArg("0")
				q.CheckFailedWithErrorSubstring("arg(s)")
			}
			// invalid callID
			{
				q, _ := ct.QueryMultiSigCall(0)
				q.ChangeCmdArg("0", "abc")
				q.CheckFailedWithErrorSubstring("id")
			}
			// non-existing callID
			{
				q, _ := ct.QueryMultiSigCall(2)
				q.CheckFailedWithSDKError(msTypes.ErrWrongCallId)
			}
		}
	}

	// check uniqueCall query
	{
		q, call := ct.QueryMultiSigUnique(callUniqueId1)
		q.CheckSucceeded()

		checkCall(*call, false, 0, callUniqueId1, ccRecipients[0], ccRecipients[0])

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
				q.CheckFailedWithSDKError(msTypes.ErrNotFoundUniqueID)
			}
		}
	}

	// check lastId query
	{
		q, lastId := ct.QueryMultiLastId()
		q.CheckSucceeded()

		require.Equal(t, uint64(1), lastId.LastId)
	}

	// check confirm call Tx
	{
		// add votes for existing call from an other senders
		callID, callUniqueID := uint64(0), callUniqueId1
		votes := []string{ccRecipients[0]}
		for i := 1; i < len(ccRecipients) / 2 + 1; i++ {
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
				tx.RemoveCmdArg(strconv.FormatUint(callID, 10))
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
				tx.ChangeCmdArg(strconv.FormatUint(callID, 10), "not_int")
				tx.CheckFailedWithErrorSubstring("not_int")
			}
		}
	}

	// check revoke confirm Tx
	{
		ct.TxMultiSigRevokeConfirm(ccRecipients[1], 1).CheckSucceeded()

		// check call removed
		q, _ := ct.QueryMultiSigCall(1)
		q.CheckFailedWithSDKError(msTypes.ErrWrongCallId)

		// check incorrect inputs
		{
			// invalid number of args
			{
				tx := ct.TxMultiSigRevokeConfirm(ccRecipients[0], 0)
				tx.RemoveCmdArg(strconv.FormatUint(0, 10))
				tx.CheckFailedWithErrorSubstring("arg(s)")
			}
			// non-existing fromAddress
			{
				tx := ct.TxMultiSigRevokeConfirm(nonExistingAddress.String(), 0)
				tx.CheckFailedWithErrorSubstring(NotFoundErrSubString)
			}
			// invalid callID
			{
				tx := ct.TxMultiSigRevokeConfirm(ccRecipients[0], 0)
				tx.ChangeCmdArg(strconv.FormatUint(0, 10), "not_int")
				tx.CheckFailedWithErrorSubstring("callId")
			}
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
		require.Equal(t, ct.ChainID, resultBlock.Block.ChainID)
		require.GreaterOrEqual(t, resultBlock.Block.Height, int64(1))
	}
}
