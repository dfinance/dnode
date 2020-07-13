package clitester

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	tmCoreTypes "github.com/tendermint/tendermint/rpc/core/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
)

func (ct *CLITester) QueryTx(txHash string) (*QueryRequest, *sdk.TxResponse) {
	resObj := &sdk.TxResponse{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("tx", txHash)

	return q, resObj
}

func (ct *CLITester) QueryStatus() (*QueryRequest, *tmCoreTypes.ResultStatus) {
	resObj := &tmCoreTypes.ResultStatus{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("status")
	q.RemoveCmdArg("query")

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesIssue(id string) (*QueryRequest, *currencies.Issue) {
	resObj := &currencies.Issue{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "issue", id)

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesWithdraw(id dnTypes.ID) (*QueryRequest, *currencies.Withdraw) {
	resObj := &currencies.Withdraw{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "withdraw", id.String())

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesWithdraws(page, limit int) (*QueryRequest, *currencies.Withdraws) {
	resObj := &currencies.Withdraws{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "withdraws")

	if page > 0 {
		q.cmd.AddArg("page", strconv.FormatInt(int64(page), 10))
	}
	if limit > 0 {
		q.cmd.AddArg("limit", strconv.FormatInt(int64(limit), 10))
	}

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesCurrency(denom string) (*QueryRequest, *ccstorage.Currency) {
	resObj := &ccstorage.Currency{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "currency", denom)

	return q, resObj
}

func (ct *CLITester) QueryOracleAssets() (*QueryRequest, *oracle.Assets) {
	resObj := &oracle.Assets{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("oracle", "assets")

	return q, resObj
}

func (ct *CLITester) QueryOracleRawPrices(assetCode dnTypes.AssetCode, blockHeight int64) (*QueryRequest, *[]oracle.PostedPrice) {
	resObj := &[]oracle.PostedPrice{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd(
		"oracle",
		"rawprices",
		assetCode.String(),
		strconv.FormatInt(blockHeight, 10))

	return q, resObj
}

func (ct *CLITester) QueryOraclePrice(assetCode dnTypes.AssetCode) (*QueryRequest, *oracle.CurrentPrice) {
	resObj := &oracle.CurrentPrice{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd(
		"oracle",
		"price",
		assetCode.String())

	return q, resObj
}

func (ct *CLITester) QueryPoaValidators() (*QueryRequest, *poa.ValidatorsConfirmationsResp) {
	resObj := &poa.ValidatorsConfirmationsResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("poa", "validators")

	return q, resObj
}

func (ct *CLITester) QueryPoaValidator(address string) (*QueryRequest, *poa.Validator) {
	resObj := &poa.Validator{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("poa", "validator", address)

	return q, resObj
}

func (ct *CLITester) QueryPoaMinMax() (*QueryRequest, *poa.Params) {
	resObj := &poa.Params{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("poa", "minmax")

	return q, resObj
}

func (ct *CLITester) QueryMultiSigUnique(uniqueID string) (*QueryRequest, *multisig.CallResp) {
	resObj := &multisig.CallResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "unique", uniqueID)

	return q, resObj
}

func (ct *CLITester) QueryMultiSigCall(callID dnTypes.ID) (*QueryRequest, *multisig.CallResp) {
	resObj := &multisig.CallResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "call", callID.String())

	return q, resObj
}

func (ct *CLITester) QueryMultiSigCalls() (*QueryRequest, *multisig.CallsResp) {
	resObj := &multisig.CallsResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "calls")

	return q, resObj
}

func (ct *CLITester) QueryMultiLastId() (*QueryRequest, *multisig.LastCallIdResp) {
	resObj := &multisig.LastCallIdResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "lastId")

	return q, resObj
}

func (ct *CLITester) QueryVmCompile(moveFilePath, savePath, accountAddress string) *QueryRequest {
	q := ct.newQueryRequest(nil)
	q.SetCmd("vm", "compile", moveFilePath, accountAddress)
	q.cmd.AddArg("compiler", ct.VMConnection.CompilerAddress)
	q.cmd.AddArg("to-file", savePath)

	return q
}

func (ct *CLITester) QueryVmData(hexAddress, hexPath string) (*QueryRequest, *vm.QueryValueResp) {
	resObj := &vm.QueryValueResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("vm", "get-data", hexAddress, hexPath)

	return q, resObj
}

func (ct *CLITester) QueryAccount(address string) (*QueryRequest, *auth.BaseAccount) {
	resObj := &auth.BaseAccount{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("account", address)
	q.cmd.AddArg("node", ct.NodePorts.RPCAddress)

	return q, resObj
}

func (ct *CLITester) QueryModuleAccount(address string) (*QueryRequest, *supply.ModuleAccount) {
	resObj := &supply.ModuleAccount{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("account", address)
	q.cmd.AddArg("node", ct.NodePorts.RPCAddress)

	return q, resObj
}

func (ct *CLITester) QueryAuthAccount(address string) (*QueryRequest, *auth.BaseAccount) {
	resObj := &auth.BaseAccount{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("auth", "account", address)
	q.cmd.AddArg("node", ct.NodePorts.RPCAddress)

	return q, resObj
}

func (ct *CLITester) QueryOrdersOrder(id dnTypes.ID) (*QueryRequest, *orders.Order) {
	resObj := &orders.Order{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("orders", "order", id.String())

	return q, resObj
}

func (ct *CLITester) QueryOrdersList(page, limit int, marketIDFilter *dnTypes.ID, directionFilter *orders.Direction, ownerFilter *string) (*QueryRequest, *orders.Orders) {
	resObj := &orders.Orders{}

	q := ct.newQueryRequest(resObj)
	q.SetCmd("orders", "list")

	if page > 0 {
		q.cmd.AddArg("page", strconv.FormatInt(int64(page), 10))
	}
	if limit > 0 {
		q.cmd.AddArg("limit", strconv.FormatInt(int64(limit), 10))
	}
	if marketIDFilter != nil {
		q.cmd.AddArg("market-id", marketIDFilter.String())
	}
	if directionFilter != nil {
		q.cmd.AddArg("direction", directionFilter.String())
	}
	if ownerFilter != nil {
		q.cmd.AddArg("owner", *ownerFilter)
	}

	return q, resObj
}

func (ct *CLITester) QueryMarketsMarket(id dnTypes.ID) (*QueryRequest, *markets.Market) {
	resObj := &markets.Market{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("markets", "market", id.String())

	return q, resObj
}

func (ct *CLITester) QueryMarketsList(page, limit int, baseDenom, quoteDenom *string) (*QueryRequest, *markets.Markets) {
	resObj := &markets.Markets{}

	q := ct.newQueryRequest(resObj)
	q.SetCmd("markets", "list")

	if page > 0 {
		q.cmd.AddArg("page", strconv.FormatInt(int64(page), 10))
	}
	if limit > 0 {
		q.cmd.AddArg("limit", strconv.FormatInt(int64(limit), 10))
	}
	if baseDenom != nil {
		q.cmd.AddArg("base-asset-denom", *baseDenom)
	}
	if quoteDenom != nil {
		q.cmd.AddArg("quote-asset-denom", *quoteDenom)
	}

	return q, resObj
}

func (ct *CLITester) QueryGovProposal(id uint64) (*QueryRequest, *gov.Proposal) {
	resObj := &gov.Proposal{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("gov", "proposal", strconv.FormatUint(id, 10))

	return q, resObj
}

func (ct *CLITester) QueryGovProposals(page, limit int, depositorFilter, voterFilter *string, statusFilter *govTypes.ProposalStatus) (*QueryRequest, *gov.Proposals) {
	resObj := &gov.Proposals{}

	q := ct.newQueryRequest(resObj)
	q.SetCmd("gov", "proposals")

	if page > 0 {
		q.cmd.AddArg("page", strconv.FormatInt(int64(page), 10))
	}
	if limit > 0 {
		q.cmd.AddArg("limit", strconv.FormatInt(int64(limit), 10))
	}
	if depositorFilter != nil {
		q.cmd.AddArg("depositor", *depositorFilter)
	}
	if statusFilter != nil {
		q.cmd.AddArg("status", statusFilter.String())
	}
	if voterFilter != nil {
		q.cmd.AddArg("voter", *voterFilter)
	}

	return q, resObj
}

func (ct *CLITester) QuerySupply(denom string) (*QueryRequest, *sdk.Int) {
	resObj := &sdk.Int{}

	q := ct.newQueryRequest(resObj)
	q.SetCmd("supply", "total", denom)

	return q, resObj
}

func (ct *CLITester) QueryStakingValidators() (*QueryRequest, *staking.Validators) {
	resObj := &staking.Validators{}

	q := ct.newQueryRequest(resObj)
	q.SetCmd("staking", "validators")

	return q, resObj
}
