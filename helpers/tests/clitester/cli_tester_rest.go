package clitester

import (
	"fmt"
	"net/url"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkAuthRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/stretchr/testify/require"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

func (ct *CLITester) newRestTxRequest(httpMethod, subPath string, accName string, msg sdk.Msg, isSync bool) (r *RestRequest, txResp *sdk.TxResponse) {
	accInfo := ct.Accounts[accName]
	require.NotNil(ct.t, accInfo, "account %s: not found", accName)

	accQuery, accData := ct.QueryAccount(accInfo.Address)
	accQuery.CheckSucceeded()

	// build broadcast Tx request
	txFee := auth.StdFee{
		Amount: sdk.Coins{{Denom: dnConfig.MainDenom, Amount: sdk.NewInt(1)}},
		Gas:    200000,
	}
	txMemo := "restTxMemo"

	signBytes := auth.StdSignBytes(ct.ChainID, accData.GetAccountNumber(), accData.GetSequence(), txFee, []sdk.Msg{msg}, txMemo)
	signature, _, err := ct.keyBase.Sign(accName, ct.AccountPassphrase, signBytes)
	require.NoError(ct.t, err, "signing Tx")

	stdSig := auth.StdSignature{
		PubKey:    accData.GetPubKey(),
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
