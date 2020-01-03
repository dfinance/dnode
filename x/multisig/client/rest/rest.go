package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"net/http"
	"wings-blockchain/x/multisig/types"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/call/{id}", types.ModuleName), getCall(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/calls", types.ModuleName), getCalls(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unique/{unique}", types.ModuleName), getCallByUnique(cliCtx)).Methods("GET")
}

func getCallByUnique(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["unique"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/unique/%s", types.ModuleName, id), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCall(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/call/%s", types.ModuleName, id), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCalls(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		mode := vars["mode"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/calls/%s", types.ModuleName, mode), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
