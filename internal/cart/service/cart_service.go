package service

import (
	"context"
	"errors"

	"github.com/ecommerce-platform/internal/cart/domain"
	"github.com/ecommerce-platform/internal/cart/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

// ---- 请求结构体 ----

// AddCartItemReq 新增购物车项请求
type AddCartItemReq struct {
	UserID       uint   `json:"user_id"`       // 用户ID（从登录态获取）
	ProductID    uint   `json:"product_id"`    // 商品ID
	MerchantID   uint   `json:"merchant_id"`   // 商家ID
	Quantity     int    `json:"quantity"`      // 商品数量（≥1）
	ProductName  string `json:"product_name"`  // 商品名称（快照）
	ProductPrice int64  `json:"product_price"` // 商品单价（快照，单位：分）
	ProductImage string `json:"product_image"` // 商品图片（快照）
}

// UpdateCartItemQuantityReq 更新购物车项数量请求
type UpdateCartItemQuantityReq struct {
	UserID   uint `json:"user_id"`  // 用户ID
	ItemID   uint `json:"item_id"`  // 购物车项ID
	Quantity int  `json:"quantity"` // 新数量（≥1）
}

// UpdateCartItemSelectedReq 更新购物车项选中状态请求
type UpdateCartItemSelectedReq struct {
	UserID   uint `json:"user_id"`  // 用户ID
	ItemID   uint `json:"item_id"`  // 购物车项ID
	Selected bool `json:"selected"` // 是否选中
}

// DeleteCartItemReq 删除购物车项请求
type DeleteCartItemReq struct {
	UserID    uint `json:"user_id"`    // 用户ID
	ItemID    uint `json:"item_id"`    // 购物车项ID（按 item 删）
	ProductID uint `json:"product_id"` // 商品ID（按商品删该商品所有规格，与 ItemID 二选一）
}

// ---- 响应结构体 ----

// CartItemVO 购物车项视图对象
type CartItemVO struct {
	ID           uint   `json:"id"`
	ProductID    uint   `json:"product_id"`
	MerchantID   uint   `json:"merchant_id"`
	Quantity     int    `json:"quantity"`
	Selected     bool   `json:"selected"`
	ProductName  string `json:"product_name"`
	ProductPrice int64  `json:"product_price"` // 单位：分
	ProductImage string `json:"product_image"`
}

// CartVO 购物车视图对象（含派生数据）
type CartVO struct {
	UserID         uint          `json:"user_id"`
	ItemCount      int           `json:"item_count"`      // 购物车项总数
	TotalPrice     int64         `json:"total_price"`     // 全部商品总价（分）
	SelectedCount  int           `json:"selected_count"`  // 选中项数量
	SelectedAmount int64         `json:"selected_amount"` // 选中商品总价（分）
	Items          []*CartItemVO `json:"items"`
}

// ---- 接口定义 ----

type CartService interface {
	// AddCartItem 加购（已有则累加数量）
	AddCartItem(ctx context.Context, req *AddCartItemReq) error
	// GetUserCart 获取购物车（含派生数据）
	GetUserCart(ctx context.Context, userID uint) (*CartVO, error)
	// UpdateCartItemQuantity 修改某项数量
	UpdateCartItemQuantity(ctx context.Context, req *UpdateCartItemQuantityReq) error
	// UpdateCartItemSelected 修改某项选中状态
	UpdateCartItemSelected(ctx context.Context, req *UpdateCartItemSelectedReq) error
	// DeleteCartItem 删除购物车项
	DeleteCartItem(ctx context.Context, req *DeleteCartItemReq) error
	// ClearCart 清空购物车（下单后调用）
	ClearCart(ctx context.Context, userID uint) error
}

// ---- 实现 ----

type cartService struct {
	cartRepo     repository.CartRepository
	cartItemRepo repository.CartItemRepository
}

func NewCartService(cartRepo repository.CartRepository, cartItemRepo repository.CartItemRepository) CartService {
	return &cartService{
		cartRepo:     cartRepo,
		cartItemRepo: cartItemRepo,
	}
}

// AddCartItem 加购：先获取/创建购物车，再写入商品项
func (s *cartService) AddCartItem(ctx context.Context, req *AddCartItemReq) error {
	if req.Quantity <= 0 {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrParam))
	}

	// 1. 获取或创建用户购物车
	cart, err := s.getUserCart(ctx, req.UserID)
	if err != nil {
		return err
	}

	// 2. 写入购物车项（repo 层已处理"已存在则累加数量"的逻辑）
	item := &domain.CartItem{
		CartID:       cart.ID,
		ProductID:    req.ProductID,
		MerchantID:   req.MerchantID,
		Quantity:     req.Quantity,
		ProductName:  req.ProductName,
		ProductPrice: req.ProductPrice,
		ProductImage: req.ProductImage,
		Selected:     true, // 新加购默认选中
	}
	return s.cartItemRepo.Add(ctx, item)
}

// GetUserCart 获取购物车，组装 CartVO 并计算派生数据
func (s *cartService) GetUserCart(ctx context.Context, userID uint) (*CartVO, error) {
	cart, err := s.getUserCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. 查询购物车项列表
	items, err := s.cartItemRepo.ListByCartID(ctx, cart.ID)
	if err != nil {
		return nil, err
	}

	// 3. 计算派生数据 + 组装 VO
	totalPrice, selectedCount, selectedAmount := s.calculateDerivedData(items)

	vos := make([]*CartItemVO, 0, len(items))
	for _, item := range items {
		vos = append(vos, toCartItemVO(item))
	}

	return &CartVO{
		UserID:         userID,
		ItemCount:      len(items),
		TotalPrice:     totalPrice,
		SelectedCount:  selectedCount,
		SelectedAmount: selectedAmount,
		Items:          vos,
	}, nil
}

// UpdateCartItemQuantity 修改购物车项数量
func (s *cartService) UpdateCartItemQuantity(ctx context.Context, req *UpdateCartItemQuantityReq) error {
	if req.Quantity <= 0 {
		return errors.New(pkgerrors.Msg(pkgerrors.ErrParam))
	}
	if _, err := s.verifyItemOwnership(ctx, req.ItemID, req.UserID); err != nil {
		return err
	}
	return s.cartItemRepo.UpdateQuantity(ctx, req.ItemID, req.Quantity)
}

// UpdateCartItemSelected 修改购物车项选中状态
func (s *cartService) UpdateCartItemSelected(ctx context.Context, req *UpdateCartItemSelectedReq) error {
	if _, err := s.verifyItemOwnership(ctx, req.ItemID, req.UserID); err != nil {
		return err
	}
	return s.cartItemRepo.UpdateSelected(ctx, req.ItemID, req.Selected)
}

// DeleteCartItem 删除购物车项：
//   - 传 ItemID > 0：按 item 删除单条
//   - 传 ProductID > 0：删除该商品的所有规格
func (s *cartService) DeleteCartItem(ctx context.Context, req *DeleteCartItemReq) error {
	if req.ItemID > 0 {
		if _, err := s.verifyItemOwnership(ctx, req.ItemID, req.UserID); err != nil {
			return err
		}
		return s.cartItemRepo.Delete(ctx, req.ItemID)
	}

	if req.ProductID > 0 {
		cart, err := s.getUserCart(ctx, req.UserID)
		if err != nil {
			return err
		}
		return s.cartItemRepo.DeleteByCartIDAndProductID(ctx, cart.ID, req.ProductID)
	}

	return errors.New(pkgerrors.Msg(pkgerrors.ErrParam))
}

// ClearCart 清空购物车（下单成功后调用）
func (s *cartService) ClearCart(ctx context.Context, userID uint) error {
	cart, err := s.getUserCart(ctx, userID)
	if err != nil {
		return err
	}
	return s.cartItemRepo.ClearByCartID(ctx, cart.ID)
}

// ---- 私有辅助方法 ----

// getUserCart 获取用户购物车，统一封装错误
func (s *cartService) getUserCart(ctx context.Context, userID uint) (*domain.Cart, error) {
	cart, err := s.cartRepo.GetOrCreateByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

// verifyItemOwnership 验证 item 存在且属于当前用户的购物车
// 返回 item 本身供调用方直接使用，避免二次查询
func (s *cartService) verifyItemOwnership(ctx context.Context, itemID, userID uint) (*domain.CartItem, error) {
	item, err := s.cartItemRepo.GetItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrNotFound))
		}
		return nil, err
	}
	cart, err := s.getUserCart(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !item.BelongsToCart(cart.ID) {
		return nil, errors.New(pkgerrors.Msg(pkgerrors.ErrForbidden))
	}
	return item, nil
}

// calculateDerivedData 计算购物车派生数据
func (s *cartService) calculateDerivedData(items []*domain.CartItem) (totalPrice int64, selectedCount int, selectedAmount int64) {
	for _, item := range items {
		totalPrice += item.ProductPrice * int64(item.Quantity)
		if item.Selected {
			selectedCount++
			selectedAmount += item.ProductPrice * int64(item.Quantity)
		}
	}
	return
}

// toCartItemVO domain 转 VO
func toCartItemVO(item *domain.CartItem) *CartItemVO {
	return &CartItemVO{
		ID:           item.ID,
		ProductID:    item.ProductID,
		MerchantID:   item.MerchantID,
		Quantity:     item.Quantity,
		Selected:     item.Selected,
		ProductName:  item.ProductName,
		ProductPrice: item.ProductPrice,
		ProductImage: item.ProductImage,
	}
}
