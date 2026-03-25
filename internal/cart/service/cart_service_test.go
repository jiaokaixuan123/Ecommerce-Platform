package service_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/cart/domain"
	"github.com/ecommerce-platform/internal/cart/mocks"
	"github.com/ecommerce-platform/internal/cart/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newService(cartRepo *mocks.MockCartRepository, itemRepo *mocks.MockCartItemRepository) service.CartService {
	return service.NewCartService(cartRepo, itemRepo)
}

// ---- GetUserCart ----

// TestGetUserCart_Success 测试正常获取购物车（含商品项）
func TestGetUserCart_Success(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	items := []*domain.CartItem{
		{ID: 1, CartID: 1, ProductID: 101, ProductName: "手机", ProductPrice: 5000, Quantity: 2, Selected: true},
		{ID: 2, CartID: 1, ProductID: 102, ProductName: "耳机", ProductPrice: 500, Quantity: 1, Selected: false},
	}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)
	itemRepo.On("ListByCartID", ctx, uint(1)).Return(items, nil)

	resp, err := svc.GetUserCart(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(10500), resp.TotalPrice)     // 5000*2 + 500*1
	assert.Equal(t, int64(10000), resp.SelectedAmount) // 5000*2（只有第一项选中）
	assert.Equal(t, 2, resp.ItemCount)
	assert.Len(t, resp.Items, 2)
	cartRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
}

// TestGetUserCart_Empty 测试购物车为空
func TestGetUserCart_Empty(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{}
	items := []*domain.CartItem{}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(2)).Return(cart, nil)
	itemRepo.On("ListByCartID", ctx, uint(cart.ID)).Return(items, nil)
	cartVO, err := svc.GetUserCart(ctx, 2)
	assert.NoError(t, err)
	assert.Equal(t, cartVO.TotalPrice, int64(0))
	assert.Zero(t, cartVO.ItemCount)
}

// ---- AddCartItem ----

// TestAddCartItem_Success 测试正常添加商品到购物车
func TestAddCartItem_Success(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)
	itemRepo.On("Add", ctx, mock.MatchedBy(func(item *domain.CartItem) bool {
		return item.CartID == 1 && item.ProductID == 101 && item.Quantity == 2
	})).Return(nil)

	err := svc.AddCartItem(ctx, &service.AddCartItemReq{
		UserID:       1,
		ProductID:    101,
		Quantity:     2,
		ProductName:  "手机",
		ProductPrice: 5000,
	})

	assert.NoError(t, err)
	cartRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
}

// TestAddCartItem_InvalidQuantity 测试数量<=0 时返回参数错误
func TestAddCartItem_InvalidQuantity(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	err := svc.AddCartItem(ctx, &service.AddCartItemReq{Quantity: 0})
	assert.Equal(t, err.Error(), pkgerrors.Msg(pkgerrors.ErrParam))
	cartRepo.AssertNotCalled(t, "GetOrCreateByUserID")
	itemRepo.AssertNotCalled(t, "UpdateQuantity")
}

// ---- UpdateCartItemQuantity ----

// TestUpdateCartItemQuantity_Success 测试正常更新数量
func TestUpdateCartItemQuantity_Success(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	item := &domain.CartItem{ID: 10, CartID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)
	itemRepo.On("GetItemByID", ctx, uint(10)).Return(item, nil)
	itemRepo.On("UpdateQuantity", ctx, uint(10), 5).Return(nil)

	err := svc.UpdateCartItemQuantity(ctx, &service.UpdateCartItemQuantityReq{
		UserID:   1,
		ItemID:   10,
		Quantity: 5,
	})

	assert.NoError(t, err)
	cartRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
}

// TestUpdateCartItemQuantity_NotBelongToUser 测试操作不属于该用户的购物车项
func TestUpdateCartItemQuantity_NotBelongToUser(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)

	item := &domain.CartItem{ID: 10, CartID: 999}
	itemRepo.On("GetItemByID", ctx, uint(10)).Return(item, nil)

	err := svc.UpdateCartItemQuantity(ctx, &service.UpdateCartItemQuantityReq{
		UserID:   1,
		ItemID:   10,
		Quantity: 1,
	})

	assert.Equal(t, err.Error(), pkgerrors.Msg(pkgerrors.ErrForbidden)) // 无权限
	itemRepo.AssertNotCalled(t, "UpdateQuantity", ctx, uint(10), 1)
}

// ---- UpdateCartItemSelected ----

// TestUpdateCartItemSelected_Success 测试正常更新选中状态
func TestUpdateCartItemSelected_Success(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)

	item := &domain.CartItem{ID: 10, CartID: 1}
	itemRepo.On("GetItemByID", ctx, uint(10)).Return(item, nil)
	itemRepo.On("UpdateSelected", ctx, uint(10), true).Return(nil)

	err := svc.UpdateCartItemSelected(ctx, &service.UpdateCartItemSelectedReq{
		UserID:   1,
		ItemID:   10,
		Selected: true,
	})

	assert.NoError(t, err)
}

// ---- DeleteCartItem ----

// TestDeleteCartItem_ByItemID 测试按 itemID 删除
func TestDeleteCartItem_ByItemID(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	item := &domain.CartItem{ID: 10, CartID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)
	itemRepo.On("GetItemByID", ctx, uint(10)).Return(item, nil)
	itemRepo.On("Delete", ctx, uint(10)).Return(nil)

	err := svc.DeleteCartItem(ctx, &service.DeleteCartItemReq{
		UserID: 1,
		ItemID: 10,
	})

	assert.NoError(t, err)
	cartRepo.AssertExpectations(t)
	itemRepo.AssertExpectations(t)
}

// TestDeleteCartItem_ByProductID 测试按 productID 删除
func TestDeleteCartItem_ByProductID(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	item := &domain.CartItem{ID: 10, CartID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)
	itemRepo.On("GetItemByID", ctx, uint(10)).Return(item, nil)

	itemRepo.On("DeleteByCartIDAndProductID", ctx, uint(1), uint(101)).Return(nil)
	err := svc.DeleteCartItem(ctx, &service.DeleteCartItemReq{
		UserID:    1,
		ProductID: 101,
	})
	assert.NoError(t, err)
}

// ---- ClearCart ----

// TestClearCart_Success 测试正常清空购物车
func TestClearCart_Success(t *testing.T) {
	cartRepo := new(mocks.MockCartRepository)
	itemRepo := new(mocks.MockCartItemRepository)
	svc := newService(cartRepo, itemRepo)
	ctx := context.Background()

	cart := &domain.Cart{ID: 1, UserID: 1}
	item := &domain.CartItem{ID: 10, CartID: 1}
	cartRepo.On("GetOrCreateByUserID", ctx, uint(1)).Return(cart, nil)
	itemRepo.On("GetItemByID", ctx, uint(10)).Return(item, nil)

	itemRepo.On("ClearByCartID", ctx, uint(1)).Return(nil)
	err := svc.ClearCart(ctx, uint(1))
	assert.NoError(t, err)
}
