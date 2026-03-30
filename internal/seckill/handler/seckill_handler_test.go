package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecommerce-platform/internal/seckill/domain"
	"github.com/ecommerce-platform/internal/seckill/handler"
	"github.com/ecommerce-platform/internal/seckill/mocks"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/ecommerce-platform/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(h *handler.SeckillHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.UserIDKey, uint(1))
		c.Next()
	})
	r.POST("/seckill/:id", h.DoSeckill)
	r.GET("/seckill/:id", h.GetSeckillProduct)
	return r
}

func parseCode(t *testing.T, body []byte) int {
	t.Helper()
	var resp map[string]any
	err := json.Unmarshal(body, &resp)
	assert.NoError(t, err)
	return int(resp["code"].(float64))
}

func TestDoSeckill_InvalidID(t *testing.T) {
	svc := new(mocks.MockSeckillService)
	h := handler.NewSeckillHandler(svc)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/seckill/xx", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, pkgerrors.ErrParam, parseCode(t, w.Body.Bytes()))
}

func TestDoSeckill_BizErrorCodePassThrough(t *testing.T) {
	svc := new(mocks.MockSeckillService)
	h := handler.NewSeckillHandler(svc)
	r := setupRouter(h)

	svc.On("DoSeckill", mock.Anything, mock.AnythingOfType("*service.DoSeckillReq")).
		Return(pkgerrors.New(pkgerrors.ErrSeckillRepeat))

	req := httptest.NewRequest(http.MethodPost, "/seckill/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, pkgerrors.ErrSeckillRepeat, parseCode(t, w.Body.Bytes()))
}

func TestDoSeckill_Success(t *testing.T) {
	svc := new(mocks.MockSeckillService)
	h := handler.NewSeckillHandler(svc)
	r := setupRouter(h)

	svc.On("DoSeckill", mock.Anything, mock.AnythingOfType("*service.DoSeckillReq")).
		Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/seckill/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, pkgerrors.ErrSuccess, parseCode(t, w.Body.Bytes()))
}

func TestGetSeckillProduct_NotFound(t *testing.T) {
	svc := new(mocks.MockSeckillService)
	h := handler.NewSeckillHandler(svc)
	r := setupRouter(h)

	svc.On("GetSeckillProduct", mock.Anything, uint(1)).
		Return(nil, pkgerrors.New(pkgerrors.ErrProductNotFound))

	req := httptest.NewRequest(http.MethodGet, "/seckill/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, pkgerrors.ErrProductNotFound, parseCode(t, w.Body.Bytes()))
}

func TestGetSeckillProduct_Success(t *testing.T) {
	svc := new(mocks.MockSeckillService)
	h := handler.NewSeckillHandler(svc)
	r := setupRouter(h)

	svc.On("GetSeckillProduct", mock.Anything, uint(1)).
		Return(&domain.SeckillProduct{ID: 1, ProductName: "p"}, nil)

	req := httptest.NewRequest(http.MethodGet, "/seckill/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, pkgerrors.ErrSuccess, parseCode(t, w.Body.Bytes()))
}
