package handler

// Handler 层 / 控制器
// 接收 HTTP 请求，解析参数，调用 Service 层逻辑，统一返回响应（成功 / 失败）

import (
	"github.com/ecommerce-platform/internal/user/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
	"strconv"
)

// UserHandler：封装用户服务
type UserHandler struct {
	userService service.UserService	// 用户服务接口：注册、登录和获取用户信息
}

// NewUserHandler：创建 Handler 实体
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// 实现 UserService 接口

// Register：实现注册
func (h *UserHandler) Register(c *gin.Context) {
	// 解析注册请求参数
	var req service.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	// 调用 Service 注册逻辑
	if err := h.userService.Register(c.Request.Context(), &req); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	// 返回响应
	response.Success(c, nil)
}

// Login：实现登录
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

// GetUserInfo：实现获取用户信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	user, err := h.userService.GetUserInfo(c.Request.Context(), uint(userID))
	if err != nil {
		response.Fail(c, pkgerrors.ErrUserNotFound, err.Error())
		return
	}

	response.Success(c, user)
}
