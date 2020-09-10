package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
	msClient "github.com/dfinance/dnode/x/multisig/client"
)

const (
	Denom      = "denom"
	IssueID    = "issueID"
	WithdrawID = "withdrawID"
)

type SubmitIssueReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Issue unique ID (could be txHash of transaction in another blockchain)
	ID string `json:"id" yaml:"id"`
	// Target currency issue coin
	Coin sdk.Coin `json:"coin" yaml:"coin"`
	// Payee account (whose balance is increased)
	Payee string `json:"payee" yaml:"payee" format:"bech32/hex" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

type UnstakeReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Unstake unique ID (could be txHash of transaction in another blockchain)
	ID string `json:"id" yaml:"id"`
	// Staker account (whose balance is increased)
	Staker string `json:"staker" yaml:"staker" format:"bech32/hex" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

type WithdrawReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Target currency withdraw coin
	Coin sdk.Coin `json:"coin" yaml:"coin"`
	// Second blockchain: payee account (whose balance is increased)
	PegZonePayee string `json:"pegzone_payee" yaml:"pegzone_payee"`
	// Second blockchain: ID
	PegZoneChainID string `json:"pegzone_chain_id" yaml:"pegzone_chain_id"`
}

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/currency/{%s}", types.ModuleName, Denom), getCurrency(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s", types.ModuleName), getCurrencies(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/issue/{%s}", types.ModuleName, IssueID), getIssue(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/withdraw/{%s}", types.ModuleName, WithdrawID), getWithdraw(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/withdraws", types.ModuleName), getWithdraws(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/issue", types.ModuleName), submitIssue(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/unstake", types.ModuleName), submitUnstake(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/withdraw", types.ModuleName), withdraw(cliCtx)).Methods("PUT")
}

// GetCurrency godoc
// @Tags Currencies
// @Summary Get currency
// @Description Get currency by denom
// @ID currenciesGetCurrency
// @Accept  json
// @Produce json
// @Param denom path string true "currency denomination symbol"
// @Success 200 {object} CCRespGetCurrency
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/currency/{denom} [get]
func getCurrency(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)

		req := types.CurrencyReq{Denom: vars[Denom]}
		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCurrency), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetCurrencies godoc
// @Tags Currencies
// @Summary Get all registered currencies
// @ID currenciesGetCurrencies
// @Accept  json
// @Produce json
// @Success 200 {object} CCRespGetCurrencies
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies [get]
func getCurrencies(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCurrencies), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetIssue godoc
// @Tags Currencies
// @Summary Get currency issue
// @Description Get currency issue by issueID
// @ID currenciesGetIssue
// @Accept  json
// @Produce json
// @Param issueID path string true "issueID"
// @Success 200 {object} CCRespGetIssue
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/issue/{issueID} [get]
func getIssue(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)

		req := types.IssueReq{ID: vars[IssueID]}
		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryIssue), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetWithdraws godoc
// @Tags Currencies
// @Summary Get currency withdraws
// @Description Get array of Withdraw objects with pagination
// @ID currenciesGetWithdraws
// @Accept  json
// @Produce json
// @Param page query uint false "page number (first page: 1)"
// @Param limit query uint false "items per page (default: 100)"
// @Success 200 {object} CCRespGetWithdraws
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/withdraws [get]
func getWithdraws(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")
		page, limit, err := helpers.ParsePaginationParams(pageStr, limitStr, helpers.ParamTypeRestQuery)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// prepare request
		req := types.WithdrawsReq{
			Page:  page,
			Limit: limit,
		}
		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryWithdraws), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetWithdraw godoc
// @Tags Currencies
// @Summary Get currency withdraw
// @Description Get currency withdraw by withdrawID
// @ID currenciesGetWithdraw
// @Accept  json
// @Produce json
// @Param withdrawID path int true "withdrawID"
// @Success 200 {object} CCRespGetWithdraw
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/withdraw/{withdrawID} [get]
func getWithdraw(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)
		id, err := helpers.ParseDnIDParam(WithdrawID, vars[WithdrawID], helpers.ParamTypeRestQuery)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req := types.WithdrawReq{
			ID: id,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryWithdraw), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// SubmitIssue godoc
// @Tags Currencies
// @Summary Submit issue
// @Description Get submit new issue multi signature message stdTx object
// @ID currenciesSubmitUnstake
// @Accept  json
// @Produce json
// @Param request body SubmitIssueReq true "Submit issue request"
// @Success 200 {object} CCRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/issue [put]
func submitIssue(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req SubmitIssueReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		coin, err := helpers.ParseCoinParam("coin", req.Coin.String(), helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		payeeAddr, err := helpers.ParseSdkAddressParam("payee", req.Payee, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		issueID := req.ID

		// create the message
		msg := types.NewMsgIssueCurrency(issueID, coin, payeeAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		callMsg := msClient.NewMsgSubmitCall(msg, issueID, fromAddr)
		if err := callMsg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{callMsg})
	}
}

// Submit unstake godoc
// @Tags Currencies
// @Summary Unstake tx
// @Description Get new unstake multi signature message stdTx object
// @ID currenciesSubmitIssue
// @Accept  json
// @Produce json
// @Param request body UnstakeReq true "Submit unstake request"
// @Success 200 {object} CCRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/unstake [put]
func submitUnstake(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req UnstakeReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		stakerAddr, err := helpers.ParseSdkAddressParam("staker", req.Staker, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		issueID := req.ID

		// create the message
		msg := types.NewMsgUnstakeCurrency(issueID, stakerAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		callMsg := msClient.NewMsgSubmitCall(msg, issueID, fromAddr)
		if err := callMsg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{callMsg})
	}
}

// Withdraw godoc
// @Tags Currencies
// @Summary Withdraw currency
// @Description Get withdraw currency coins from account balance stdTx object
// @ID currenciesWithdraw
// @Accept  json
// @Produce json
// @Param request body WithdrawReq true "Withdraw request"
// @Success 200 {object} CCRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/withdraw [put]
func withdraw(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req WithdrawReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		coin, err := helpers.ParseCoinParam("coin", req.Coin.String(), helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		pzPayee := req.PegZonePayee
		pzChainID := req.PegZoneChainID

		// create the message
		msg := types.NewMsgWithdrawCurrency(coin, fromAddr, pzPayee, pzChainID)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
