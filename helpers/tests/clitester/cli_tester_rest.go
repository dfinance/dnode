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

	dnConfig "github.com/dfinance/dnode/cmd/config"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

func (ct *CLITester) newRestTxRequest(accName string, acc *auth.BaseAccount, msg sdk.Msg, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	// build broadcast Tx request
	txFee := auth.StdFee{
		Amount: sdk.Coins{{Denom: dnConfig.MainDenom, Amount: sdk.NewInt(1)}},
		Gas:    200000,
	}
	txMemo := "restTxMemo"

	signBytes := auth.StdSignBytes(ct.ChainID, acc.GetAccountNumber(), acc.GetSequence(), txFee, []sdk.Msg{msg}, txMemo)

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

func (ct *CLITester) RestQueryCurrenciesIssue(issueId string) (*RestRequest, *ccTypes.Issue) {
	reqSubPath := fmt.Sprintf("%s/issue/%s", ccTypes.ModuleName, issueId)
	respMsg := &ccTypes.Issue{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesCurrency(symbol string) (*RestRequest, *ccTypes.Currency) {
	reqSubPath := fmt.Sprintf("%s/currency/%s", ccTypes.ModuleName, symbol)
	respMsg := &ccTypes.Currency{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesDestroy(id sdk.Int) (*RestRequest, *ccTypes.Destroy) {
	reqSubPath := fmt.Sprintf("%s/destroy/%d", ccTypes.ModuleName, id.Int64())
	respMsg := &ccTypes.Destroy{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryCurrenciesDestroys(page int, limit *int) (*RestRequest, *ccTypes.Destroys) {
	reqSubPath := fmt.Sprintf("%s/destroys/%d", ccTypes.ModuleName, page)
	respMsg := &ccTypes.Destroys{}
	var reqValues url.Values
	if limit != nil {
		reqValues = url.Values{}
		reqValues.Set("limit", strconv.Itoa(*limit))
	}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, reqValues, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMultiSigCalls() (*RestRequest, *msTypes.CallsResp) {
	reqSubPath := fmt.Sprintf("%s/calls", msTypes.ModuleName)
	respMsg := &msTypes.CallsResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMultiSigCall(callID uint64) (*RestRequest, *msTypes.CallResp) {
	reqSubPath := fmt.Sprintf("%s/call/%d", msTypes.ModuleName, callID)
	respMsg := &msTypes.CallResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryMultiSigUnique(uniqueID string) (*RestRequest, *msTypes.CallResp) {
	reqSubPath := fmt.Sprintf("%s/unique/%s", msTypes.ModuleName, uniqueID)
	respMsg := &msTypes.CallResp{}

	r := ct.newRestRequest().SetQuery("GET", reqSubPath, nil, nil, respMsg)

	return r, respMsg
}

func (ct *CLITester) RestQueryPoaValidators() (*RestRequest, *poaTypes.ValidatorsConfirmations) {
	reqSubPath := fmt.Sprintf("%s/validators", poaTypes.ModuleName)
	respMsg := &poaTypes.ValidatorsConfirmations{}

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

func (ct *CLITester) RestTxOraclePostPrice(accName, assetCode string, price sdk.Int, receivedAt time.Time) (*RestRequest, *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, acc := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	msg := oracle.NewMsgPostPrice(acc.Address, assetCode, price, receivedAt)

	return ct.newRestTxRequest(accName, acc, msg, false)
}
