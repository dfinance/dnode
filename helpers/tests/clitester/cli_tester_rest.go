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
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	ccsTypes "github.com/dfinance/dnode/x/cc_storage"
	ccTypes "github.com/dfinance/dnode/x/currencies"
	marketTypes "github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	orderTypes "github.com/dfinance/dnode/x/orders"
	poaTypes "github.com/dfinance/dnode/x/poa"
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

func (ct *CLITester) RestQueryCurrenciesIssue(id string) (*RestRequest, *ccTypes.Issue) {
	reqSubPath := fmt.Sprintf("%s/%s/%s", ccTypes.ModuleName, ccTypes.QueryIssue, id)
	respMsg := &ccTypes.Issue{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesCurrency(symbol string) (*RestRequest, *ccsTypes.Currency) {
	reqSubPath := fmt.Sprintf("%s/%s/%s", ccTypes.ModuleName, ccTypes.QueryCurrency, symbol)
	respMsg := &ccsTypes.Currency{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesWithdraw(id sdk.Int) (*RestRequest, *ccTypes.Withdraw) {
	reqSubPath := fmt.Sprintf("%s/%s/%d", ccTypes.ModuleName, ccTypes.QueryWithdraw, id.Int64())
	respMsg := &ccTypes.Withdraw{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesWithdraws(page, limit *int) (*RestRequest, *ccTypes.Withdraws) {
	reqSubPath := fmt.Sprintf("%s/%s", ccTypes.ModuleName, ccTypes.QueryWithdraws)
	respMsg := &ccTypes.Withdraws{}
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

func (ct *CLITester) RestQueryPoaValidators() (*RestRequest, *poaTypes.ValidatorsConfirmationsResp) {
	reqSubPath := fmt.Sprintf("%s/%s", poaTypes.ModuleName, poaTypes.QueryValidators)
	respMsg := &poaTypes.ValidatorsConfirmationsResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOracleAssets() (*RestRequest, *oracle.Assets) {
	reqSubPath := fmt.Sprintf("%s/assets", oracle.ModuleName)
	respMsg := &oracle.Assets{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOracleRawPrices(assetCode string, blockHeight int64) (*RestRequest, *[]oracle.PostedPrice) {
	reqSubPath := fmt.Sprintf("%s/rawprices/%s/%d", oracle.ModuleName, assetCode, blockHeight)
	respMsg := &[]oracle.PostedPrice{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryOraclePrice(assetCode string) (*RestRequest, *oracle.CurrentPrice) {
	reqSubPath := fmt.Sprintf("%s/currentprice/%s", oracle.ModuleName, assetCode)
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

func (ct *CLITester) RestQueryMarket(id dnTypes.ID) (*RestRequest, *marketTypes.Market) {
	reqSubPath := fmt.Sprintf("markets/%s", id.String())
	respMsg := &marketTypes.Market{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMarkets(page, limit int, baseDenom, quoteDenom *string) (*RestRequest, *marketTypes.Markets) {
	reqSubPath := "markets"
	respMsg := &marketTypes.Markets{}

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

func (ct *CLITester) RestQueryOrders(page, limit int, marketIDFilter *dnTypes.ID, directionFilter *orderTypes.Direction, ownerFilter *string) (*RestRequest, *orderTypes.Orders) {
	reqSubPath := "orders"
	respMsg := &orderTypes.Orders{}

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

func (ct *CLITester) RestQueryOrder(id dnTypes.ID) (*RestRequest, *orderTypes.Order) {
	reqSubPath := fmt.Sprintf("%s/%s", orderTypes.ModuleName, id.String())
	respMsg := &orderTypes.Order{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestTxOraclePostPrice(accName, assetCode string, price sdk.Int, receivedAt time.Time) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := oracle.NewMsgPostPrice(acc.Address, assetCode, price, receivedAt)

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestTxOrdersPostOrder(accName string, marketID dnTypes.ID, direction orderTypes.Direction, price, quantity sdk.Uint, ttlInSec uint64) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := orderTypes.MsgPostOrder{
		Owner:     acc.Address,
		MarketID:  marketID,
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

	msg := orderTypes.MsgRevokeOrder{
		Owner:   acc.Address,
		OrderID: id,
	}

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestTxOrdersPostOrderRaw(accName string, accAddress sdk.AccAddress, accNumber, accSequence uint64, marketID dnTypes.ID, direction orderTypes.Direction, price, quantity sdk.Uint, ttlInSec uint64) (*RestRequest, *sdk.TxResponse) {
	msg := orderTypes.MsgPostOrder{
		Owner:     accAddress,
		MarketID:  marketID,
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

	msg := marketTypes.MsgCreateMarket{
		From:            acc.Address,
		BaseAssetDenom:  baseDenom,
		QuoteAssetDenom: quoteDenom,
	}

	return ct.newRestTxRequest(accName, acc, msg, false)
}

func (ct *CLITester) RestLatestBlock() (*RestRequest, *coreTypes.ResultBlock) {
	reqSubPath := "blocks/latest"
	respMsg := &coreTypes.ResultBlock{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}
