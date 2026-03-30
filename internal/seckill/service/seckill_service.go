package service

import (
	"context"
	"fmt"
	"time"

	ordersvc "github.com/ecommerce-platform/internal/order/service"
	paymentdomain "github.com/ecommerce-platform/internal/payment/domain"
	paymentservice "github.com/ecommerce-platform/internal/payment/service"
	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/ecommerce-platform/internal/seckill/repository"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	stockKey = "SECKILL:STOCK:%d"   // 剩余库存，value = int
	userKey  = "SECKILL:USER:%d:%d" // 用户购买标记，key=seckillID:userID
)

const (
	defaultOutboxPollInterval = 1 * time.Second
	defaultOutboxBatchSize    = 128
	defaultOutboxMaxRetry     = 20
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

type SeckillService interface {
	DoSeckill(ctx context.Context, req *DoSeckillReq) error
	GetSeckillProduct(ctx context.Context, id uint) (*domain.SeckillProduct, error)
	CreateSeckillProduct(ctx context.Context, req *CreateSeckillReq) (*domain.SeckillProduct, error)
	PrewarmStock(ctx context.Context, id uint) error
}

// seckillService:
// 1) 接口层执行 Redis 预扣
// 2) 持久化 saga + outbox
// 3) worker 基于 outbox 持续编排，失败重试，最终补偿
type seckillService struct {
	seckillRepo    repository.SeckillRepository
	sagaRepo       repository.SeckillSagaRepository
	outboxRepo     repository.SeckillOutboxRepository
	orderService   ordersvc.OrderService
	paymentService paymentservice.PaymentService
	rdb            *redis.Client

	eventQueue chan uint // 持久化 outbox 事件 ID
}

func NewSeckillService(
	seckillRepo repository.SeckillRepository,
	sagaRepo repository.SeckillSagaRepository,
	outboxRepo repository.SeckillOutboxRepository,
	orderService ordersvc.OrderService,
	paymentService paymentservice.PaymentService,
	rdb *redis.Client,
	queueSize int,
) SeckillService {
	if queueSize <= 0 {
		queueSize = 1024
	}

	svc := &seckillService{
		seckillRepo:    seckillRepo,
		sagaRepo:       sagaRepo,
		outboxRepo:     outboxRepo,
		orderService:   orderService,
		paymentService: paymentService,
		rdb:            rdb,
		eventQueue:     make(chan uint, queueSize),
	}

	go svc.outboxDispatcher()  // 启动定时扫描
	go svc.asyncWorker()       // 启动执行 worker
	return svc
}

// DoSeckill：秒杀主入口
func (s *seckillService) DoSeckill(ctx context.Context, req *DoSeckillReq) error {
	if req == nil || req.SeckillID == 0 || req.UserID == 0 {
		return pkgerrors.New(pkgerrors.ErrParam)
	}

	product, err := s.seckillRepo.GetByID(ctx, req.SeckillID)
	if err != nil {
		return err
	}

	now := time.Now()
	if now.Before(product.StartAt) {
		return pkgerrors.New(pkgerrors.ErrSeckillNotStarted)
	}
	if now.After(product.EndAt) {
		return pkgerrors.New(pkgerrors.ErrSeckillOver)
	}
	if product.Status == domain.SeckillStatusDisabled {
		return pkgerrors.New(pkgerrors.ErrSeckillDisabled)
	}

	sk := fmt.Sprintf(stockKey, req.SeckillID)
	uk := fmt.Sprintf(userKey, req.SeckillID, req.UserID)
	ttlSeconds := int64(time.Until(product.EndAt).Seconds())
	if ttlSeconds <= 0 {
		return pkgerrors.New(pkgerrors.ErrSeckillOver)
	}

	resp, err := seckillLua.Run(ctx, s.rdb, []string{sk, uk}, ttlSeconds).Result()
	if err != nil {
		return pkgerrors.Wrap(pkgerrors.ErrServer, err)
	}

	code, ok := resp.(int64)
	if !ok {
		return pkgerrors.New(pkgerrors.ErrServer)
	}

	switch code {
	case 1:
		return pkgerrors.New(pkgerrors.ErrSeckillRepeat)
	case 2:
		return pkgerrors.New(pkgerrors.ErrProductOutOfStock)
	case 0:
		// 秒杀成功 → 创建 Saga + Outbox
		saga := &domain.SeckillSaga{		// 初始化步骤、待处理状态
			SeckillID: req.SeckillID,
			UserID:    req.UserID,
			Amount:    product.SeckillPrice,
			Status:    domain.SagaStatusPending,
			Step:      domain.SagaStepInit,
		}
		outbox := &domain.SeckillOutboxEvent{
			MaxRetry:  defaultOutboxMaxRetry,
			NextRunAt: now,
		}
		// 同一事务保存
		if err := s.sagaRepo.CreateSagaWithOutbox(ctx, saga, outbox); err != nil {
			// 如果出错，尽量回滚
			s.compensateRedisBestEffort(req.SeckillID, req.UserID)
			return err
		}

		// 不依赖内存队列可靠性，落库成功就认为请求成功。
		s.tryEnqueueEvent(outbox.ID)
		return nil
	default:
		return pkgerrors.New(pkgerrors.ErrServer)
	}
}

func (s *seckillService) GetSeckillProduct(ctx context.Context, id uint) (*domain.SeckillProduct, error) {
	return s.seckillRepo.GetByID(ctx, id)
}

// CreateSeckillProduct：创建秒杀商品
func (s *seckillService) CreateSeckillProduct(ctx context.Context, req *CreateSeckillReq) (*domain.SeckillProduct, error) {
	if req == nil {
		return nil, pkgerrors.New(pkgerrors.ErrParam)
	}
	if req.StartAt.After(req.EndAt) {
		return nil, pkgerrors.New(pkgerrors.ErrParam)
	}

	seckillProduct := &domain.SeckillProduct{
		ProductID:    req.ProductID,
		ProductName:  req.ProductName,
		ProductImage: req.ProductImage,
		SeckillPrice: req.SeckillPrice,
		TotalStock:   req.TotalStock,
		RemainStock:  req.TotalStock,
		StartAt:      req.StartAt,
		Status:       domain.SeckillStatusPending,
		EndAt:        req.EndAt,
		CreatedAt:    time.Now(),
	}
	// 创建秒杀商品
	if err := s.seckillRepo.Create(ctx, seckillProduct); err != nil {
		return nil, pkgerrors.Wrap(pkgerrors.ErrServer, err)
	}

	// 预热
	if err := s.PrewarmStock(ctx, seckillProduct.ID); err != nil {
		return nil, err
	}

	return seckillProduct, nil
}

// Prewarmstock：预热库存到 Redis
func (s *seckillService) PrewarmStock(ctx context.Context, id uint) error {
	product, err := s.seckillRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product.RemainStock <= 0 {
		return nil
	}

	ttl := time.Until(product.EndAt)
	if ttl <= 0 {
		return pkgerrors.New(pkgerrors.ErrSeckillOver)
	}

	// 
	sk := fmt.Sprintf(stockKey, id)
	if err := s.rdb.Set(ctx, sk, product.RemainStock, ttl).Err(); err != nil {
		return pkgerrors.Wrap(pkgerrors.ErrServer, err)
	}
	return nil
}

// outboxDispatcher 定时扫描待执行事件并派发给 worker。
func (s *seckillService) outboxDispatcher() {
	ticker := time.NewTicker(defaultOutboxPollInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		ids, err := s.outboxRepo.ListDueIDs(ctx, time.Now(), defaultOutboxBatchSize)
		cancel()
		if err != nil {
			continue
		}
		for _, id := range ids {
			s.tryEnqueueEvent(id)
		}
	}
}

// 将事件放入事件队列 s.eventQueue
func (s *seckillService) tryEnqueueEvent(eventID uint) {
	select {
	case s.eventQueue <- eventID:
	default:
		// 队列满不影响最终执行，dispatcher 会继续扫描重试。
	}
}

// asyncWorker 按事件维度执行 saga 步骤并处理重试/补偿。
func (s *seckillService) asyncWorker() {
	for eventID := range s.eventQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		s.handleOutboxEvent(ctx, eventID)
		cancel()
	}
}

// handleOutboxEvent：事件处理（核心调度）
func (s *seckillService) handleOutboxEvent(ctx context.Context, eventID uint) {
	// 抢占事件
	event, claimed, err := s.outboxRepo.ClaimByID(ctx, eventID)
	if err != nil || !claimed {
		return
	}

	// 获取 Saga
	saga, err := s.sagaRepo.GetByID(ctx, event.SagaID)
	if err != nil {
		_ = s.outboxRepo.MarkRetry(ctx, eventID, time.Now().Add(nextRetryDelay(event.RetryCount)), err.Error())
		return
	}

	// 执行状态机
	err = s.executeSaga(ctx, saga)
	if err == nil {
		_ = s.outboxRepo.MarkDone(ctx, eventID)
		return
	}

	// 达到最大重试阈值后进入补偿，防止局部成功状态悬挂。
	if event.RetryCount+1 >= event.MaxRetry {
		compensateErr := s.compensateSaga(ctx, saga, err)
		if compensateErr != nil {
			_ = s.outboxRepo.MarkDead(ctx, eventID, compensateErr.Error())
			return
		}
		_ = s.outboxRepo.MarkDead(ctx, eventID, err.Error())
		return
	}

	_ = s.outboxRepo.MarkRetry(ctx, eventID, time.Now().Add(nextRetryDelay(event.RetryCount)), err.Error())
}

// executeSaga 状态机执行
func (s *seckillService) executeSaga(ctx context.Context, saga *domain.SeckillSaga) error {
	// 检查并更新状态
	if saga.Status == domain.SagaStatusCompleted || saga.Status == domain.SagaStatusCompensated {
		return nil
	}

	if saga.Status != domain.SagaStatusProcessing {
		saga.Status = domain.SagaStatusProcessing
		if err := s.sagaRepo.Save(ctx, saga); err != nil {
			return err
		}
	}

	// 获取商品、创建订单
	product, err := s.seckillRepo.GetByID(ctx, saga.SeckillID)
	if err != nil {
		return err
	}

	if saga.Step == domain.SagaStepInit {
		order, err := s.orderService.CreateOrder(ctx, &ordersvc.CreateOrderReq{
			UserID: saga.UserID,
			Items: []*ordersvc.CreateOrderItem{
				{
					ProductID:    product.ProductID,
					MerchantID:   0,
					ProductName:  product.ProductName,
					ProductImage: product.ProductImage,
					Price:        product.SeckillPrice,
					Quantity:     1,
				},
			},
			Remark: "秒杀订单",
		})
		if err != nil {
			saga.LastError = err.Error()
			_ = s.sagaRepo.Save(ctx, saga)
			return err
		}

		saga.OrderID = order.ID
		saga.Step = domain.SagaStepOrderCreated
		saga.LastError = ""
		if err := s.sagaRepo.Save(ctx, saga); err != nil {
			return err
		}
	}

	// 创建订单
	if saga.Step == domain.SagaStepOrderCreated {
		_, err := s.paymentService.CreatePayment(ctx, &paymentservice.CreatePaymentReq{
			OrderID: saga.OrderID,
			UserID:  saga.UserID,
			Amount:  saga.Amount,
			Channel: paymentdomain.PaymentChannelMock,
		})
		if err != nil {
			saga.LastError = err.Error()
			_ = s.sagaRepo.Save(ctx, saga)
			return err
		}

		saga.Step = domain.SagaStepPaymentCreated
		saga.LastError = ""
		if err := s.sagaRepo.Save(ctx, saga); err != nil {
			return err
		}
	}

	// 同步 DB 库存
	if saga.Step == domain.SagaStepPaymentCreated {
		if err := s.seckillRepo.DecrStock(ctx, saga.SeckillID, 1); err != nil {
			saga.LastError = err.Error()
			_ = s.sagaRepo.Save(ctx, saga)
			return err
		}

		saga.Step = domain.SagaStepStockSynced
		saga.Status = domain.SagaStatusCompleted
		saga.LastError = ""
		if err := s.sagaRepo.Save(ctx, saga); err != nil {
			return err
		}
	}

	return nil
}

// compensateSaga：自动补偿
func (s *seckillService) compensateSaga(ctx context.Context, saga *domain.SeckillSaga, cause error) error {
	saga.Status = domain.SagaStatusCompensating
	saga.LastError = cause.Error()
	if err := s.sagaRepo.Save(ctx, saga); err != nil {
		return err
	}

	// 若订单已创建，尽力取消（幂等，失败会保留失败状态供排查）。
	if saga.OrderID > 0 {
		if err := s.orderService.CancelOrder(ctx, saga.UserID, saga.OrderID); err != nil {
			saga.Status = domain.SagaStatusFailed
			saga.LastError = err.Error()
			_ = s.sagaRepo.Save(ctx, saga)
			return err
		}
	}

	// 回滚 Redis 库存
	if err := s.compensateRedis(ctx, saga.SeckillID, saga.UserID); err != nil {
		saga.Status = domain.SagaStatusFailed
		saga.LastError = err.Error()
		_ = s.sagaRepo.Save(ctx, saga)
		return err
	}

	saga.Status = domain.SagaStatusCompensated
	if err := s.sagaRepo.Save(ctx, saga); err != nil {
		return err
	}
	return nil
}

// compensateRedis：回滚库存
func (s *seckillService) compensateRedis(ctx context.Context, seckillID, userID uint) error {
	sk := fmt.Sprintf(stockKey, seckillID)
	uk := fmt.Sprintf(userKey, seckillID, userID)
	_, err := s.rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(ctx, sk)  // 库存 +1
		pipe.Del(ctx, uk)   // 删除用户标记
		return nil
	})
	return err
}

func (s *seckillService) compensateRedisBestEffort(seckillID, userID uint) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.compensateRedis(ctx, seckillID, userID)
}

// 指数退避重试
func nextRetryDelay(retryCount int) time.Duration {
	if retryCount < 0 {
		retryCount = 0
	}
	// 1s,2s,4s...最大 60s
	delay := time.Second * time.Duration(1<<minInt(retryCount, 6))
	if delay > 60*time.Second {
		return 60 * time.Second
	}
	return delay
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
