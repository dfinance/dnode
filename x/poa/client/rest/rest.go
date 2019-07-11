package rest


import (
    "github.com/cosmos/cosmos-sdk/client/context"
    "github.com/gorilla/mux"
    "fmt"
    "wings-blockchain/x/poa/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "net/http"
    "github.com/cosmos/cosmos-sdk/types/rest"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
    r.HandleFunc(fmt.Sprintf("/%s/validators", types.ModuleName), getValidators(cdc, cliCtx)).Methods("GET")
}

func getValidators(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validators", types.ModuleName), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}
