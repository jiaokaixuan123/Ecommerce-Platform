package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ecommerce-platform/internal/order/service"
	ordersvc "github.com/ecommerce-platform/internal/order/service"
	paymentservice "github.com/ecommerce-platform/internal/payment/service"
	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/ecommerce-platform/internal/seckill/repository"
	"github.com/redis/go-redis/v9"
)

// Redis key 模板
const (
	stockKey  = "SECKILL:STOCK:%d"       // 剩余库存，value = int
	userKey   = "SECKILL:USER:%d:%d"     // 用户购买标记，key=seckillID:userID
)

// Lua 脚本：原子检查用户限购 + 库存扣减
// KEYS[1] = stockKey, KEYS[2] = userKey
// ARGV[1] = TTL(seconds)
// 返回：0=成功，1=已购，2=库存不足
var seckillLua = redis.NewScript(`
local bought = redis.call('EXISTS', KEYS[2])
if bought == 1 then return 1 end
local stock = tonumber(redis.call('GET', KEYS[1]))
if stock == nil or stock <= 0 then return 2 end
redis.call('DECR', KEYS[1])
redis.call('SET', KEYS[2], 1, 'EX', ARGV[1])
return 0
`)

// ---- 请求/响应结构体 ----

type DoSeckillReq struct {
	SeckillID uint `json:"seckill_id" binding:"required"`
	UserID    uint `json:"user_id"` // 从登录态获取
}

type CreateSeckillReq struct {
	ProductID    uint      `json:"product_id" binding:"required"`
	ProductName  string    `json:"product_name" binding:"required"`
	ProductImage string    `json:"product_image"`
	SeckillPrice int64     `json:"seckill_price" binding:"required,min=1"`
	TotalStock   int       `json:"total_stock" binding:"required,min=1"`
	StartAt      time.Time `json:"start_at" binding:"required"`
	EndAt        time.Time `json:"end_at" binding:"required"`
}

// ---- 接口定义 ----

type SeckillService interface {
	// DoSeckill 发起秒杀（Redis 预扣 + 异步落库）
	DoSeckill(ctx context.Context, req *DoSeckillReq) error

	// GetSeckillProduct 查询秒杀商品详情
	GetSeckillProduct(ctx context.Context, id uint) (*domain.SeckillProduct, error)

	// CreateSeckillProduct 创建秒杀活动（管理员）
	CreateSeckillProduct(ctx context.Context, req *CreateSeckillReq) (*domain.SeckillProduct, error)

	// PrewarmStock 预热库存到 Redis（启动时或管理接口触发）
	PrewarmStock(ctx context.Context, id uint) error
}

// ---- 实现 ----

type seckillService struct {
	seckillRepo    repository.SeckillRepository
	orderService   ordersvc.OrderService
	paymentService paymentservice.PaymentService
	rdb            *redis.Client
	// 异步落库队列（生产环境替换为 Kafka/RocketMQ）
	orderQueue     chan *domain.SeckillOrder
}

func NewSeckillService(
	seckillRepo repository.SeckillRepository,
	orderService service.OrderService,
	paymentService paymentservice.PaymentService,
	rdb *redis.Client,
	queueSize int,
) SeckillService {
	svc := &seckillService{
		seckillRepo:    seckillRepo,
		orderService:   orderService,
		paymentService: paymentService,
		rdb:            rdb,
		orderQueue:     make(chan *domain.SeckillOrder, queueSize),
	}
	// 启动后台 worker
	go svc.asyncWorker()
	return svc
}

// DoSeckill 核心秒杀逻辑
func (s *seckillService) DoSeckill(ctx context.Context, req *DoSeckillReq) error {
	// TODO: 1. 查询秒杀商品，校验活动时间窗口
	// TODO: 2. 执行 Lua 脚本原子扣减 Redis 库存
	//          code=0: 成功，入队; code=1: 重复购买; code=2: 库存不足
	// TODO: 3. 构造 SeckillOrder 写入 orderQueue（非阻塞，队满返回错误）
	_ = fmt.Sprintf(stockKey, req.SeckillID)
	_ = fmt.Sprintf(userKey, req.SeckillID, req.UserID)
	_ = seckillLua
	_ = errors.New("")
	panic("not implemented")
}

func (s *seckillService) GetSeckillProduct(ctx context.Context, id uint) (*domain.SeckillProduct, error) {
	// TODO: 从 repo 查询，可加 Redis 缓存
	return s.seckillRepo.GetByID(ctx, id)
}

func (s *seckillService) CreateSeckillProduct(ctx context.Context, req *CreateSeckillReq) (*domain.SeckillProduct, error) {
	// TODO: 构造 SeckillProduct，调用 repo.Create，然后 PrewarmStock
	panic("not implemented")
}

func (s *seckillService) PrewarmStock(ctx context.Context, id uint) error {
	// TODO: 从 DB 读取 remain_stock，写入 Redis stockKey，设置 TTL=EndAt-Now
	panic("not implemented")
}

// asyncWorker 后台消费队列，落库订单+支付
func (s *seckillService) asyncWorker() {
	for order := range s.orderQueue {
		ctx := context.Background()
		// TODO: 1. 调用 orderService.CreateOrder 创建订单
		// TODO: 2. 调用 paymentService.CreatePayment 创建支付记录
		// TODO: 3. 调用 seckillRepo.DecrStock 同步 DB 库存
		// TODO: 4. 失败时补偿：Redis 库存 +1，删除 userKey
		_ = order
		_ = ctx
	}
}
