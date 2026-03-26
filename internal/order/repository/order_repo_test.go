package repository_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/order/domain"
	"github.com/ecommerce-platform/internal/order/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Order{}, &domain.OrderItem{}))
	return db
}

func newRepos(db *gorm.DB) (repository.OrderRepository, repository.OrderItemRepository) {
	return repository.NewOrderRepository(db), repository.NewOrderItemRepository(db)
}

func sampleOrder(userID uint) *domain.Order {
	return &domain.Order{
		OrderNo:     "ORD20240101000001",
		UserID:      userID,
		TotalAmount: 10000,
		Status:      domain.OrderStatusPending,
	}
}

func sampleItems(orderID uint) []*domain.OrderItem {
	return []*domain.OrderItem{
		{OrderID: orderID, ProductID: 1, MerchantID: 1, ProductName: "手机", Price: 5000, Quantity: 2, Subtotal: 10000},
	}
}

// TestCreate_And_GetByID 测试创建订单并按 ID 查询
func TestCreate_And_GetByID(t *testing.T) {
	db := setupDB(t)
	orderRepo, _ := newRepos(db)
	ctx := context.Background()

	order := sampleOrder(1)
	items := []*domain.OrderItem{
		{ProductID: 1, MerchantID: 1, ProductName: "手机", Price: 5000, Quantity: 2, Subtotal: 10000},
	}

	require.NoError(t, orderRepo.Create(ctx, order, items))
	assert.NotZero(t, order.ID)

	got, err := orderRepo.GetByID(ctx, order.ID)
	require.NoError(t, err)
	assert.Equal(t, order.OrderNo, got.OrderNo)
	assert.Equal(t, int64(10000), got.TotalAmount)
}

// TestGetByID_NotFound 测试查询不存在的订单
func TestGetByID_NotFound(t *testing.T) {
	db := setupDB(t)
	orderRepo, _ := newRepos(db)
	ctx := context.Background()

	_, err := orderRepo.GetByID(ctx, 999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestGetByOrderNo 测试按业务订单号查询
func TestGetByOrderNo(t *testing.T) {
	db := setupDB(t)
	orderRepo, _ := newRepos(db)
	ctx := context.Background()

	order := sampleOrder(1)
	require.NoError(t, orderRepo.Create(ctx, order, nil))

	got, err := orderRepo.GetByOrderNo(ctx, order.OrderNo)
	require.NoError(t, err)
	assert.Equal(t, order.ID, got.ID)
}

// TestListByUserID 测试分页查询用户订单
func TestListByUserID(t *testing.T) {
	db := setupDB(t)
	orderRepo, _ := newRepos(db)
	ctx := context.Background()

	// 创建 3 个订单，2 个属于 userID=1，1 个属于 userID=2
	for i, no := range []string{"ORD001", "ORD002"} {
		o := &domain.Order{OrderNo: no, UserID: 1, TotalAmount: int64((i + 1) * 1000), Status: domain.OrderStatusPending}
		require.NoError(t, orderRepo.Create(ctx, o, nil))
	}
	other := &domain.Order{OrderNo: "ORD003", UserID: 2, TotalAmount: 500, Status: domain.OrderStatusPending}
	require.NoError(t, orderRepo.Create(ctx, other, nil))

	orders, total, err := orderRepo.ListByUserID(ctx, 1, 0, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, orders, 2)
}

// TestUpdateStatus_Success 测试 CAS 状态更新成功
func TestUpdateStatus_Success(t *testing.T) {
	db := setupDB(t)
	orderRepo, _ := newRepos(db)
	ctx := context.Background()

	order := sampleOrder(1)
	require.NoError(t, orderRepo.Create(ctx, order, nil))

	err := orderRepo.UpdateStatus(ctx, order.ID, domain.OrderStatusPending, domain.OrderStatusPaid)
	require.NoError(t, err)

	got, _ := orderRepo.GetByID(ctx, order.ID)
	assert.Equal(t, domain.OrderStatusPaid, got.Status)
}

// TestUpdateStatus_CASFail 测试 CAS 状态不匹配时返回错误
func TestUpdateStatus_CASFail(t *testing.T) {
	db := setupDB(t)
	orderRepo, _ := newRepos(db)
	ctx := context.Background()

	order := sampleOrder(1)
	require.NoError(t, orderRepo.Create(ctx, order, nil))

	// 当前状态是 Pending，用 Paid 作为 from 必然失败
	err := orderRepo.UpdateStatus(ctx, order.ID, domain.OrderStatusPaid, domain.OrderStatusCompleted)
	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrOrderStatusInvalid))
}

// TestListByOrderID 测试查询订单下的商品项
func TestListByOrderID(t *testing.T) {
	db := setupDB(t)
	orderRepo, itemRepo := newRepos(db)
	ctx := context.Background()

	order := sampleOrder(1)
	items := sampleItems(0) // orderID 在 Create 内部赋值
	require.NoError(t, orderRepo.Create(ctx, order, items))

	got, err := itemRepo.ListByOrderID(ctx, order.ID)
	require.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, "手机", got[0].ProductName)
	assert.Equal(t, int64(10000), got[0].Subtotal)
}
