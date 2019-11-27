package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
	"wings-blockchain/x/currencies/types"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	r.HandleFunc(fmt.Sprintf("/%s/issue/{issueID}", types.ModuleName), getIssue(cdc, cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currency/{symbol}", types.ModuleName), getCurrency(cdc, cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/destroy/{destroyID}", types.ModuleName), getDestroy(cdc, cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/destroys/{page}", types.ModuleName), getDestroys(cdc, cliCtx)).Methods("GET")
}

func getDestroys(cdc *codec.Codec, cliContext context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		page := vars["page"]
		limit := r.URL.Query().Get("limit")

		if limit == "" {
			limit = "100"
		}

		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/destroys/%s/%s", types.ModuleName, page, limit), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliContext, res)
	}
}

func getDestroy(cdc *codec.Codec, cliContext context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		destroyID := vars["destroyID"]

		res, _, err := cliContext.QueryWithData(fmt.Sprintf("custom/%s/destroy/%s", types.ModuleName, destroyID), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliContext, res)
	}
}

func getIssue(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		issueID := vars["issueID"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/issue/%s", types.ModuleName, issueID), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCurrency(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		symbol := vars["symbol"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/currency/%s", types.ModuleName, symbol), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
