package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/ecommerce-platform/internal/payment/mocks"
	"github.com/ecommerce-platform/internal/payment/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newService(paymentRepo *mocks.MockPaymentRepository) service.PaymentService {
	return service.NewPaymentService(paymentRepo)
}

// ---- CreatePayment ----

func TestCreatePayment_Success(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	// 第一次查询：record not found，应继续创建
	paymentRepo.On("GetByOrderID", ctx, uint(1)).Return(nil, errors.New("record not found"))
	paymentRepo.On("Create", ctx, mock.MatchedBy(func(p *domain.Payment) bool {
		return p.OrderID == 1 && p.UserID == 1 && p.Amount == 10000 && p.Status == domain.PaymentStatusPending
	})).Return(nil)

	resp, err := svc.CreatePayment(ctx, &service.CreatePaymentReq{
		OrderID: 1,
		UserID:  1,
		Amount:  10000,
		Channel: domain.PaymentChannelMock,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint(1), resp.OrderID)
	assert.Equal(t, int64(10000), resp.Amount)
	assert.Equal(t, domain.PaymentStatusPending, resp.Status)
	assert.NotEmpty(t, resp.PaymentNo)
	paymentRepo.AssertExpectations(t)
}

func TestCreatePayment_Idempotent(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	existing := &domain.Payment{
		ID:      1,
		OrderID: 1,
		Amount:  10000,
		Status:  domain.PaymentStatusPending,
	}
	// 已存在支付记录，直接返回，不再调用 Create
	paymentRepo.On("GetByOrderID", ctx, uint(1)).Return(existing, nil)

	resp, err := svc.CreatePayment(ctx, &service.CreatePaymentReq{
		OrderID: 1,
		UserID:  1,
		Amount:  10000,
		Channel: domain.PaymentChannelMock,
	})

	assert.NoError(t, err)
	assert.Equal(t, existing, resp)
	paymentRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	paymentRepo.AssertExpectations(t)
}

func TestCreatePayment_DBError(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	dbErr := errors.New("connection refused")
	paymentRepo.On("GetByOrderID", ctx, uint(1)).Return(nil, dbErr)

	resp, err := svc.CreatePayment(ctx, &service.CreatePaymentReq{
		OrderID: 1,
		UserID:  1,
		Amount:  10000,
		Channel: domain.PaymentChannelMock,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	paymentRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	paymentRepo.AssertExpectations(t)
}

func TestCreatePayment_CreateFails(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	paymentRepo.On("GetByOrderID", ctx, uint(2)).Return(nil, errors.New("record not found"))
	paymentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(errors.New("insert failed"))

	resp, err := svc.CreatePayment(ctx, &service.CreatePaymentReq{
		OrderID: 2,
		UserID:  1,
		Amount:  500,
		Channel: domain.PaymentChannelAlipay,
	})

	assert.Error(t, err)
	assert.Nil(t, resp)
	paymentRepo.AssertExpectations(t)
}

// ---- HandleCallback ----

func TestHandleCallback_Success(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	payment := &domain.Payment{
		ID:        10,
		PaymentNo: "PAY123",
		Status:    domain.PaymentStatusPending,
	}
	paymentRepo.On("GetByPaymentNo", ctx, "PAY123").Return(payment, nil)
	paymentRepo.On("UpdateStatus", ctx, uint(10),
		domain.PaymentStatusPending, domain.PaymentStatusSuccess, "TP456",
	).Return(nil)

	err := svc.HandleCallback(ctx, &service.PaymentCallbackReq{
		PaymentNo:    "PAY123",
		ThirdPartyNo: "TP456",
		Success:      true,
	})

	assert.NoError(t, err)
	paymentRepo.AssertExpectations(t)
}

func TestHandleCallback_Failed(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	payment := &domain.Payment{
		ID:        10,
		PaymentNo: "PAY123",
		Status:    domain.PaymentStatusPending,
	}
	paymentRepo.On("GetByPaymentNo", ctx, "PAY123").Return(payment, nil)
	paymentRepo.On("UpdateStatus", ctx, uint(10),
		domain.PaymentStatusPending, domain.PaymentStatusFailed, "TP456",
	).Return(nil)

	err := svc.HandleCallback(ctx, &service.PaymentCallbackReq{
		PaymentNo:    "PAY123",
		ThirdPartyNo: "TP456",
		Success:      false,
	})

	assert.NoError(t, err)
	paymentRepo.AssertExpectations(t)
}

func TestHandleCallback_PaymentNotFound(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	paymentRepo.On("GetByPaymentNo", ctx, "NOTEXIST").Return(nil, errors.New("record not found"))

	err := svc.HandleCallback(ctx, &service.PaymentCallbackReq{
		PaymentNo:    "NOTEXIST",
		ThirdPartyNo: "TP789",
		Success:      true,
	})

	assert.Error(t, err)
	paymentRepo.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything)
	paymentRepo.AssertExpectations(t)
}

func TestHandleCallback_UpdateStatusFails(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	payment := &domain.Payment{
		ID:        10,
		PaymentNo: "PAY123",
		Status:    domain.PaymentStatusPending,
	}
	paymentRepo.On("GetByPaymentNo", ctx, "PAY123").Return(payment, nil)
	paymentRepo.On("UpdateStatus", ctx, uint(10),
		domain.PaymentStatusPending, domain.PaymentStatusSuccess, "TP456",
	).Return(errors.New("支付流水重复"))

	err := svc.HandleCallback(ctx, &service.PaymentCallbackReq{
		PaymentNo:    "PAY123",
		ThirdPartyNo: "TP456",
		Success:      true,
	})

	assert.Error(t, err)
	paymentRepo.AssertExpectations(t)
}

// ---- GetPaymentByOrderID ----

func TestGetPaymentByOrderID_Success(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	expected := &domain.Payment{ID: 5, OrderID: 99, Amount: 8888}
	paymentRepo.On("GetByOrderID", ctx, uint(99)).Return(expected, nil)

	result, err := svc.GetPaymentByOrderID(ctx, 99)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	paymentRepo.AssertExpectations(t)
}

func TestGetPaymentByOrderID_NotFound(t *testing.T) {
	paymentRepo := new(mocks.MockPaymentRepository)
	svc := newService(paymentRepo)
	ctx := context.Background()

	paymentRepo.On("GetByOrderID", ctx, uint(404)).Return(nil, errors.New("record not found"))

	result, err := svc.GetPaymentByOrderID(ctx, 404)

	assert.Error(t, err)
	assert.Nil(t, result)
	paymentRepo.AssertExpectations(t)
}
