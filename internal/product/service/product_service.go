package service

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/product/domain"
	"github.com/ecommerce-platform/internal/product/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

// ---- 请求/响应结构体 ----

// CreateProductReq 创建商品请求
type CreateProductReq struct {
	MerchantID  uint   `json:"-"`                                     // 从 JWT 上下文注入，不由客户端传入
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description"`
	Price       int64  `json:"price" binding:"required,gt=0"`         // 单位：分
	Stock       int    `json:"stock" binding:"required,gte=0"`
	CategoryID  uint   `json:"category_id" binding:"required"`
}

// UpdateProductReq 更新商品请求（字段全部可选）
type UpdateProductReq struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Description *string `json:"description"`
	Price       *int64  `json:"price" binding:"omitempty,gt=0"`        // 单位：分
	Stock       *int    `json:"stock" binding:"omitempty,gte=0"`
	Status      *int8   `json:"status" binding:"omitempty,oneof=0 1"`
}

// ListProductReq 分页查询请求
type ListProductReq struct {
	Page       int  `form:"page" binding:"omitempty,min=1"`
	PageSize   int  `form:"page_size" binding:"omitempty,min=1,max=100"`
	CategoryID uint `form:"category_id"`
}

// ListProductResp 分页查询响应
type ListProductResp struct {
	Total    int64             `json:"total"`
	Products []*domain.Product `json:"products"`
}

// ---- 接口定义 ----

// ProductService 商品服务接口
type ProductService interface {
	// 创建商品
	CreateProduct(ctx context.Context, req *CreateProductReq) error

	// 获取商品详情
	GetProduct(ctx context.Context, id uint) (*domain.Product, error)

	// 分页查询商品列表
	ListProducts(ctx context.Context, req *ListProductReq) (*ListProductResp, error)

	// 更新商品信息
	UpdateProduct(ctx context.Context, id uint, req *UpdateProductReq) error

	// 删除商品
	DeleteProduct(ctx context.Context, id uint) error

	// 扣减库存（供订单服务调用）
	DeductStock(ctx context.Context, id uint, quantity int) error
}

// ---- 实现 ----

type productService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) ProductService {
	return &productService{repo: repo}
}

// 实现 CreateProduct
func (s *productService) CreateProduct(ctx context.Context, req *CreateProductReq) error {

	product := &domain.Product{
		MerchantID:  req.MerchantID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		Status:      1,
	}

	return s.repo.Create(ctx, product)
}

// 实现 GetProduct
func (s *productService) GetProduct(ctx context.Context, id uint) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound))
		}
		return nil, err
	}
	return product, nil
}

// 实现 ListProducts
func (s *productService) ListProducts(ctx context.Context, req *ListProductReq) (*ListProductResp, error) {
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	products, total, err := s.repo.List(ctx, offset, pageSize, req.CategoryID)
	if err != nil {
		return nil, err
	}

	return &ListProductResp{
		Total:    total,
		Products: products,
	}, nil
}

// 实现 UpdateProduct
func (s *productService) UpdateProduct(ctx context.Context, id uint, req *UpdateProductReq) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(pkgerrors.Msg(pkgerrors.ErrProductNotFound))
		} else {
			return err
		}
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}

	return s.repo.Update(ctx, product)
}

// 实现 DeleteProduct
func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		} else {
			return err
		}
	}

	return s.repo.Delete(ctx, id)
}

// 实现 DeductStock
func (s *productService) DeductStock(ctx context.Context, id uint, quantity int) error {
	if quantity <= 0 {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrParam))
	}
	return s.repo.DeductStock(ctx, id, quantity)
}
