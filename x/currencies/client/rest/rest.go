package rest

import (
    "github.com/cosmos/cosmos-sdk/client/context"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/gorilla/mux"
    "fmt"
    "net/http"
    "github.com/cosmos/cosmos-sdk/types/rest"
    "wings-blockchain/x/currencies/types"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
    r.HandleFunc(fmt.Sprintf("/%s/issue/{issueID}", types.ModuleName), getIssue(cdc, cliCtx)).Methods("GET")
    r.HandleFunc(fmt.Sprintf("/%s/get/{symbol}", types.ModuleName), getCurrency(cdc, cliCtx)).Methods("GET")
}

func getIssue(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        issueID := vars["issueID"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/issue/%s", types.ModuleName, issueID), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}

func getCurrency(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars   := mux.Vars(r)
        symbol := vars["symbol"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get/%s", types.ModuleName, symbol), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}
