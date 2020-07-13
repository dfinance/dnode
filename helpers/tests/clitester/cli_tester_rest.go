package clitester

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkAuthRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/stretchr/testify/require"
	tmCoreTypes "github.com/tendermint/tendermint/rpc/core/types"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orders"
	ordersRest "github.com/dfinance/dnode/x/orders/client/rest"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
)

func (ct *CLITester) newRestTxRequest(accName string, acc *auth.BaseAccount, msg sdk.Msg, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	return ct.newRestTxRequestRaw(accName, acc.GetAccountNumber(), acc.GetSequence(), msg, isSync)
}

func (ct *CLITester) newRestTxRequestRaw(accName string, accNumber, accSequence uint64, msg sdk.Msg, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	// build broadcast Tx request
	txFee := auth.StdFee{
		Amount: sdk.Coins{{Denom: dnConfig.MainDenom, Amount: sdk.NewInt(1)}},
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

func (ct *CLITester) RestQueryOraclePrice(assetCode dnTypes.AssetCode) (*RestRequest, *oracle.CurrentPrice) {
	reqSubPath := fmt.Sprintf("%s/currentprice/%s", oracle.ModuleName, assetCode.String())
	respMsg := &oracle.CurrentPrice{}

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

func (ct *CLITester) RestTxOraclePostPrice(accName string, assetCode dnTypes.AssetCode, price sdk.Int, receivedAt time.Time) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := oracle.NewMsgPostPrice(acc.Address, assetCode, price, receivedAt)

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
