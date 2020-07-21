package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	accountAddrName = "accountAddr"
	vmPathName      = "vmPath"
	txHash          = "txHash"
)

type CompileReq struct {
	// Script source code
	Code string `json:"code"`
	// Account address
	Account string `json:"address" format:"bech32/hex" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

// Registering routes for REST API.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/compile", types.ModuleName), compile(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/data/{%s}/{%s}", types.ModuleName, accountAddrName, vmPathName), getData(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/tx/{%s}", types.ModuleName, txHash), getTxVMStatus(cliCtx)).Methods("GET")
}

// Compile godoc
// @Tags VM
// @Summary Get compiled byteCode
// @Description Compile script / module code using VM and return byteCode
// @ID vmCompile
// @Accept  json
// @Produce json
// @Param getRequest body CompileReq true "Code with metadata"
// @Success 200 {object} VmRespCompile
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/compile [get]
func compile(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		compilerAddr := viper.GetString(vm_client.FlagCompilerAddr)

		// parse inputs and prepare request
		req := CompileReq{}
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		address, err := helpers.ParseSdkAddressParam("address", req.Account, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		sourceFile := &vm_grpc.SourceFile{
			Text:    req.Code,
			Address: common_vm.Bech32ToLibra(address),
		}

		// compile and process response
		byteCode, err := vm_client.Compile(compilerAddr, sourceFile)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		resp := vm_client.MoveFile{
			Code: hex.EncodeToString(byteCode),
		}

		rest.PostProcessResponse(w, cliCtx, resp)
	}
}

// GetData godoc
// @Tags VM
// @Summary Get writeSet data from VM
// @Description Get writeSet data from VM by accountAddr and path
// @ID vmGetData
// @Accept  json
// @Produce json
// @Param accountAddr path string true "account address (Libra HEX  Bech32)"
// @Param vmPath path string true "VM path (HEX string)"
// @Success 200 {object} VmData
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/data/{accountAddr}/{vmPath} [get]
func getData(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)

		address, err := helpers.ParseSdkAddressParam(accountAddrName, vars[accountAddrName], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		_, path, err := helpers.ParseHexStringParam(vmPathName, vars[vmPathName], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		bz, err := cliCtx.Codec.MarshalJSON(types.ValueReq{
			Address: common_vm.Bech32ToLibra(address),
			Path:    path,
		})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryValue), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := types.ValueResp{Value: hex.EncodeToString(res)}

		rest.PostProcessResponse(w, cliCtx, resp)
	}
}

// GetTxVMStatus godoc
// @Tags VM
// @Summary Get TX VM execution status
// @Description Get TX VM execution status by hash
// @ID vmTxStatus
// @Accept  json
// @Produce json
// @Param txHash path string true "transaction hash"
// @Success 200 {object} VmTxStatus
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 404 {object} rest.ErrorResponse "Returned if the requested data wasn't found"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/tx/{txHash} [get]
func getTxVMStatus(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
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
			rest.WriteErrorResponse(w, http.StatusNotFound, fmt.Sprintf("transaction with %q hash: not found", txHash))
			return
		}

		resp := types.NewVMStatusFromABCILogs(output)
		rest.PostProcessResponse(w, cliCtx, resp)
	}
}
