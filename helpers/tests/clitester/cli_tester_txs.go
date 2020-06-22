package clitester

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
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
		ct.IDs.ChainID,
		symbol,
		amount.String(),
		recipientAddr)

	return r
}

func (ct *CLITester) TxOracleAddAsset(nomineeAddress, assetCode string, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"add-asset",
		nomineeAddress,
		assetCode,
		strings.Join(oracleAddresses, ","))

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

func (ct *CLITester) TxOracleSetAsset(nomineeAddress, assetCode string, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"set-asset",
		nomineeAddress,
		assetCode,
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOracleAddOracle(nomineeAddress, assetCode string, oracleAddress string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"add-oracle",
		nomineeAddress,
		assetCode,
		oracleAddress)

	return r
}

func (ct *CLITester) TxOracleSetOracles(nomineeAddress, assetCode string, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"set-oracles",
		nomineeAddress,
		assetCode,
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOraclePostPrice(nomineeAddress, assetCode string, price sdk.Int, receivedAt time.Time) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
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
	cmdArgs = append(cmdArgs, "execute")
	cmdArgs = append(cmdArgs, filePath)
	cmdArgs = append(cmdArgs, args...)

	r := ct.newTxRequest()
	r.SetCmd(
		"vm",
		fromAddress,
		cmdArgs...)
	r.cmd.AddArg("compiler", ct.VMConnection.CompilerAddress)

	return r
}

func (ct *CLITester) TxOrdersPost(ownerAddress string, marketID dnTypes.ID, direction orderTypes.Direction, price, quantity sdk.Uint, ttlInSec int) *TxRequest {
	cmdArgs := []string{
		"post",
		marketID.String(),
		direction.String(),
		price.String(),
		quantity.String(),
		strconv.FormatInt(int64(ttlInSec), 10),
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"orders",
		ownerAddress,
		cmdArgs...)

	return r
}

func (ct *CLITester) TxOrdersRevoke(ownerAddress string, orderID dnTypes.ID) *TxRequest {
	cmdArgs := []string{
		"revoke",
		orderID.String(),
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"orders",
		ownerAddress,
		cmdArgs...)

	return r
}

func (ct *CLITester) TxMarketsAdd(fromAddress string, baseDenom, quoteDenom string) *TxRequest {
	cmdArgs := []string{
		"add",
		baseDenom,
		quoteDenom,
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"markets",
		fromAddress,
		cmdArgs...)

	return r
}

func (ct *CLITester) TxVmDeployModule(fromAddress, filePath string) *TxRequest {
	cmdArgs := []string{
		"deploy",
		filePath,
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"vm",
		fromAddress,
		cmdArgs...)

	return r
}

func (ct *CLITester) TxVmStdlibUpdateProposal(fromAddress, filePath, sourceUrl, updateDesc string, plannedBlockHeight int64, deposit sdk.Coin) *TxRequest {
	cmdArgs := []string{
		"update-stdlib-proposal",
		filePath,
		strconv.FormatInt(plannedBlockHeight, 10),
		sourceUrl,
		strconv.Quote(updateDesc),
		fmt.Sprintf("--deposit=%s", deposit.String()),
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"vm",
		fromAddress,
		cmdArgs...)

	return r
}

func (ct *CLITester) TxGovDeposit(fromAddress string, id uint64, deposit sdk.Coin) *TxRequest {
	cmdArgs := []string{
		"deposit",
		strconv.FormatUint(id, 10),
		deposit.String(),
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"gov",
		fromAddress,
		cmdArgs...)

	return r
}

func (ct *CLITester) TxGovVote(fromAddress string, id uint64, option gov.VoteOption) *TxRequest {
	cmdArgs := []string{
		"vote",
		strconv.FormatUint(id, 10),
		strings.ToLower(option.String()),
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"gov",
		fromAddress,
		cmdArgs...)

	return r
}
