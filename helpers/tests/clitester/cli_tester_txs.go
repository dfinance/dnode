package clitester

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders"
)

func (ct *CLITester) TxCurrenciesIssue(payeeAddr, fromAddr, issueID, denom string, amount sdk.Int) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
		fromAddr,
		"ms-issue",
		issueID,
		sdk.NewCoin(denom, amount).String(),
		payeeAddr)

	return r
}

func (ct *CLITester) TxCurrenciesWithdraw(recipientAddr, fromAddr, denom string, amount sdk.Int) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
		fromAddr,
		"withdraw",
		sdk.NewCoin(denom, amount).String(),
		recipientAddr,
		ct.IDs.ChainID)

	return r
}

func (ct *CLITester) TxOracleAddAsset(nomineeAddress string, assetCode dnTypes.AssetCode, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"add-asset",
		nomineeAddress,
		assetCode.String(),
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxPoaAddValidator(fromAddr, address, ethAddress, issueID string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"poa",
		fromAddr,
		"ms-add-validator",
		issueID,
		address,
		ethAddress,
	)

	return r
}

func (ct *CLITester) TxPoaRemoveValidator(fromAddr, address, issueID string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"poa",
		fromAddr,
		"ms-remove-validator",
		issueID,
		address,
	)

	return r
}

func (ct *CLITester) TxPoaReplaceValidator(fromAddr, targetAddress, address, ethAddress, issueID string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"poa",
		fromAddr,
		"ms-replace-validator",
		issueID,
		targetAddress,
		address,
		ethAddress,
	)

	return r
}

func (ct *CLITester) TxOracleSetAsset(nomineeAddress string, assetCode dnTypes.AssetCode, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"set-asset",
		nomineeAddress,
		assetCode.String(),
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOracleAddOracle(nomineeAddress string, assetCode dnTypes.AssetCode, oracleAddress string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"add-oracle",
		nomineeAddress,
		assetCode.String(),
		oracleAddress)

	return r
}

func (ct *CLITester) TxOracleSetOracles(nomineeAddress string, assetCode dnTypes.AssetCode, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"set-oracles",
		nomineeAddress,
		assetCode.String(),
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOraclePostPrice(nomineeAddress string, assetCode dnTypes.AssetCode, price sdk.Int, receivedAt time.Time) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"postprice",
		nomineeAddress,
		assetCode.String(),
		price.String(),
		strconv.FormatInt(receivedAt.Unix(), 10),
	)

	return r
}

func (ct *CLITester) TxMultiSigConfirmCall(fromAddress string, callID dnTypes.ID) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"multisig",
		fromAddress,
		"confirm-call",
		callID.String(),
	)

	return r
}

func (ct *CLITester) TxMultiSigRevokeConfirm(fromAddress string, callID dnTypes.ID) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"multisig",
		fromAddress,
		"revoke-confirm",
		callID.String(),
	)

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

func (ct *CLITester) TxOrdersPost(ownerAddress string, assetCode dnTypes.AssetCode, direction orders.Direction, price, quantity sdk.Uint, ttlInSec int) *TxRequest {
	cmdArgs := []string{
		"post",
		assetCode.String(),
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

func (ct *CLITester) TxCCAddCurrencyProposal(fromAddress, denom, balancePath, infoPath string, decimals uint8, deposit sdk.Coin) *TxRequest {
	cmdArgs := []string{
		"add-currency-proposal",
		denom,
		strconv.FormatUint(uint64(decimals), 10),
		balancePath,
		infoPath,
		fmt.Sprintf("--deposit=%s", deposit.String()),
	}

	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
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
