package handler

import (
	"github.com/ecommerce-platform/internal/user/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

type UserHandler struct {
	userService service.UserService	// 用户服务接口：注册、登录和获取用户信息
}

// NewUserHandler 新用户Handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	if err := h.userService.Register(c.Request.Context(), &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	resp, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	user, err := h.userService.GetUserInfo(c.Request.Context(), uint(userID))
	if err != nil {
		response.Fail(c, pkgerrors.ErrUserNotFound, err.Error())
		return
	}

	response.Success(c, user)
}
