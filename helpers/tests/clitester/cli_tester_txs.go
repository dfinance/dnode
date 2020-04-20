package clitester

import (
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ct *CLITester) TxCurrenciesIssue(recipientAddr, fromAddr, symbol string, amount sdk.Int, decimals int8, issueID string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
		fromAddr,
		"ms-issue-currency",
		symbol,
		amount.String(),
		strconv.Itoa(int(decimals)),
		recipientAddr,
		issueID)

	return r
}

func (ct *CLITester) TxCurrenciesDestroy(recipientAddr, fromAddr, symbol string, amount sdk.Int) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
		fromAddr,
		"destroy-currency",
		ct.ChainID,
		symbol,
		amount.String(),
		recipientAddr)

	return r
}

func (ct *CLITester) TxPriceFeedAddAsset(nomineeAddress, assetCode string, pricefeedAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"pricefeed",
		"",
		"add-asset",
		nomineeAddress,
		assetCode,
		strings.Join(pricefeedAddresses, ","))

	return r
}

func (ct *CLITester) TxPoaAddValidator(fromAddr, address, ethAddress, issueId string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"poa",
		fromAddr,
		"ms-add-validator",
		address,
		ethAddress,
		issueId)

	return r
}

func (ct *CLITester) TxPoaRemoveValidator(fromAddr, address, issueId string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"poa",
		fromAddr,
		"ms-remove-validator",
		address,
		issueId)

	return r
}

func (ct *CLITester) TxPoaReplaceValidator(fromAddr, targetAddress, address, ethAddress, issueId string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"poa",
		fromAddr,
		"ms-replace-validator",
		targetAddress,
		address,
		ethAddress,
		issueId)

	return r
}

func (ct *CLITester) TxPriceFeedSetAsset(nomineeAddress, assetCode string, pricefeedAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"pricefeed",
		"",
		"set-asset",
		nomineeAddress,
		assetCode,
		strings.Join(pricefeedAddresses, ","))

	return r
}

func (ct *CLITester) TxPriceFeedAddPriceFeed(nomineeAddress, assetCode string, pricefeedAddress string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"pricefeed",
		"",
		"add-pricefeed",
		nomineeAddress,
		assetCode,
		pricefeedAddress)

	return r
}

func (ct *CLITester) TxPriceFeedSetPriceFeeds(nomineeAddress, assetCode string, pricefeedAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"pricefeed",
		"",
		"set-pricefeeds",
		nomineeAddress,
		assetCode,
		strings.Join(pricefeedAddresses, ","))

	return r
}

func (ct *CLITester) TxPriceFeedPostPrice(nomineeAddress, assetCode string, price sdk.Int, receivedAt time.Time) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"pricefeed",
		"",
		"postprice",
		nomineeAddress,
		assetCode,
		price.String(),
		strconv.FormatInt(receivedAt.Unix(), 10))

	return r
}

func (ct *CLITester) TxMultiSigConfirmCall(fromAddress string, callID uint64) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"multisig",
		fromAddress,
		"confirm-call",
		strconv.FormatUint(callID, 10))

	return r
}

func (ct *CLITester) TxMultiSigRevokeConfirm(fromAddress string, callID uint64) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"multisig",
		fromAddress,
		"revoke-confirm",
		strconv.FormatUint(callID, 10))

	return r
}

func (ct *CLITester) TxVmExecuteScript(fromAddress, filePath string, args ...string) *TxRequest {
	cmdArgs := make([]string, 0, 2+len(args))
	cmdArgs = append(cmdArgs, "execute-script")
	cmdArgs = append(cmdArgs, filePath)
	cmdArgs = append(cmdArgs, args...)

	r := ct.newTxRequest()
	r.SetCmd(
		"vm",
		fromAddress,
		cmdArgs...)
	r.cmd.AddArg("compiler", ct.vmCompilerAddress)

	return r
}
