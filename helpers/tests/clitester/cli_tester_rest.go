package clitester

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	restTypes "github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkAuthRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/stretchr/testify/require"
	tmCoreTypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
	ccRest "github.com/dfinance/dnode/x/currencies/client/rest"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	msRest "github.com/dfinance/dnode/x/multisig/client/rest"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orders"
	ordersRest "github.com/dfinance/dnode/x/orders/client/rest"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
	vmRest "github.com/dfinance/dnode/x/vm/client/rest"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
)

// buildBaseReq returns BaseReq used to prepare REST Tx send.
func (ct *CLITester) buildBaseReq(accName, memo string) restTypes.BaseReq {
	accInfo, ok := ct.Accounts[accName]
	require.True(ct.t, ok, "account %q: not found", accName)

	return restTypes.BaseReq{
		ChainID: ct.IDs.ChainID,
		From:    accInfo.Address,
		Fees:    sdk.Coins{defaults.FeeCoin},
		Gas:     strconv.Itoa(DefaultGas),
		Memo:    memo,
	}
}

func (ct *CLITester) newRestTxRequest(accName string, acc *auth.BaseAccount, msg sdk.Msg, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	return ct.newRestTxRequestRaw(accName, acc.GetAccountNumber(), acc.GetSequence(), msg, isSync)
}

func (ct *CLITester) newRestTxRequestRaw(accName string, accNumber, accSequence uint64, msg sdk.Msg, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	// build broadcast Tx request
	txFee := auth.StdFee{
		Amount: sdk.Coins{{Denom: defaults.MainDenom, Amount: sdk.NewInt(1)}},
		Gas:    DefaultGas,
	}
	txMemo := "restTxMemo"

	signBytes := auth.StdSignBytes(ct.IDs.ChainID, accNumber, accSequence, txFee, []sdk.Msg{msg}, txMemo)

	signature, pubKey, err := ct.keyBase.Sign(accName, ct.AccountPassphrase, signBytes)
	require.NoError(ct.t, err, "signing Tx")

	stdSig := auth.StdSignature{
		PubKey:    pubKey,
		Signature: signature,
	}
	tx := auth.NewStdTx([]sdk.Msg{msg}, txFee, []auth.StdSignature{stdSig}, txMemo)

	txBroadcastReq := sdkAuthRest.BroadcastReq{
		Tx:   tx,
		Mode: "block",
	}
	if isSync {
		txBroadcastReq.Mode = "sync"
	}

	// build REST request
	txResp = &sdk.TxResponse{}
	r = ct.newRestRequest()
	r.SetQuery("POST", "txs", nil, txBroadcastReq, txResp)

	return
}

// NewRestStdTxRequest signs {stdTx} by {accName} and prepares a REST request to send the Tx.
func (ct *CLITester) NewRestStdTxRequest(accName string, stdTx auth.StdTx, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	// get current account info
	accInfo, ok := ct.Accounts[accName]
	require.True(ct.t, ok, "account %q: not found", accName)
	q, acc := ct.QueryAccount(accInfo.Address)
	q.CheckSucceeded()

	// get signature
	signBytes := auth.StdSignBytes(ct.IDs.ChainID, acc.AccountNumber, acc.Sequence, stdTx.Fee, stdTx.Msgs, stdTx.Memo)
	signature, pubKey, err := ct.keyBase.Sign(accName, ct.AccountPassphrase, signBytes)
	require.NoError(ct.t, err, "signing signBytes")
	stdTx.Signatures = []auth.StdSignature{
		{
			PubKey:    pubKey,
			Signature: signature,
		},
	}

	// prepare request
	txBroadcastReq := sdkAuthRest.BroadcastReq{Tx: stdTx, Mode: "block"}
	if isSync {
		txBroadcastReq.Mode = "sync"
	}

	txResp = &sdk.TxResponse{}
	r = ct.newRestRequest()
	r.SetQuery("POST", "txs", nil, txBroadcastReq, txResp)

	return
}

func (ct *CLITester) RestQueryCurrenciesIssue(id string) (*RestRequest, *currencies.Issue) {
	reqSubPath := fmt.Sprintf("%s/%s/%s", currencies.ModuleName, currencies.QueryIssue, id)
	respMsg := &currencies.Issue{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesCurrency(symbol string) (*RestRequest, *ccstorage.Currency) {
	reqSubPath := fmt.Sprintf("%s/%s/%s", currencies.ModuleName, currencies.QueryCurrency, symbol)
	respMsg := &ccstorage.Currency{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesWithdraw(id sdk.Int) (*RestRequest, *currencies.Withdraw) {
	reqSubPath := fmt.Sprintf("%s/%s/%d", currencies.ModuleName, currencies.QueryWithdraw, id.Int64())
	respMsg := &currencies.Withdraw{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesWithdraws(page, limit *int) (*RestRequest, *currencies.Withdraws) {
	reqSubPath := fmt.Sprintf("%s/%s", currencies.ModuleName, currencies.QueryWithdraws)
	respMsg := &currencies.Withdraws{}
	var reqValues url.Values
	if page != nil {
		reqValues = url.Values{}
		reqValues.Set("page", strconv.Itoa(*page))
	}
	if limit != nil {
		reqValues = url.Values{}
		reqValues.Set("limit", strconv.Itoa(*limit))
	}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, reqValues, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMultiSigCalls() (*RestRequest, *multisig.CallsResp) {
	reqSubPath := fmt.Sprintf("%s/calls", multisig.ModuleName)
	respMsg := &multisig.CallsResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMultiSigCall(callID dnTypes.ID) (*RestRequest, *multisig.CallResp) {
	reqSubPath := fmt.Sprintf("%s/call/%s", multisig.ModuleName, callID.String())
	respMsg := &multisig.CallResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMultiSigUnique(uniqueID string) (*RestRequest, *multisig.CallResp) {
	reqSubPath := fmt.Sprintf("%s/unique/%s", multisig.ModuleName, uniqueID)
	respMsg := &multisig.CallResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryPoaValidators() (*RestRequest, *poa.ValidatorsConfirmationsResp) {
	reqSubPath := fmt.Sprintf("%s/%s", poa.ModuleName, poa.QueryValidators)
	respMsg := &poa.ValidatorsConfirmationsResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOracleAssets() (*RestRequest, *oracle.Assets) {
	reqSubPath := fmt.Sprintf("%s/assets", oracle.ModuleName)
	respMsg := &oracle.Assets{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOracleRawPrices(assetCode dnTypes.AssetCode, blockHeight int64) (*RestRequest, *[]oracle.PostedPrice) {
	reqSubPath := fmt.Sprintf("%s/rawprices/%s/%d", oracle.ModuleName, assetCode.String(), blockHeight)
	respMsg := &[]oracle.PostedPrice{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOraclePrice(assetCode dnTypes.AssetCode) (*RestRequest, *oracle.CurrentAssetPrice) {
	reqSubPath := fmt.Sprintf("%s/currentprice/%s", oracle.ModuleName, assetCode.String())
	respMsg := &oracle.CurrentAssetPrice{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryVMGetData(accAddr, path string) (*RestRequest, *vm.QueryValueResp) {
	reqSubPath := fmt.Sprintf("%s/data/%s/%s", vm.ModuleName, accAddr, path)
	respMsg := &vm.QueryValueResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryAuthAccount(address string) (*RestRequest, *auth.BaseAccount) {
	reqSubPath := fmt.Sprintf("auth/accounts/%s", address)
	respMsg := &auth.BaseAccount{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMarket(id dnTypes.ID) (*RestRequest, *markets.Market) {
	reqSubPath := fmt.Sprintf("markets/%s", id.String())
	respMsg := &markets.Market{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMarkets(page, limit int, baseDenom, quoteDenom *string) (*RestRequest, *markets.Markets) {
	reqSubPath := "markets"
	respMsg := &markets.Markets{}

	reqValues := url.Values{}
	if page != -1 {
		reqValues.Set("page", strconv.Itoa(page))
	}
	if limit != -1 {
		reqValues.Set("limit", strconv.Itoa(limit))
	}
	if baseDenom != nil {
		reqValues.Set("baseAssetDenom", *baseDenom)
	}
	if quoteDenom != nil {
		reqValues.Set("quoteAssetDenom", *quoteDenom)
	}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, reqValues, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOrders(page, limit int, marketIDFilter *dnTypes.ID, directionFilter *orders.Direction, ownerFilter *string) (*RestRequest, *orders.Orders) {
	reqSubPath := "orders"
	respMsg := &orders.Orders{}

	reqValues := url.Values{}
	if page != -1 {
		reqValues.Set("page", strconv.Itoa(page))
	}
	if limit != -1 {
		reqValues.Set("limit", strconv.Itoa(limit))
	}
	if marketIDFilter != nil {
		reqValues.Set("marketID", marketIDFilter.String())
	}
	if directionFilter != nil {
		reqValues.Set("direction", directionFilter.String())
	}
	if ownerFilter != nil {
		reqValues.Set("owner", *ownerFilter)
	}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, reqValues, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOrder(id dnTypes.ID) (*RestRequest, *orders.Order) {
	reqSubPath := fmt.Sprintf("%s/%s", orders.ModuleName, id.String())
	respMsg := &orders.Order{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOrderPost(rq ordersRest.PostOrderReq) (*RestRequest, *auth.StdTx) {
	reqSubPath := fmt.Sprintf("%s/%s", orders.ModuleName, "post")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryOrderRevoke(rq ordersRest.RevokeOrderReq) (*RestRequest, *auth.StdTx) {
	reqSubPath := fmt.Sprintf("%s/%s", orders.ModuleName, "revoke")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesIssueStdTx(creatorAccName, payeeAccName, issueID string, coin sdk.Coin, memo string) (*RestRequest, *auth.StdTx) {
	payeeAccInfo, ok := ct.Accounts[payeeAccName]
	require.True(ct.t, ok, "payee account %q: not found", payeeAccName)

	rq := ccRest.SubmitIssueReq{
		BaseReq: ct.buildBaseReq(creatorAccName, memo),
		ID:      issueID,
		Coin:    coin,
		Payee:   payeeAccInfo.Address,
	}

	reqSubPath := fmt.Sprintf("%s/%s", currencies.ModuleName, "issue")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesWithdrawStdTx(payerAccName, pegZonePayeeAddress, pegZoneChainID string, coin sdk.Coin, memo string) (*RestRequest, *auth.StdTx) {
	rq := ccRest.WithdrawReq{
		BaseReq:        ct.buildBaseReq(payerAccName, memo),
		Coin:           coin,
		PegZonePayee:   pegZonePayeeAddress,
		PegZoneChainID: pegZoneChainID,
	}

	reqSubPath := fmt.Sprintf("%s/%s", currencies.ModuleName, "withdraw")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryMultisigConfirmStdTx(validatorAccName string, callID dnTypes.ID, memo string) (*RestRequest, *auth.StdTx) {
	rq := msRest.ConfirmReq{
		BaseReq: ct.buildBaseReq(validatorAccName, memo),
		CallID:  callID.String(),
	}

	reqSubPath := fmt.Sprintf("%s/%s", multisig.ModuleName, "confirm")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryMultisigRevokeStdTx(validatorAccName string, callID dnTypes.ID, memo string) (*RestRequest, *auth.StdTx) {
	rq := msRest.RevokeReq{
		BaseReq: ct.buildBaseReq(validatorAccName, memo),
		CallID:  callID.String(),
	}

	reqSubPath := fmt.Sprintf("%s/%s", multisig.ModuleName, "revoke")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryVMCompile(address, code string) (*RestRequest, *vm_client.CompiledItems) {
	req := vmRest.CompileReq{
		Code:    code,
		Account: address,
	}

	reqSubPath := fmt.Sprintf("%s/%s", vm.ModuleName, "compile")
	respMsg := &vm_client.CompiledItems{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, req, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryVMExecuteScriptStdTx(senderAccName string, byteCode, memo string, args ...string) (*RestRequest, *auth.StdTx) {
	rq := vmRest.ExecuteScriptReq{
		BaseReq:  ct.buildBaseReq(senderAccName, memo),
		MoveCode: byteCode,
		MoveArgs: args,
	}

	reqSubPath := fmt.Sprintf("%s/%s", vm.ModuleName, "execute")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryVMPublishModuleStdTx(senderAccName string, byteCode []string, memo string) (*RestRequest, *auth.StdTx) {
	rq := vmRest.PublishModuleReq{
		BaseReq:  ct.buildBaseReq(senderAccName, memo),
		MoveCode: byteCode,
	}

	reqSubPath := fmt.Sprintf("%s/%s", vm.ModuleName, "publish")
	respMsg := &auth.StdTx{}
	r := ct.newRestRequest().SetQuery("PUT", reqSubPath, nil, rq, respMsg)
	return r, respMsg
}

func (ct *CLITester) RestQueryVMLcsView(address, movePath, viewRequest string) (*RestRequest, *vmRest.LcsViewResp) {
	req := vmRest.LcsViewReq{
		Account:     address,
		MovePath:    movePath,
		ViewRequest: viewRequest,
	}

	reqSubPath := fmt.Sprintf("%s/%s", vm.ModuleName, "view")
	respMsg := &vmRest.LcsViewResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, req, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestTxOraclePostPrice(accName string, assetCode dnTypes.AssetCode, askPrice, bidPrice sdk.Int, receivedAt time.Time) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := oracle.NewMsgPostPrice(acc.Address, assetCode, askPrice, bidPrice, receivedAt)

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestTxOrdersPostOrder(accName string, assetCode dnTypes.AssetCode, direction orders.Direction, price, quantity sdk.Uint, ttlInSec uint64) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := orders.MsgPostOrder{
		Owner:     acc.Address,
		AssetCode: assetCode,
		Direction: direction,
		Price:     price,
		Quantity:  quantity,
		TtlInSec:  ttlInSec,
	}

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestTxOrdersRevokeOrder(accName string, id dnTypes.ID) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := orders.MsgRevokeOrder{
		Owner:   acc.Address,
		OrderID: id,
	}

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestTxOrdersPostOrderRaw(accName string, accAddress sdk.AccAddress, accNumber, accSequence uint64, assetCode dnTypes.AssetCode, direction orders.Direction, price, quantity sdk.Uint, ttlInSec uint64) (*RestRequest, *sdk.TxResponse) {
	msg := orders.MsgPostOrder{
		Owner:     accAddress,
		AssetCode: assetCode,
		Direction: direction,
		Price:     price,
		Quantity:  quantity,
		TtlInSec:  ttlInSec,
	}

	return ct.newRestTxRequestRaw(accName, accNumber, accSequence, msg, true)
}

func (ct *CLITester) RestTxBankTransfer(from, to string, amount uint64, denom string) (*RestRequest, *sdk.TxResponse) {
	fromAcc := ct.Accounts[from]
	require.NotNil(ct.t, fromAcc, "account %s: not found", from)

	toAcc := ct.Accounts[to]
	require.NotNil(ct.t, toAcc, "account %s: not found", to)

	accQuery, acc := ct.QueryAccount(fromAcc.Address)
	accQuery.CheckSucceeded()

	fromHexAcc, err := sdk.AccAddressFromBech32(fromAcc.Address)
	require.Nil(ct.t, err)

	toHexAcc, err := sdk.AccAddressFromBech32(toAcc.Address)
	require.Nil(ct.t, err)

	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromUint64(amount)))
	msg := bank.NewMsgSend(fromHexAcc, toHexAcc, coins)

	return ct.newRestTxRequest(from, acc, msg, false)
}

func (ct *CLITester) RestTxGovTransfer(from string, proposalId uint64, amount uint64, denom string) (*RestRequest, *sdk.TxResponse) {
	fromAcc := ct.Accounts[from]
	require.NotNil(ct.t, fromAcc, "account %s: not found", from)

	accQuery, acc := ct.QueryAccount(fromAcc.Address)
	accQuery.CheckSucceeded()

	fromHexAcc, err := sdk.AccAddressFromBech32(fromAcc.Address)
	require.Nil(ct.t, err)

	coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewIntFromUint64(amount)))
	msg := gov.NewMsgDeposit(fromHexAcc, proposalId, coins)

	return ct.newRestTxRequest(from, acc, msg, false)
}

func (ct *CLITester) RestTxMarketsAdd(accName, baseDenom, quoteDenom string) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := markets.MsgCreateMarket{
		From:            acc.Address,
		BaseAssetDenom:  baseDenom,
		QuoteAssetDenom: quoteDenom,
	}

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestLatestBlock() (*RestRequest, *tmCoreTypes.ResultBlock) {
	reqSubPath := "blocks/latest"
	respMsg := &tmCoreTypes.ResultBlock{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}
