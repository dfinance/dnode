// Implements REST API calls for currency module.
package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/x/currencies/types"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/issue/{issueID}", types.ModuleName), getIssue(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currency/{symbol}", types.ModuleName), getCurrency(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/destroy/{destroyID}", types.ModuleName), getDestroy(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/destroys/{page}", types.ModuleName), getDestroys(cliCtx)).Methods("GET")
}

// GetDestroys godoc
// @Tags currencies
// @Summary Get currency destroys
// @Description Get array of Destroy objects with pagination
// @ID currenciesGetDestroys
// @Accept  json
// @Produce json
// @Param page path int true "page number"
// @Param limit query int false "items per page (default: 100)"
// @Success 200 {object} CCRespGetDestroys
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/destroys/{page} [get]
func getDestroys(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		page, isOk := sdk.NewIntFromString(vars["page"])
		if !isOk {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("%s is not a number, cant parse int", page))
			return
		}

		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "100"
		}

		parsedLimit, isOk := sdk.NewIntFromString(limit)
		if !isOk {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("%s is not a number, cant parse int", page))
			return
		}

		req := types.DestroysReq{
			Page:  page,
			Limit: parsedLimit,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/destroys", types.ModuleName), bz)
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
		vars := mux.Vars(r)
		destroyID, isOk := sdk.NewIntFromString(vars["destroyID"])
		if !isOk {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("%s is not a number, cant parse int", destroyID))
			return
		}

		req := types.DestroyReq{
			DestroyId: destroyID,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/destroy", types.ModuleName), bz)
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
		vars := mux.Vars(r)
		req := types.IssueReq{IssueID: vars["issueID"]}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/issue", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetCurrency godoc
// @Tags currencies
// @Summary Get currency
// @Description Get currency by symbol
// @ID currenciesGetCurrency
// @Accept  json
// @Produce json
// @Param symbol path string true "symbol / denom"
// @Success 200 {object} CCRespGetCurrency
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /currencies/currency/{symbol} [get]
func getCurrency(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		req := types.CurrencyReq{Symbol: vars["symbol"]}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/currency", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
