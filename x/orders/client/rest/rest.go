package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

const (
	OrderID        = "orderID"
	OrderOwner     = "owner"
	OrderDirection = "direction"
	OrderMarketID  = "marketID"
)

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s", types.ModuleName), getOrdersWithParams(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/{%s}", types.ModuleName, OrderID), getOrder(cliCtx)).Methods("GET")
}

// GetOrdersWithParams godoc
// @Tags orders
// @Summary Get orders
// @Description Get array of Order objects with pagination and filters
// @ID ordersGetOrdersWithParams
// @Accept  multipart/form-data
// @Produce json
// @Param page formData int false "page number (first page: 1)"
// @Param limit formData int false "items per page (default: 100)"
// @Param owner query string false "owner filter"
// @Param direction query string false "direction filter (bid/ask)"
// @Param marketID query string false "marketID filter (bid/ask)"
// @Success 200 {object} OrdersRespGetOrders
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders [get]
func getOrdersWithParams(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		ownerFilterStr := r.URL.Query().Get(OrderOwner)
		directionFilterStr := r.URL.Query().Get(OrderDirection)
		marketIDFitler := r.URL.Query().Get(OrderMarketID)

		ownerFilter := sdk.AccAddress{}
		if ownerFilterStr != "" {
			var err error
			ownerFilter, err = sdk.AccAddressFromBech32(ownerFilterStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%s param parsing: %v", OrderOwner, err))
				return
			}
		}

		// prepare request
		req := types.OrdersReq{
			Page:      page,
			Limit:     limit,
			Owner:     ownerFilter,
			Direction: types.NewDirectionRaw(directionFilterStr),
			MarketID:  marketIDFitler,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query and parse the result
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/list", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetOrder godoc
// @Tags orders
// @Summary Get order
// @Description Get Order object by orderID
// @ID ordersGetOrder
// @Accept  json
// @Produce json
// @Param orderID path string true "orderID"
// @Success 200 {object} OrdersRespGetOrder
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders/{orderID} [get]
func getOrder(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		vars := mux.Vars(r)
		id, err := dnTypes.NewIDFromString(vars[OrderID])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%q param parsing: %v", OrderID, err))
			return
		}

		// prepare request
		req := types.OrderReq{
			ID: id,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query and parse the result
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/order", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
