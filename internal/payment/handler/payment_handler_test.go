package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/ecommerce-platform/internal/payment/handler"
	"github.com/ecommerce-platform/internal/payment/mocks"
	"github.com/ecommerce-platform/internal/payment/service"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRouter(h *handler.PaymentHandler) *gin.Engine {
	r := gin.New()
	// 注入 user_id，模拟已登录用户
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, uint(1))
		c.Next()
	})
	r.POST("/payments", h.CreatePayment)
	r.POST("/payments/callback", h.HandleCallback)
	r.GET("/payments/order/:order_id", h.GetPaymentByOrderID)
	return r
}

// ---- CreatePayment ----

func TestHandlerCreatePayment_Success(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	expected := &domain.Payment{ID: 1, OrderID: 10, Amount: 5000, Status: domain.PaymentStatusPending}
	svc.On("CreatePayment", mock.Anything, mock.MatchedBy(func(req *service.CreatePaymentReq) bool {
		return req.OrderID == 10 && req.Amount == 5000
	})).Return(expected, nil)

	body, _ := json.Marshal(map[string]any{
		"order_id": 10,
		"amount":   5000,
		"channel":  "mock",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	svc.AssertExpectations(t)
}

func TestHandlerCreatePayment_InvalidBody(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(`{invalid}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10001), resp["code"]) // ErrParam
	svc.AssertNotCalled(t, "CreatePayment")
}

func TestHandlerCreatePayment_ServiceError(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	svc.On("CreatePayment", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

	body, _ := json.Marshal(map[string]any{
		"order_id": 1,
		"amount":   1000,
		"channel":  "mock",
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10000), resp["code"]) // ErrServer
	svc.AssertExpectations(t)
}

// ---- HandleCallback ----

func TestHandlerHandleCallback_Success(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	svc.On("HandleCallback", mock.Anything, mock.MatchedBy(func(req *service.PaymentCallbackReq) bool {
		return req.PaymentNo == "PAY123" && req.ThirdPartyNo == "TP456" && req.Success
	})).Return(nil)

	body, _ := json.Marshal(map[string]any{
		"payment_no":    "PAY123",
		"third_party_no": "TP456",
		"success":       true,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments/callback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	svc.AssertExpectations(t)
}

func TestHandlerHandleCallback_ServiceError(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	svc.On("HandleCallback", mock.Anything, mock.Anything).Return(errors.New("update failed"))

	body, _ := json.Marshal(map[string]any{
		"payment_no":    "PAY123",
		"third_party_no": "TP456",
		"success":       true,
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/payments/callback", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10000), resp["code"])
	svc.AssertExpectations(t)
}

// ---- GetPaymentByOrderID ----

func TestHandlerGetPaymentByOrderID_Success(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	expected := &domain.Payment{ID: 3, OrderID: 7, Amount: 2000}
	svc.On("GetPaymentByOrderID", mock.Anything, uint(7)).Return(expected, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payments/order/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	svc.AssertExpectations(t)
}

func TestHandlerGetPaymentByOrderID_InvalidID(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payments/order/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10001), resp["code"]) // ErrParam
	svc.AssertNotCalled(t, "GetPaymentByOrderID")
}

func TestHandlerGetPaymentByOrderID_NotFound(t *testing.T) {
	svc := new(mocks.MockPaymentService)
	h := handler.NewPaymentHandler(svc)
	r := setupRouter(h)

	svc.On("GetPaymentByOrderID", mock.Anything, uint(99)).Return(nil, errors.New("record not found"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/payments/order/99", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(10000), resp["code"])
	svc.AssertExpectations(t)
}
