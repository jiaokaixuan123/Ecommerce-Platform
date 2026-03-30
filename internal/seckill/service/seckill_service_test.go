package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	orderdomain "github.com/ecommerce-platform/internal/order/domain"
	ordermocks "github.com/ecommerce-platform/internal/order/mocks"
	paymentdomain "github.com/ecommerce-platform/internal/payment/domain"
	paymentmocks "github.com/ecommerce-platform/internal/payment/mocks"
	"github.com/ecommerce-platform/internal/seckill/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSeckillRepository struct{ mock.Mock }

func (m *mockSeckillRepository) GetByID(ctx context.Context, id uint) (*domain.SeckillProduct, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SeckillProduct), args.Error(1)
}
func (m *mockSeckillRepository) Create(ctx context.Context, sp *domain.SeckillProduct) error {
	args := m.Called(ctx, sp)
	return args.Error(0)
}
func (m *mockSeckillRepository) DecrStock(ctx context.Context, id uint, quantity int) error {
	args := m.Called(ctx, id, quantity)
	return args.Error(0)
}

type mockSagaRepository struct{ mock.Mock }

func (m *mockSagaRepository) CreateSagaWithOutbox(ctx context.Context, saga *domain.SeckillSaga, outbox *domain.SeckillOutboxEvent) error {
	args := m.Called(ctx, saga, outbox)
	return args.Error(0)
}
func (m *mockSagaRepository) GetByID(ctx context.Context, id uint) (*domain.SeckillSaga, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SeckillSaga), args.Error(1)
}
func (m *mockSagaRepository) Save(ctx context.Context, saga *domain.SeckillSaga) error {
	args := m.Called(ctx, saga)
	return args.Error(0)
}

type mockOutboxRepository struct{ mock.Mock }

func (m *mockOutboxRepository) ListDueIDs(ctx context.Context, now time.Time, limit int) ([]uint, error) {
	args := m.Called(ctx, now, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]uint), args.Error(1)
}
func (m *mockOutboxRepository) ClaimByID(ctx context.Context, id uint) (*domain.SeckillOutboxEvent, bool, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*domain.SeckillOutboxEvent), args.Bool(1), args.Error(2)
}
func (m *mockOutboxRepository) MarkDone(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockOutboxRepository) MarkRetry(ctx context.Context, id uint, nextRunAt time.Time, lastErr string) error {
	args := m.Called(ctx, id, nextRunAt, lastErr)
	return args.Error(0)
}
func (m *mockOutboxRepository) MarkDead(ctx context.Context, id uint, lastErr string) error {
	args := m.Called(ctx, id, lastErr)
	return args.Error(0)
}

func newTestService(
	seckillRepo *mockSeckillRepository,
	sagaRepo *mockSagaRepository,
	outboxRepo *mockOutboxRepository,
	orderService *ordermocks.MockOrderService,
	paymentService *paymentmocks.MockPaymentService,
) *seckillService {
	return &seckillService{
		seckillRepo:    seckillRepo,
		sagaRepo:       sagaRepo,
		outboxRepo:     outboxRepo,
		orderService:   orderService,
		paymentService: paymentService,
		rdb:            redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}),
		eventQueue:     make(chan uint, 16),
	}
}

func TestDoSeckill_InvalidParam(t *testing.T) {
	s := newTestService(
		new(mockSeckillRepository),
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	err := s.DoSeckill(context.Background(), nil)
	assert.Equal(t, pkgerrors.ErrParam, pkgerrors.CodeOf(err))
}

func TestDoSeckill_NotStarted(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	s := newTestService(
		seckillRepo,
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	ctx := context.Background()

	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:          1,
		StartAt:     time.Now().Add(10 * time.Minute),
		EndAt:       time.Now().Add(1 * time.Hour),
		Status:      domain.SeckillStatusPending,
		RemainStock: 100,
	}, nil)

	err := s.DoSeckill(ctx, &DoSeckillReq{SeckillID: 1, UserID: 2})
	assert.Equal(t, pkgerrors.ErrSeckillNotStarted, pkgerrors.CodeOf(err))
}

func TestExecuteSaga_Success(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	outboxRepo := new(mockOutboxRepository)
	orderService := new(ordermocks.MockOrderService)
	paymentService := new(paymentmocks.MockPaymentService)
	s := newTestService(seckillRepo, sagaRepo, outboxRepo, orderService, paymentService)
	ctx := context.Background()

	saga := &domain.SeckillSaga{
		ID:        10,
		SeckillID: 1,
		UserID:    2,
		Amount:    5000,
		Status:    domain.SagaStatusPending,
		Step:      domain.SagaStepInit,
	}

	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:           1,
		ProductID:    101,
		ProductName:  "phone",
		ProductImage: "img",
		SeckillPrice: 5000,
	}, nil)
	orderService.On("CreateOrder", ctx, mock.AnythingOfType("*service.CreateOrderReq")).
		Return(&orderdomain.Order{ID: 100, UserID: 2, TotalAmount: 5000}, nil)
	paymentService.On("CreatePayment", ctx, mock.AnythingOfType("*service.CreatePaymentReq")).
		Return(&paymentdomain.Payment{ID: 200, OrderID: 100}, nil)
	seckillRepo.On("DecrStock", ctx, uint(1), 1).Return(nil)
	sagaRepo.On("Save", ctx, mock.AnythingOfType("*domain.SeckillSaga")).Return(nil)

	err := s.executeSaga(ctx, saga)
	assert.NoError(t, err)
	assert.Equal(t, domain.SagaStatusCompleted, saga.Status)
	assert.Equal(t, domain.SagaStepStockSynced, saga.Step)
	assert.Equal(t, uint(100), saga.OrderID)
}

func TestHandleOutboxEvent_RetryOnFailure(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	outboxRepo := new(mockOutboxRepository)
	orderService := new(ordermocks.MockOrderService)
	paymentService := new(paymentmocks.MockPaymentService)
	s := newTestService(seckillRepo, sagaRepo, outboxRepo, orderService, paymentService)
	ctx := context.Background()

	event := &domain.SeckillOutboxEvent{
		ID:         1,
		SagaID:     10,
		RetryCount: 0,
		MaxRetry:   3,
		Status:     domain.OutboxStatusProcessing,
	}
	saga := &domain.SeckillSaga{
		ID:        10,
		SeckillID: 1,
		UserID:    2,
		Amount:    5000,
		Status:    domain.SagaStatusPending,
		Step:      domain.SagaStepInit,
	}

	outboxRepo.On("ClaimByID", ctx, uint(1)).Return(event, true, nil)
	sagaRepo.On("GetByID", ctx, uint(10)).Return(saga, nil)
	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:           1,
		ProductID:    101,
		ProductName:  "phone",
		ProductImage: "img",
		SeckillPrice: 5000,
	}, nil)
	orderService.On("CreateOrder", ctx, mock.AnythingOfType("*service.CreateOrderReq")).Return(nil, errors.New("create order failed"))
	sagaRepo.On("Save", ctx, mock.AnythingOfType("*domain.SeckillSaga")).Return(nil)
	outboxRepo.On("MarkRetry", ctx, uint(1), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Return(nil)

	s.handleOutboxEvent(ctx, 1)
	outboxRepo.AssertExpectations(t)
}

func TestHandleOutboxEvent_MaxRetryMarkDead(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	outboxRepo := new(mockOutboxRepository)
	orderService := new(ordermocks.MockOrderService)
	paymentService := new(paymentmocks.MockPaymentService)
	s := newTestService(seckillRepo, sagaRepo, outboxRepo, orderService, paymentService)
	ctx := context.Background()

	event := &domain.SeckillOutboxEvent{
		ID:         2,
		SagaID:     20,
		RetryCount: 0,
		MaxRetry:   1,
		Status:     domain.OutboxStatusProcessing,
	}
	saga := &domain.SeckillSaga{
		ID:        20,
		SeckillID: 1,
		UserID:    2,
		Amount:    5000,
		Status:    domain.SagaStatusPending,
		Step:      domain.SagaStepInit,
	}

	outboxRepo.On("ClaimByID", ctx, uint(2)).Return(event, true, nil)
	sagaRepo.On("GetByID", ctx, uint(20)).Return(saga, nil)
	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:           1,
		ProductID:    101,
		ProductName:  "phone",
		ProductImage: "img",
		SeckillPrice: 5000,
	}, nil)
	orderService.On("CreateOrder", ctx, mock.AnythingOfType("*service.CreateOrderReq")).Return(nil, errors.New("create order failed"))
	sagaRepo.On("Save", ctx, mock.AnythingOfType("*domain.SeckillSaga")).Return(nil)
	outboxRepo.On("MarkDead", ctx, uint(2), mock.AnythingOfType("string")).Return(nil)

	s.handleOutboxEvent(ctx, 2)
	outboxRepo.AssertCalled(t, "MarkDead", ctx, uint(2), mock.AnythingOfType("string"))
	outboxRepo.AssertNotCalled(t, "MarkRetry", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestCompensateSaga_CancelOrderFailed(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	outboxRepo := new(mockOutboxRepository)
	orderService := new(ordermocks.MockOrderService)
	paymentService := new(paymentmocks.MockPaymentService)
	s := newTestService(seckillRepo, sagaRepo, outboxRepo, orderService, paymentService)
	ctx := context.Background()

	saga := &domain.SeckillSaga{
		ID:        30,
		SeckillID: 1,
		UserID:    2,
		OrderID:   99,
		Status:    domain.SagaStatusProcessing,
		Step:      domain.SagaStepOrderCreated,
	}
	orderService.On("CancelOrder", ctx, uint(2), uint(99)).Return(errors.New("cancel failed"))
	sagaRepo.On("Save", ctx, mock.AnythingOfType("*domain.SeckillSaga")).Return(nil)

	err := s.compensateSaga(ctx, saga, errors.New("execute failed"))
	assert.Error(t, err)
	assert.Equal(t, domain.SagaStatusFailed, saga.Status)
}

func TestNextRetryDelay_Bounds(t *testing.T) {
	assert.Equal(t, 1*time.Second, nextRetryDelay(0))
	assert.Equal(t, 2*time.Second, nextRetryDelay(1))
	assert.Equal(t, 60*time.Second, nextRetryDelay(6))
	assert.Equal(t, 60*time.Second, nextRetryDelay(100))
	assert.Equal(t, 1*time.Second, nextRetryDelay(-1))
}

func TestDoSeckill_RedisLuaRepeat(t *testing.T) {
	mr := newMiniRedisOrSkip(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	s := &seckillService{
		seckillRepo: seckillRepo,
		sagaRepo:    sagaRepo,
		outboxRepo:  new(mockOutboxRepository),
		rdb:         rdb,
		eventQueue:  make(chan uint, 1),
	}

	ctx := context.Background()
	secID, userID := uint(1), uint(2)
	seckillRepo.On("GetByID", ctx, secID).Return(&domain.SeckillProduct{
		ID:           secID,
		StartAt:      time.Now().Add(-time.Minute),
		EndAt:        time.Now().Add(time.Hour),
		Status:       domain.SeckillStatusActive,
		RemainStock:  10,
		SeckillPrice: 100,
	}, nil)
	mr.Set(fmt.Sprintf(stockKey, secID), "10")
	mr.Set(fmt.Sprintf(userKey, secID, userID), "1")

	err := s.DoSeckill(ctx, &DoSeckillReq{SeckillID: secID, UserID: userID})
	assert.Equal(t, pkgerrors.ErrSeckillRepeat, pkgerrors.CodeOf(err))
	sagaRepo.AssertNotCalled(t, "CreateSagaWithOutbox", mock.Anything, mock.Anything, mock.Anything)
}

func TestDoSeckill_RedisLuaOutOfStock(t *testing.T) {
	mr := newMiniRedisOrSkip(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	s := &seckillService{
		seckillRepo: seckillRepo,
		sagaRepo:    sagaRepo,
		outboxRepo:  new(mockOutboxRepository),
		rdb:         rdb,
		eventQueue:  make(chan uint, 1),
	}

	ctx := context.Background()
	secID, userID := uint(1), uint(3)
	seckillRepo.On("GetByID", ctx, secID).Return(&domain.SeckillProduct{
		ID:           secID,
		StartAt:      time.Now().Add(-time.Minute),
		EndAt:        time.Now().Add(time.Hour),
		Status:       domain.SeckillStatusActive,
		RemainStock:  0,
		SeckillPrice: 100,
	}, nil)
	mr.Set(fmt.Sprintf(stockKey, secID), "0")

	err := s.DoSeckill(ctx, &DoSeckillReq{SeckillID: secID, UserID: userID})
	assert.Equal(t, pkgerrors.ErrProductOutOfStock, pkgerrors.CodeOf(err))
	sagaRepo.AssertNotCalled(t, "CreateSagaWithOutbox", mock.Anything, mock.Anything, mock.Anything)
}

func TestDoSeckill_SuccessAndPersistSaga(t *testing.T) {
	mr := newMiniRedisOrSkip(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	s := &seckillService{
		seckillRepo: seckillRepo,
		sagaRepo:    sagaRepo,
		outboxRepo:  new(mockOutboxRepository),
		rdb:         rdb,
		eventQueue:  make(chan uint, 1),
	}

	ctx := context.Background()
	secID, userID := uint(1), uint(4)
	seckillRepo.On("GetByID", ctx, secID).Return(&domain.SeckillProduct{
		ID:           secID,
		StartAt:      time.Now().Add(-time.Minute),
		EndAt:        time.Now().Add(time.Hour),
		Status:       domain.SeckillStatusActive,
		RemainStock:  2,
		SeckillPrice: 100,
	}, nil)
	mr.Set(fmt.Sprintf(stockKey, secID), "2")
	sagaRepo.On("CreateSagaWithOutbox", ctx, mock.AnythingOfType("*domain.SeckillSaga"), mock.AnythingOfType("*domain.SeckillOutboxEvent")).
		Return(nil)

	err := s.DoSeckill(ctx, &DoSeckillReq{SeckillID: secID, UserID: userID})
	assert.NoError(t, err)

	gotStock, err := mr.Get(fmt.Sprintf(stockKey, secID))
	assert.NoError(t, err)
	assert.Equal(t, "1", gotStock)
	_, err = mr.Get(fmt.Sprintf(userKey, secID, userID))
	assert.NoError(t, err)
}

func TestDoSeckill_PersistSagaFailedShouldCompensateRedis(t *testing.T) {
	mr := newMiniRedisOrSkip(t)
	defer mr.Close()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	seckillRepo := new(mockSeckillRepository)
	sagaRepo := new(mockSagaRepository)
	s := &seckillService{
		seckillRepo: seckillRepo,
		sagaRepo:    sagaRepo,
		outboxRepo:  new(mockOutboxRepository),
		rdb:         rdb,
		eventQueue:  make(chan uint, 1),
	}

	ctx := context.Background()
	secID, userID := uint(1), uint(5)
	seckillRepo.On("GetByID", ctx, secID).Return(&domain.SeckillProduct{
		ID:           secID,
		StartAt:      time.Now().Add(-time.Minute),
		EndAt:        time.Now().Add(time.Hour),
		Status:       domain.SeckillStatusActive,
		RemainStock:  2,
		SeckillPrice: 100,
	}, nil)
	mr.Set(fmt.Sprintf(stockKey, secID), "2")
	sagaRepo.On("CreateSagaWithOutbox", ctx, mock.AnythingOfType("*domain.SeckillSaga"), mock.AnythingOfType("*domain.SeckillOutboxEvent")).
		Return(errors.New("db write failed"))

	err := s.DoSeckill(ctx, &DoSeckillReq{SeckillID: secID, UserID: userID})
	assert.Error(t, err)

	gotStock, err := mr.Get(fmt.Sprintf(stockKey, secID))
	assert.NoError(t, err)
	assert.Equal(t, "2", gotStock)
	_, err = mr.Get(fmt.Sprintf(userKey, secID, userID))
	assert.Error(t, err)
}

func newMiniRedisOrSkip(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Skipf("miniredis unavailable in current environment: %v", err)
	}
	return mr
}

func TestCreateSeckillProduct_InvalidParam(t *testing.T) {
	s := newTestService(
		new(mockSeckillRepository),
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	_, err := s.CreateSeckillProduct(context.Background(), nil)
	assert.Equal(t, pkgerrors.ErrParam, pkgerrors.CodeOf(err))
}

func TestCreateSeckillProduct_StartAfterEnd(t *testing.T) {
	s := newTestService(
		new(mockSeckillRepository),
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	_, err := s.CreateSeckillProduct(context.Background(), &CreateSeckillReq{
		ProductID:    1,
		ProductName:  "p",
		SeckillPrice: 100,
		TotalStock:   1,
		StartAt:      time.Now().Add(time.Hour),
		EndAt:        time.Now(),
	})
	assert.Equal(t, pkgerrors.ErrParam, pkgerrors.CodeOf(err))
}

func TestCreateSeckillProduct_CreateFailed(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	s := newTestService(
		seckillRepo,
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	ctx := context.Background()
	seckillRepo.On("Create", ctx, mock.AnythingOfType("*domain.SeckillProduct")).Return(errors.New("db fail"))

	_, err := s.CreateSeckillProduct(ctx, &CreateSeckillReq{
		ProductID:    1,
		ProductName:  "p",
		SeckillPrice: 100,
		TotalStock:   1,
		StartAt:      time.Now(),
		EndAt:        time.Now().Add(time.Hour),
	})
	assert.Equal(t, pkgerrors.ErrServer, pkgerrors.CodeOf(err))
}

func TestPrewarmStock_RemainStockZero(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	s := newTestService(
		seckillRepo,
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	ctx := context.Background()
	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:          1,
		RemainStock: 0,
		EndAt:       time.Now().Add(time.Hour),
	}, nil)
	assert.NoError(t, s.PrewarmStock(ctx, 1))
}

func TestPrewarmStock_TTLExpired(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	s := newTestService(
		seckillRepo,
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	ctx := context.Background()
	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:          1,
		RemainStock: 10,
		EndAt:       time.Now().Add(-time.Second),
	}, nil)
	err := s.PrewarmStock(ctx, 1)
	assert.Equal(t, pkgerrors.ErrSeckillOver, pkgerrors.CodeOf(err))
}

func TestPrewarmStock_RedisSetFailed(t *testing.T) {
	seckillRepo := new(mockSeckillRepository)
	s := newTestService(
		seckillRepo,
		new(mockSagaRepository),
		new(mockOutboxRepository),
		new(ordermocks.MockOrderService),
		new(paymentmocks.MockPaymentService),
	)
	ctx := context.Background()
	seckillRepo.On("GetByID", ctx, uint(1)).Return(&domain.SeckillProduct{
		ID:          1,
		RemainStock: 10,
		EndAt:       time.Now().Add(time.Hour),
	}, nil)
	err := s.PrewarmStock(ctx, 1)
	assert.Equal(t, pkgerrors.ErrServer, pkgerrors.CodeOf(err))
}
