package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ecommerce-platform/internal/order/domain"
	"github.com/ecommerce-platform/internal/order/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

// ---- 请求/响应结构体 ----

// CreateOrderReq 创建订单请求
// 来源：用户勾选购物车中的商品后提交
type CreateOrderReq struct {
	UserID uint               `json:"user_id"` // 从登录态获取
	Items  []*CreateOrderItem `json:"items" binding:"required,min=1"`
	Remark string             `json:"remark"`
}

type CreateOrderItem struct {
	ProductID    uint   `json:"product_id" binding:"required"`
	MerchantID   uint   `json:"merchant_id" binding:"required"`
	ProductName  string `json:"product_name" binding:"required"`
	ProductImage string `json:"product_image"`
	Price        int64  `json:"price" binding:"required"` // 前端传入（需与商品服务校验）
	Quantity     int    `json:"quantity" binding:"required,min=1"`
}

// OrderDetailResp 订单详情响应
type OrderDetailResp struct {
	*domain.Order
	Items []*domain.OrderItem `json:"items"`
}

// ListOrderResp 订单列表响应
type ListOrderResp struct {
	Orders []*domain.Order `json:"orders"`
	Total  int64           `json:"total"`
}

// ---- 接口定义 ----

type OrderService interface {
	// CreateOrder 创建订单
	// 流程：生成订单号 → 计算总金额 → 扣减库存 → 写入订单 → 清除已购购物车项
	CreateOrder(ctx context.Context, req *CreateOrderReq) (*domain.Order, error)

	// GetOrderDetail 获取订单详情（含商品项）
	GetOrderDetail(ctx context.Context, userID, orderID uint) (*OrderDetailResp, error)

	// ListUserOrders 分页查询用户订单列表
	ListUserOrders(ctx context.Context, userID uint, page, pageSize int) (*ListOrderResp, error)

	// CancelOrder 取消订单（只允许取消待支付状态的订单）
	CancelOrder(ctx context.Context, userID, orderID uint) error

	// ConfirmOrder 确认订单已支付（由支付服务回调触发，将状态从待支付→已支付）
	ConfirmOrder(ctx context.Context, orderID uint) error
}

// ---- 结构体和构造函数 ----

type orderService struct {
	orderRepo repository.OrderRepository
	itemRepo  repository.OrderItemRepository
}

func NewOrderService(orderRepo repository.OrderRepository, itemRepo repository.OrderItemRepository) OrderService {
	return &orderService{
		orderRepo: orderRepo,
		itemRepo:  itemRepo,
	}
}

// 实现
func (s *orderService) CreateOrder(ctx context.Context, req *CreateOrderReq) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrParam))
	}

	var totalAmount int64
	var orderItems []*domain.OrderItem

	orderNo := fmt.Sprintf("ORD%s%06d", time.Now().Format("20060102150405"), req.UserID)

	for _, item := range req.Items {
		totalAmount += int64(item.Price) * int64(item.Quantity)
		orderItems = append(orderItems, &domain.OrderItem{
			ProductID:    item.ProductID,
			MerchantID:   item.MerchantID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			Price:        item.Price,
			Quantity:     item.Quantity,
			Subtotal: 	  item.Price * int64(item.Quantity),
		})
	}
	order := &domain.Order{
		UserID:      req.UserID,
		OrderNo:     orderNo,
		TotalAmount: int64(totalAmount),
		Status:      domain.OrderStatusPending,
		Remark:      "",
	}

	return order, s.orderRepo.Create(ctx, order, orderItems)
}

func (s *orderService) GetOrderDetail(ctx context.Context, userID, orderID uint) (*OrderDetailResp, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrOrderNotFound))
		}
		return nil, err
	}
	if order.UserID != userID {
		return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrForbidden))
	}

	items, err := s.itemRepo.ListByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return &OrderDetailResp{
		Order: order,
		Items: items,
	}, nil
}

func (s *orderService) ListUserOrders(ctx context.Context, userID uint, page, pageSize int) (*ListOrderResp, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	offset := (page - 1) * pageSize
	orders, total, err := s.orderRepo.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &ListOrderResp{
		Orders: orders,
		Total: total,
	}, nil
}

func (s *orderService) CancelOrder(ctx context.Context, userID, orderID uint) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(pkgerrors.Msg(pkgerrors.ErrOrderNotFound))
		}
		return err
	}
	if order.UserID != userID {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrForbidden))
	}
	if order.Status != domain.OrderStatusPending {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrOrderStatusInvalid))
	}
	return s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPending, domain.OrderStatusCancelled)
}

func (s *orderService) ConfirmOrder(ctx context.Context, orderID uint) error {
	return s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusPending, domain.OrderStatusPaid)
}
