package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/dfinance/dnode/x/common_vm"
	vmClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	accountAddrName = "accountAddr"
	vmPathName      = "vmPath"
	txHash          = "txHash"
)

type compileReq struct {
	Code    string `json:"code"`                                                            // Script code
	Account string `json:"address" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"` // Code address
}

// Registering routes for REST API.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/compile-script", types.ModuleName), compileScript(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/compile-module", types.ModuleName), compileModule(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/data/{%s}/{%s}", types.ModuleName, accountAddrName, vmPathName), getData(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/tx/{%s}", types.ModuleName, txHash), getTxVMStatus(cliCtx)).Methods("GET")
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

		compilerAddr := viper.GetString(vmClient.FlagCompilerAddr)
		sourceFile := &vm_grpc.MvIrSourceFile{
			Text:    req.Code,
			Address: []byte(req.Account),
			Type:    compileType,
		}

		byteCode, err := vmClient.Compile(compilerAddr, sourceFile)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := vmClient.MoveFile{
			Code: hex.EncodeToString(byteCode),
		}
		rest.PostProcessResponse(w, cliCtx, resp)
	}
}

// GetData godoc
// @Tags vm
// @Summary Get data from data source
// @Description Get data from data source by accountAddr and path
// @ID vmGetData
// @Accept  json
// @Produce json
// @Param accountAddr path string true "account address (Libra HEX  Bech32)"
// @Param vmPath path string true "VM path (HEX string)"
// @Success 200 {object} VmData
// @Failure 422 {object} rest.ErrorResponse "Returned if the request doesn't have valid path params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/data/{accountAddr}/{vmPath} [get]
func getData(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		rawAddress := vars[accountAddrName]
		rawPath := vars[vmPathName]

		var address sdk.AccAddress
		address, err := hex.DecodeString(rawAddress)
		if err != nil {
			address, err = sdk.AccAddressFromBech32(rawAddress)
			if err != nil {
				rest.WriteErrorResponse(
					w,
					http.StatusUnprocessableEntity,
					fmt.Sprintf("can't parse address %q (should be libra hex or bech32): %v", rawAddress, err),
				)
				return
			}

			address = common_vm.Bech32ToLibra(address)
		}

		path, err := hex.DecodeString(rawPath)
		if err != nil {
			rest.WriteErrorResponse(
				w,
				http.StatusUnprocessableEntity,
				fmt.Sprintf("can't parse path %q: %v", rawPath, err),
			)
			return
		}
		if len(path) > 0 && path[0] != 0x0 {
			rest.WriteErrorResponse(
				w,
				http.StatusUnprocessableEntity,
				fmt.Sprintf("path %q: first byte must be 0x0", rawPath),
			)
			return
		}

		bz, err := cliCtx.Codec.MarshalJSON(types.QueryAccessPath{
			Address: address,
			Path:    path,
		})
		if err != nil {
			rest.WriteErrorResponse(
				w,
				http.StatusInternalServerError,
				fmt.Sprintf("can't marshal query: %v", err),
			)
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/value", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(
				w,
				http.StatusInternalServerError,
				fmt.Sprintf("processing query: %v", err),
			)
			return
		}
		resp := types.QueryValueResp{Value: hex.EncodeToString(res)}

		rest.PostProcessResponse(w, cliCtx, resp)
	}
}

// GetTxVMStatus godoc
// @Tags vm
// @Summary Get tx VM execution status
// @Description Get tx VM execution status by tx hash
// @ID vmTxStatus
// @Accept  json
// @Produce json
// @Param txHash path string true "transaction hash"
// @Success 200 {object} VmTxStatus
// @Failure 422 {object} rest.ErrorResponse "Returned if the request doesn't have valid path params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/tx/{txHash} [get]
func getTxVMStatus(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		txHash := vars[txHash]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		output, err := utils.QueryTx(cliCtx, txHash)
		if err != nil {
			if strings.Contains(err.Error(), txHash) {
				rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
				return
			}
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if output.Empty() {
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("no transaction found with hash %s", txHash))
			return
		}

		resp := types.NewVMStatusFromABCILogs(output)
		rest.PostProcessResponse(w, cliCtx, resp)
	}
}
