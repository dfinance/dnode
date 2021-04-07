package rest

import (
	"encoding/hex"
	"encoding/json"
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

type ExecuteScriptReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Compiled Move code
	MoveCode string `json:"move_code" yaml:"move_code" format:"HEX encoded byte code"`
	// Script arguments
	MoveArgs []string `json:"move_args" yaml:"move_args" example:"true"`
}

type PublishModuleReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Compiled Move code
	MoveCode string `json:"move_code" yaml:"move_code" format:"HEX encoded byte code"`
}

type LcsViewReq struct {
	// Resource address
	Account string `json:"address" format:"bech32/hex" example:"0x0000000000000000000000000000000000000001"`
	// Move formatted path (ModuleName::StructName, where ::StructName is optional)
	MovePath string `json:"move_path" example:"Block::BlockMetadata"`
	// LCS view JSON formatted request (refer to docs for specs)
	ViewRequest string `json:"view_request" example:"[ { \"name\": \"height\", \"type\": \"U64\" } ]"`
}

type LcsViewResp struct {
	Value string `json:"value" format:"JSON"`
}

// Registering routes for REST API.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/compile", types.ModuleName), compile(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/data/{%s}/{%s}", types.ModuleName, accountAddrName, vmPathName), getData(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/view", types.ModuleName), lcsView(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/tx/{%s}", types.ModuleName, txHash), getTxVMStatus(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/execute", types.ModuleName), executeScript(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/publish", types.ModuleName), deployModule(cliCtx)).Methods("PUT")
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

// LCSView godoc
// @Tags VM
// @Summary Get writeSet data from VM LCS string view
// @Description Get writeSet data LCS string view for {address}::{moduleName}::{structName} Move path"
// @ID vmGetData
// @Accept  json
// @Produce json
// @Param request body LcsViewReq true "View request"
// @Success 200 {object} VmRespLcsView
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/view [get]
func lcsView(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		req := LcsViewReq{}
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		address, err := helpers.ParseSdkAddressParam("address", req.Account, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		moduleName, structName := req.MovePath, ""
		moveSepCnt := strings.Count(moduleName, "::")
		if moveSepCnt > 1 {
			err := helpers.BuildError("move_path", moduleName, helpers.ParamTypeRestRequest, "none/one :: separator is supported")
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		if moveSepCnt == 1 {
			values := strings.Split(moduleName, "::")
			moduleName, structName = values[0], values[1]
		}

		var viewRequest types.ViewerRequest
		if err := json.Unmarshal([]byte(req.ViewRequest), &viewRequest); err != nil {
			err := helpers.BuildError("view_request", "", helpers.ParamTypeRestRequest, fmt.Sprintf("JSON unmarshal: %v", err))
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		bz, err := cliCtx.Codec.MarshalJSON(types.LcsViewReq{
			Address:     address,
			ModuleName:  moduleName,
			StructName:  structName,
			ViewRequest: viewRequest,
		})
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryLcsView), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp := LcsViewResp{Value: string(res)}

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

// GetIssue godoc
// @Tags VM
// @Summary Execute Move script
// @Description Get execute Move script stdTx object
// @ID vmExecuteScript
// @Accept  json
// @Produce json
// @Param request body ExecuteScriptReq true "Execute request"
// @Success 200 {object} VmRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/execute [put]
func executeScript(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		compilerAddr := viper.GetString(vm_client.FlagCompilerAddr)

		// parse inputs
		var req ExecuteScriptReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		_, code, err := helpers.ParseHexStringParam("move_code", req.MoveCode, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		typedArgs, err := vm_client.ExtractArguments(compilerAddr, code)
		if err != nil {
			retErr := helpers.BuildError(
				"move_args",
				strings.Join(req.MoveArgs, ", "),
				helpers.ParamTypeRestRequest,
				fmt.Sprintf("extracting typed args from the code: %v", err),
			)
			rest.WriteErrorResponse(w, http.StatusBadRequest, retErr.Error())
			return
		}

		scriptArgs, err := vm_client.ConvertStringScriptArguments(req.MoveArgs, typedArgs)
		if err != nil {
			retErr := helpers.BuildError(
				"move_args",
				strings.Join(req.MoveArgs, ", "),
				helpers.ParamTypeRestRequest,
				fmt.Sprintf("converting input args to typed args: %v", err),
			)
			rest.WriteErrorResponse(w, http.StatusBadRequest, retErr.Error())
			return
		}
		if len(scriptArgs) == 0 {
			scriptArgs = nil
		}

		// create the message
		msg := types.NewMsgExecuteScript(fromAddr, code, scriptArgs)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// GetIssue godoc
// @Tags VM
// @Summary Publish Move module
// @Description Get publish Move module stdTx object
// @ID vmDeployModule
// @Accept  json
// @Produce json
// @Param request body PublishModuleReq true "Publish request"
// @Success 200 {object} VmRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /vm/publish [put]
func deployModule(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req PublishModuleReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		_, code, err := helpers.ParseHexStringParam("move_code", req.MoveCode, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgDeployModule(fromAddr, code)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
