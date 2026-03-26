package handler

import (
	"strconv"

	"github.com/ecommerce-platform/internal/seckill/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/ecommerce-platform/pkg/response"
	"github.com/gin-gonic/gin"
)

type SeckillHandler struct {
	seckillService service.SeckillService
}

func NewSeckillHandler(svc service.SeckillService) *SeckillHandler {
	return &SeckillHandler{seckillService: svc}
}

// DoSeckill POST /api/v1/seckill/:id
// 用户发起秒杀
func (h *SeckillHandler) DoSeckill(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	req := &service.DoSeckillReq{
		SeckillID: uint(id),
		UserID:    middleware.GetUserID(c),
	}

	if err := h.seckillService.DoSeckill(c.Request.Context(), req); err != nil {
		response.Fail(c, pkgerrors.ErrSeckillOver, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "排队中，请稍后查询订单"})
}

// GetSeckillProduct GET /api/v1/seckill/:id
// 查询秒杀商品详情
func (h *SeckillHandler) GetSeckillProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	sp, err := h.seckillService.GetSeckillProduct(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, pkgerrors.ErrNotFound, err.Error())
		return
	}

	response.Success(c, sp)
}

// CreateSeckillProduct POST /api/v1/seckill/admin
// 管理员创建秒杀活动
func (h *SeckillHandler) CreateSeckillProduct(c *gin.Context) {
	var req service.CreateSeckillReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	sp, err := h.seckillService.CreateSeckillProduct(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, sp)
}

// PrewarmStock POST /api/v1/seckill/admin/:id/prewarm
// 手动预热库存到 Redis
func (h *SeckillHandler) PrewarmStock(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, pkgerrors.ErrParam, err.Error())
		return
	}

	if err := h.seckillService.PrewarmStock(c.Request.Context(), uint(id)); err != nil {
		response.Fail(c, pkgerrors.ErrServer, err.Error())
		return
	}

	response.Success(c, gin.H{"message": "预热成功"})
}
