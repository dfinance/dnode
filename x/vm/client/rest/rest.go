package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/gorilla/mux"

	vmClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

type compileReq struct {
	Code        string         `json:"code"`                                                                                     // Script code
	Account     sdk.AccAddress `json:"address" swaggertype:"string" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`     // Sender address
	CompilerUrl string         `json:"compiler_url" extensions:"x-nullable" default:"127.0.0.1:50053" example:"127.0.0.1:50053"` // VM compiler URL (optional)
}

// Registering routes for REST API.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/compile-script", types.ModuleName), compileScript(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/compile-module", types.ModuleName), compileModule(cliCtx)).Methods("GET")
}

// GetCompiledScript godoc
// @Tags vm
// @Summary Get compiled script
// @Description Compile script code using VM and return byteCode
// @ID vmGetCompiledScript
// @Accept  json
// @Produce json
// @Param getRequest body compileReq true "Code with metadata"
// @Success 200 {object} VmRespCompile
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/compile-script [get]
func compileScript(cliCtx context.CLIContext) http.HandlerFunc {
	return commonCompileHandler(cliCtx, vm_grpc.ContractType_Script)
}

// GetCompiledModule godoc
// @Tags vm
// @Summary Get compiled module
// @Description Compile module code using VM and return byteCode
// @ID vmGetCompiledModule
// @Accept  json
// @Produce json
// @Param getRequest body compileReq true "Code with metadata"
// @Success 200 {object} VmRespCompile
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/compile-module [get]
func compileModule(cliCtx context.CLIContext) http.HandlerFunc {
	return commonCompileHandler(cliCtx, vm_grpc.ContractType_Module)
}

func commonCompileHandler(cliCtx context.CLIContext, compileType vm_grpc.ContractType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := compileReq{}
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		if req.CompilerUrl == "" {
			req.CompilerUrl = vmClient.DefaultCompilerAddr
		}

		sourceFile := &vm_grpc.MvIrSourceFile{
			Text:    req.Code,
			Address: []byte(req.Account.String()),
			Type:    compileType,
		}

		byteCode, err := vmClient.Compile(req.CompilerUrl, sourceFile)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := vmClient.MVFile{
			Code: hex.EncodeToString(byteCode),
		}
		rest.PostProcessResponse(w, cliCtx, resp)
	}
}
