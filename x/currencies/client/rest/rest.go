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
	Denom     = "denom"
	IssueID   = "issueID"
	DestroyID = "destroyID"
)

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/issue/{%s}", types.ModuleName, IssueID), getIssue(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currency/{%s}", types.ModuleName, Denom), getCurrency(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/destroy/{%s}", types.ModuleName, DestroyID), getDestroy(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/destroys", types.ModuleName), getDestroys(cliCtx)).Methods("GET")
}

// GetCurrency godoc
// @Tags currencies
// @Summary Get currency
// @Description Get currency by symbol
// @ID currenciesGetCurrency
// @Accept  json
// @Produce json
// @Param denom path string true "currency denom"
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
// @Tags currencies
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

// GetDestroys godoc
// @Tags currencies
// @Summary Get currency destroys
// @Description Get array of Destroy objects with pagination
// @ID currenciesGetDestroys
// @Accept  json
// @Produce json
// @Param page query uint false "page number (first page: 1)"
// @Param limit query uint false "items per page (default: 100)"
// @Success 200 {object} CCRespGetDestroys
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/destroys [get]
func getDestroys(cliCtx context.CLIContext) http.HandlerFunc {
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
		req := types.DestroysReq{
			Page:  page,
			Limit: limit,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryDestroys), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetDestroy godoc
// @Tags currencies
// @Summary Get currency destroy
// @Description Get currency destroy by destroyID
// @ID currenciesGetDestroy
// @Accept  json
// @Produce json
// @Param destroyID path int true "destroyID"
// @Success 200 {object} CCRespGetDestroy
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/destroy/{destroyID} [get]
func getDestroy(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)
		id, err := helpers.ParseDnIDParam("destroyID", vars[DestroyID], helpers.ParamTypeRestQuery)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		req := types.DestroyReq{
			ID: id,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryDestroy), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
