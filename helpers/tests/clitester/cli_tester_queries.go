package clitester

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

func (ct *CLITester) QueryCurrenciesIssue(issueID string) (*QueryRequest, *ccTypes.Issue) {
	resObj := &ccTypes.Issue{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "issue", issueID)

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesDestroy(id sdk.Int) (*QueryRequest, *ccTypes.Destroy) {
	resObj := &ccTypes.Destroy{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "destroy", id.String())

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesDestroys(page, limit int) (*QueryRequest, *ccTypes.Destroys) {
	resObj := &ccTypes.Destroys{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "destroys", strconv.Itoa(page), strconv.Itoa(limit))

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesCurrency(symbol string) (*QueryRequest, *ccTypes.Currency) {
	resObj := &ccTypes.Currency{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "currency", symbol)

	return q, resObj
}

func (ct *CLITester) QueryOracleAssets() (*QueryRequest, *oracle.Assets) {
	resObj := &oracle.Assets{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("oracle", "assets")

	return q, resObj
}

func (ct *CLITester) QueryOracleRawPrices(assetCode string, blockHeight int64) (*QueryRequest, *[]oracle.PostedPrice) {
	resObj := &[]oracle.PostedPrice{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd(
		"oracle",
		"rawprices",
		assetCode,
		strconv.FormatInt(blockHeight, 10))

	return q, resObj
}

func (ct *CLITester) QueryOraclePrice(assetCode string) (*QueryRequest, *oracle.CurrentPrice) {
	resObj := &oracle.CurrentPrice{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd(
		"oracle",
		"price",
		assetCode)

	return q, resObj
}

func (ct *CLITester) QueryPoaValidators() (*QueryRequest, *poaTypes.ValidatorsConfirmations) {
	resObj := &poaTypes.ValidatorsConfirmations{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("poa", "validators")

	return q, resObj
}

func (ct *CLITester) QueryPoaValidator(address string) (*QueryRequest, *poaTypes.Validator) {
	resObj := &poaTypes.Validator{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("poa", "validator", address)

	return q, resObj
}

func (ct *CLITester) QueryPoaMinMax() (*QueryRequest, *poaTypes.Params) {
	resObj := &poaTypes.Params{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("poa", "minmax")

	return q, resObj
}

func (ct *CLITester) QueryMultiSigUnique(uniqueID string) (*QueryRequest, *msTypes.CallResp) {
	resObj := &msTypes.CallResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "unique", uniqueID)

	return q, resObj
}

func (ct *CLITester) QueryMultiSigCall(callID uint64) (*QueryRequest, *msTypes.CallResp) {
	resObj := &msTypes.CallResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "call", strconv.FormatUint(callID, 10))

	return q, resObj
}

func (ct *CLITester) QueryMultiSigCalls() (*QueryRequest, *msTypes.CallsResp) {
	resObj := &msTypes.CallsResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "calls")

	return q, resObj
}

func (ct *CLITester) QueryMultiLastId() (*QueryRequest, *msTypes.LastIdRes) {
	resObj := &msTypes.LastIdRes{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "lastId")

	return q, resObj
}

func (ct *CLITester) QueryVmCompileScript(mvirFilePath, savePath, accountAddress string) *QueryRequest {
	q := ct.newQueryRequest(nil)
	q.SetCmd("vm", "compile-script", mvirFilePath, accountAddress)
	q.cmd.AddArg("compiler", ct.vmCompilerAddress)
	q.cmd.AddArg("to-file", savePath)

	return q
}

func (ct *CLITester) QueryAccount(address string) (*QueryRequest, *auth.BaseAccount) {
	resObj := &auth.BaseAccount{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("account", address)

	return q, resObj
}
