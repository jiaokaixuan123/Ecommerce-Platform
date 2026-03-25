package handler

import (
	"strconv"

	"github.com/ecommerce-platform/internal/cart/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(svc service.CartService) *CartHandler {
	return &CartHandler{cartService: svc}
}

func (h *CartHandler) GetUserCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	cart, err := h.cartService.GetUserCart(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, pkgerrors.ErrNotFound, err.Error())
		return
	}

	response.Success(c, cart)
}

func (h *CartHandler) AddCartItem(c *gin.Context) {
	var req service.AddCartItemReq
	// 从 body 取出商品信息
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	// 从 jwt 取 UserID
	req.UserID = middleware.GetUserID(c)	
	if err := h.cartService.AddCartItem(c.Request.Context(), &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *CartHandler) UpdateCartItemQuantity(c *gin.Context) {
	var req service.UpdateCartItemQuantityReq

	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	req.ItemID = uint(itemID)

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	req.UserID = middleware.GetUserID(c)
	if err := h.cartService.UpdateCartItemQuantity(c.Request.Context(), &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *CartHandler) UpdateCartItemSelected(c *gin.Context) {
	var req service.UpdateCartItemSelectedReq

	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	req.ItemID = uint(itemID)

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}
	req.UserID = middleware.GetUserID(c)
	if err := h.cartService.UpdateCartItemSelected(c.Request.Context(), &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *CartHandler) DeleteCartItem(c *gin.Context) {
	itemID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
    productID, _ := strconv.ParseUint(c.Query("product_id"), 10, 64)

    req := &service.DeleteCartItemReq{
        UserID:    middleware.GetUserID(c),
        ItemID:    uint(itemID),
        ProductID: uint(productID),
    }
	if err := h.cartService.DeleteCartItem(c.Request.Context(), req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *CartHandler) ClearCart(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.cartService.ClearCart(c.Request.Context(), userID); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}
