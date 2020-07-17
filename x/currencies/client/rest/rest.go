package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

const (
	Denom      = "denom"
	IssueID    = "issueID"
	WithdrawID = "withdrawID"
)

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/issue/{%s}", types.ModuleName, IssueID), getIssue(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currency/{%s}", types.ModuleName, Denom), getCurrency(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/withdraw/{%s}", types.ModuleName, WithdrawID), getWithdraw(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/withdraws", types.ModuleName), getWithdraws(cliCtx)).Methods("GET")
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
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
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

// GetIssue godoc
// @Tags Currencies
// @Summary Get currency issue
// @Description Get currency issue by issueID
// @ID currenciesGetIssue
// @Accept  json
// @Produce json
// @Param issueID path string true "issueID"
// @Success 200 {object} CCRespGetIssue
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
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
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
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
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
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
