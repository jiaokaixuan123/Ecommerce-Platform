package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ecommerce-platform/internal/product/domain"
	"github.com/ecommerce-platform/internal/product/mocks"
	"github.com/ecommerce-platform/internal/product/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func newService(repo *mocks.MockProductRepository) service.ProductService {
	return service.NewProductService(repo)
}

// ---- CreateProduct ----

// TestCreateProduct_Success 测试正常创建商品
func TestCreateProduct_Success(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("Create", ctx, mock.MatchedBy(func(p *domain.Product) bool {
		return p.MerchantID == 1 && p.Name == "手机"
	})).Return(nil)

	req := &service.CreateProductReq{
		Name:        "手机",
		MerchantID:  1,
		Description: "智能电子产品",
		Price:       1000,
		Stock:       1,
		CategoryID:  1,
	}

	err := svc.CreateProduct(ctx, req)

	assert.NoError(t, err)
	repo.AssertNumberOfCalls(t, "Create", 1)
	repo.AssertExpectations(t)
}

// TestCreateProduct_RepoError 测试 repo 返回错误时 service 正确透传
func TestCreateProduct_RepoError(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	expError := errors.New("db error")
	repo.On("Create", ctx, mock.AnythingOfType("*domain.Product")).Return(expError)

	req := &service.CreateProductReq{
		Name:        "电脑",
		MerchantID:  2,
		Description: "智能电子产品",
		Price:       4000,
		Stock:       1,
		CategoryID:  2,
	}

	err := svc.CreateProduct(ctx, req)
	assert.Error(t, err)
	assert.Equal(t, expError, err)
	repo.AssertExpectations(t)
}

// ---- GetProduct ----

// TestGetProduct_Success 测试正常获取商品
func TestGetProduct_Success(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	expectedProduct := &domain.Product{
		ID:         1,
		Name:       "耳机",
		Price:      199,
		Stock:      200,
		CategoryID: 3,
		MerchantID: 1,
		Status:     1,
	}
	repo.On("GetByID", ctx, uint(1)).Return(expectedProduct, nil)

	product, err := svc.GetProduct(ctx, expectedProduct.ID)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), product.ID)
	assert.Equal(t, "耳机", product.Name)
	assert.Equal(t, int64(199), product.Price)
	repo.AssertExpectations(t)
}

// TestGetProduct_NotFound 测试商品不存在时返回业务错误
func TestGetProduct_NotFound(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(99)).Return(nil, gorm.ErrRecordNotFound)
	product, err := svc.GetProduct(ctx, 99)

	assert.Error(t, err)
	assert.Equal(t, pkgerrors.Msg(pkgerrors.ErrProductNotFound), err.Error())
	assert.Nil(t, product)
	repo.AssertExpectations(t)
}

// ---- ListProducts ----

// TestListProducts_DefaultPagination 测试默认分页参数（page=0 时自动修正为 1）
func TestListProducts_DefaultPagination(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("List", ctx, 0, 20, uint(0)).Return([]*domain.Product{}, int64(0), nil)

	req := &service.ListProductReq{}
	resp, err := svc.ListProducts(ctx, req)

	assert.NoError(t, err)
	assert.Len(t, resp.Products, 0)
	assert.Equal(t, int64(0), resp.Total)
	// 验证 repo.List 被正确调用
	repo.AssertCalled(t, "List", ctx, 0, 20, uint(0))
	repo.AssertExpectations(t)
}

// TestListProducts_WithCategory 测试按分类过滤
func TestListProducts_WithCategory(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	products := []*domain.Product{
		{ID: 1, CategoryID: 2},
		{ID: 2, CategoryID: 2},
	}

	repo.On("List", ctx, 0, 20, uint(2)).Return(products, int64(2), nil)

	resp, err := svc.ListProducts(ctx, &service.ListProductReq{CategoryID: 2})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
	assert.Len(t, resp.Products, 2)
	repo.AssertExpectations(t)
}

// ---- UpdateProduct ----

// TestUpdateProduct_Success 测试正常更新商品
func TestUpdateProduct_Success(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	existing := &domain.Product{ID: 1, Name: "旧名称", Price: 1000}
	repo.On("GetByID", ctx, uint(1)).Return(existing, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	newName := "新名称"
	newPrice := int64(2000)
	err := svc.UpdateProduct(ctx, 1, &service.UpdateProductReq{
		Name:  &newName,
		Price: &newPrice,
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// TestUpdateProduct_NotFound 测试更新不存在的商品
func TestUpdateProduct_NotFound(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(99)).Return(nil, gorm.ErrRecordNotFound)

	err := svc.UpdateProduct(ctx, 99, &service.UpdateProductReq{})

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrProductNotFound))
	repo.AssertNotCalled(t, "Update")
}

// ---- DeleteProduct ----

// TestDeleteProduct_Success 测试正常删除商品
func TestDeleteProduct_Success(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("GetByID", ctx, uint(1)).Return(&domain.Product{ID: 1}, nil)
	repo.On("Delete", ctx, uint(1)).Return(nil)

	err := svc.DeleteProduct(ctx, 1)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// ---- DeductStock ----

// TestDeductStock_InvalidQuantity 测试 quantity<=0 时返回参数错误
func TestDeductStock_InvalidQuantity(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	err := svc.DeductStock(ctx, 1, 0)

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrParam))
	repo.AssertNotCalled(t, "DeductStock")
}

// TestDeductStock_Success 测试正常扣减库存
func TestDeductStock_Success(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	repo.On("DeductStock", ctx, uint(1), 3).Return(nil)

	err := svc.DeductStock(ctx, 1, 3)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

// TestDeductStock_OutOfStock 测试库存不足时透传错误
func TestDeductStock_OutOfStock(t *testing.T) {
	repo := new(mocks.MockProductRepository)
	svc := newService(repo)
	ctx := context.Background()

	outOfStockErr := errors.New(pkgerrors.Msg(pkgerrors.ErrProductOutOfStock))
	repo.On("DeductStock", ctx, uint(1), 100).Return(outOfStockErr)

	err := svc.DeductStock(ctx, 1, 100)

	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrProductOutOfStock))
	repo.AssertExpectations(t)
}
