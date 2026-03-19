package repository_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/product/domain"
	"github.com/ecommerce-platform/internal/product/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupDB 初始化 SQLite 内存数据库，自动迁移表结构
func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Product{}, &domain.Category{}))
	return db
}

// newProduct 构造一个测试用商品（避免重复代码）
func newProduct(name string, price int64, stock int, category uint) *domain.Product {
	return &domain.Product{
		MerchantID: 1,
		Name:       name,
		Price:      price,
		Stock:      stock,
		CategoryID: category,
		Status:     1,
	}
}

// TestCreate_And_GetByID 测试创建商品后能通过 ID 查询到
func TestCreate_And_GetByID(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	product := newProduct("testProduct", 10, 3, 1)
	require.NoError(t, repo.Create(ctx, product))
	assert.NotZero(t, product.ID)

	found, err := repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, "testProduct", found.Name)

}

// TestGetByID_NotFound 测试查询不存在的商品返回 ErrRecordNotFound
func TestGetByID_NotFound(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	product, err := repo.GetByID(ctx, 999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	assert.Nil(t, product)
}

// TestList 测试分页查询和分类过滤
func TestList(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	products := []*domain.Product{
		newProduct("product1", 5, 3, 1),
		newProduct("product2", 4, 4, 1),
		newProduct("product3", 3, 2, 2),
	}

	for _, p := range products {
		require.NoError(t, repo.Create(ctx, p))
	}

	// 查询全部
	listAll, totalAll, err := repo.List(ctx, 0, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalAll)
	assert.Len(t, listAll, 3)

	// 按分类过滤
	listOf1, totalOf1, err := repo.List(ctx, 0, 10, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalOf1)
	assert.Len(t, listOf1, 2)

	// 分页
	listPage, totalPage, err := repo.List(ctx, 0, 2, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(3), totalPage)
	assert.Len(t, listPage, 2)

}

// TestUpdate 测试更新商品信息
func TestUpdate(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	product := newProduct("testProduct", 2, 3, 1)

	require.NoError(t, repo.Create(ctx, product))
	product.Name = "modifiedProduct"
	product.Price = 100

	require.NoError(t, repo.Update(ctx, product))
	modifiedProduct, err := repo.GetByID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, "modifiedProduct", modifiedProduct.Name)
	assert.Equal(t, int64(100), modifiedProduct.Price)
}

// TestDelete 测试删除商品后无法再查询到
func TestDelete(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	product := newProduct("testProduct", 3, 4, 2)
	require.NoError(t, repo.Create(ctx, product))
	require.NoError(t, repo.Delete(ctx, product.ID))
	_, err := repo.GetByID(ctx, product.ID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

// TestDeductStock_Success 测试正常扣减库存
func TestDeductStock_Success(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	product := newProduct("testProduct", 3, 10, 1)
	require.NoError(t, repo.Create(ctx, product))
	require.NoError(t, repo.DeductStock(ctx, product.ID, 3))
	newProduct, _ := repo.GetByID(ctx, product.ID)
	assert.Equal(t, 7, newProduct.Stock)
}

// TestDeductStock_InsufficientStock 测试库存不足时返回正确错误
func TestDeductStock_InsufficientStock(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	product := newProduct("testProduct", 3, 2, 1)
	require.NoError(t, repo.Create(ctx, product))
	err := repo.DeductStock(ctx, product.ID, 5)
	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrProductOutOfStock))
}

// TestDeductStock_ProductNotFound 测试商品不存在时返回正确错误
func TestDeductStock_ProductNotFound(t *testing.T) {
	repo := repository.NewProductRepository(setupDB(t))
	ctx := context.Background()

	err := repo.DeductStock(ctx, 999, 1)
	assert.EqualError(t, err, pkgerrors.Msg(pkgerrors.ErrProductNotFound))
}
