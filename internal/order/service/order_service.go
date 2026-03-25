package service

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce-platform/internal/order/domain"
	"github.com/ecommerce-platform/internal/order/repository"
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

// ---- 方法实现（TODO：由你来完成）----

// CreateOrder TODO Step 1
// 提示：
//   - 用 fmt.Sprintf("ORD%s%06d", time.Now().Format("20060102150405"), userID) 生成订单号
//   - 遍历 req.Items 计算 TotalAmount 和构建 OrderItem 列表
//   - 调用 s.orderRepo.Create(ctx, order, items)
func (s *orderService) CreateOrder(ctx context.Context, req *CreateOrderReq) (*domain.Order, error) {
	var totalAmount int64
	var orderItems []*domain.OrderItem

	orderNo := fmt.Sprintf("ORD%s%06d", time.Now().Format("20060102150405"), req.UserID)

	for _, item := range req.Items {
		totalAmount += int64(item.Price)
		orderItems = append(orderItems, &domain.OrderItem{
			ProductID:    item.ProductID,
			MerchantID:   item.MerchantID,
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			Price:        item.Price,
			Quantity:     item.Quantity,
		})
	}
	order := &domain.Order{
		UserID:      req.UserID,
		OrderNo:     orderNo,
		TotalAmount: int64(totalAmount),
		Status:      domain.OrderStatus(domain.OrderStatusPending),
		Remark:      "",
	}

	return order, s.orderRepo.Create(ctx, order, orderItems)
}

// GetOrderDetail TODO Step 2
// 提示：
//   - GetByID 查询订单，验证 order.UserID == userID（防止越权）
//   - ListByOrderID 查询商品项
//   - 组装 OrderDetailResp 返回
func (s *orderService) GetOrderDetail(ctx context.Context, userID, orderID uint) (*OrderDetailResp, error) {
	// TODO
	panic("not implemented")
}

// ListUserOrders TODO Step 3
// 提示：
//   - page 从 1 开始，offset = (page-1) * pageSize
//   - pageSize 默认 10，最大 50
func (s *orderService) ListUserOrders(ctx context.Context, userID uint, page, pageSize int) (*ListOrderResp, error) {
	// TODO
	panic("not implemented")
}

// CancelOrder TODO Step 4
// 提示：
//   - 先 GetByID，验证 UserID 一致
//   - 调用 UpdateStatus(ctx, id, OrderStatusPending, OrderStatusCancelled)
//   - UpdateStatus 返回错误时说明状态已不是待支付，返回 ErrOrderStatusInvalid
func (s *orderService) CancelOrder(ctx context.Context, userID, orderID uint) error {
	// TODO
	panic("not implemented")
}

// ConfirmOrder TODO Step 5（支付服务回调用）
// 提示：
//   - 调用 UpdateStatus(ctx, id, OrderStatusPending, OrderStatusPaid)
func (s *orderService) ConfirmOrder(ctx context.Context, orderID uint) error {
	// TODO
	panic("not implemented")
}
