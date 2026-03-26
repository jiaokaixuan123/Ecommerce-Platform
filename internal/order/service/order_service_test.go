package service_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/order/domain"
	"github.com/ecommerce-platform/internal/order/mocks"
	"github.com/ecommerce-platform/internal/order/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newService(orderRepo *mocks.MockOrderRepository, itemRepo *mocks.MockOrderItemRepository) service.OrderService {
	return service.NewOrderService(orderRepo, itemRepo)
}

// ---- CreateOrder ----

func TestCreateOrder_Success(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	orderRepo.On("Create", ctx, mock.MatchedBy(func(o *domain.Order) bool {
		return o.UserID == 1 && o.TotalAmount == 20000 // 2件 x 10000分
	}), mock.Anything).Return(nil)

	resp, err := svc.CreateOrder(ctx, &service.CreateOrderReq{
		UserID: 1,
		Items: []*service.CreateOrderItem{
			{ProductID: 101, MerchantID: 1, ProductName: "手机", Price: 10000, Quantity: 2},
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(20000), resp.TotalAmount)
	assert.Equal(t, domain.OrderStatusPending, resp.Status)
	orderRepo.AssertExpectations(t)
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	_, err := svc.CreateOrder(ctx, &service.CreateOrderReq{
		UserID: 1,
		Items:  []*service.CreateOrderItem{},
	})

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrParam))
	orderRepo.AssertNotCalled(t, "Create")
}

func TestCreateOrder_RepoError(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	orderRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(assert.AnError)

	_, err := svc.CreateOrder(ctx, &service.CreateOrderReq{
		UserID: 1,
		Items:  []*service.CreateOrderItem{
			{ProductID: 101, MerchantID: 1, ProductName: "手机", Price: 5000, Quantity: 1},
		},
	})

	assert.Error(t, err)
}

// ---- GetOrderDetail ----

func TestGetOrderDetail_Success(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	order := &domain.Order{ID: 1, UserID: 1, TotalAmount: 5000, Status: domain.OrderStatusPending}
	items := []*domain.OrderItem{{ID: 1, OrderID: 1, ProductName: "手机", Price: 5000, Quantity: 1}}

	orderRepo.On("GetByID", ctx, uint(1)).Return(order, nil)
	itemRepo.On("ListByOrderID", ctx, uint(1)).Return(items, nil)

	resp, err := svc.GetOrderDetail(ctx, 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), resp.ID)
	assert.Len(t, resp.Items, 1)
	orderRepo.AssertExpectations(t)
}

func TestGetOrderDetail_NotFound(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	orderRepo.On("GetByID", ctx, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	_, err := svc.GetOrderDetail(ctx, 1, 99)

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrOrderNotFound))
}

func TestGetOrderDetail_Forbidden(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	// 订单属于 userID=2，但 userID=1 来查
	orderRepo.On("GetByID", ctx, uint(1)).Return(
		&domain.Order{ID: 1, UserID: 2}, nil,
	)

	_, err := svc.GetOrderDetail(ctx, 1, 1)

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrForbidden))
	itemRepo.AssertNotCalled(t, "ListByOrderID")
}

// ---- CancelOrder ----

func TestCancelOrder_Success(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	orderRepo.On("GetByID", ctx, uint(1)).Return(
		&domain.Order{ID: 1, UserID: 1, Status: domain.OrderStatusPending}, nil,
	)
	orderRepo.On("UpdateStatus", ctx, uint(1), domain.OrderStatusPending, domain.OrderStatusCancelled).Return(nil)

	err := svc.CancelOrder(ctx, 1, 1)

	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
}

func TestCancelOrder_AlreadyPaid(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	// 已支付状态不允许取消
	orderRepo.On("GetByID", ctx, uint(1)).Return(
		&domain.Order{ID: 1, UserID: 1, Status: domain.OrderStatusPaid}, nil,
	)

	err := svc.CancelOrder(ctx, 1, 1)

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrOrderStatusInvalid))
	orderRepo.AssertNotCalled(t, "UpdateStatus")
}

// ---- ListUserOrders ----

func TestListUserOrders_Success(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	orders := []*domain.Order{
		{ID: 1, UserID: 1, TotalAmount: 5000},
		{ID: 2, UserID: 1, TotalAmount: 3000},
	}
	orderRepo.On("ListByUserID", ctx, uint(1), 0, 10).Return(orders, int64(2), nil)

	resp, err := svc.ListUserOrders(ctx, 1, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.Orders, 2)
}

// ---- ConfirmOrder ----

func TestConfirmOrder_Success(t *testing.T) {
	orderRepo := new(mocks.MockOrderRepository)
	itemRepo := new(mocks.MockOrderItemRepository)
	svc := newService(orderRepo, itemRepo)
	ctx := context.Background()

	orderRepo.On("UpdateStatus", ctx, uint(1), domain.OrderStatusPending, domain.OrderStatusPaid).Return(nil)

	err := svc.ConfirmOrder(ctx, 1)

	assert.NoError(t, err)
	orderRepo.AssertExpectations(t)
}
