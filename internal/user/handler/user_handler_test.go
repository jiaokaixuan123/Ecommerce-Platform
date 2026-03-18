package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/ecommerce-platform/internal/user/handler"
	"github.com/ecommerce-platform/internal/user/mocks"
	"github.com/ecommerce-platform/internal/user/service"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newRouter(svc *mocks.MockUserService) *gin.Engine {
	r := gin.New()
	h := handler.NewUserHandler(svc)
	r.POST("/api/v1/user/register", h.Register)
	r.POST("/api/v1/user/login", h.Login)
	r.GET("/api/v1/user/:id", h.GetUserInfo)
	return r
}

func post(r *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func get(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---- Register ----

func TestHandler_Register_Success(t *testing.T) {
	svc := new(mocks.MockUserService)
	svc.On("Register", mock.Anything, mock.Anything).Return(nil)

	w := post(newRouter(svc), "/api/v1/user/register", map[string]string{
		"username": "alice", "password": "password123",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":0`)
}

func TestHandler_Register_InvalidParam(t *testing.T) {
	svc := new(mocks.MockUserService)

	// 缺少 password 字段
	w := post(newRouter(svc), "/api/v1/user/register", map[string]string{
		"username": "al", // 少于 min=3 的边界
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":10001`)
	svc.AssertNotCalled(t, "Register")
}

func TestHandler_Register_ServiceError(t *testing.T) {
	svc := new(mocks.MockUserService)
	svc.On("Register", mock.Anything, mock.Anything).Return(errors.New(pkgerrors.Msg(pkgerrors.ErrUserAlreadyExists)))

	w := post(newRouter(svc), "/api/v1/user/register", map[string]string{
		"username": "alice", "password": "password123",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":10000`)
}

// ---- Login ----

func TestHandler_Login_Success(t *testing.T) {
	svc := new(mocks.MockUserService)
	svc.On("Login", mock.Anything, mock.Anything).Return(&service.LoginResp{
		Token: "test-token",
		User:  &domain.User{ID: 1, Username: "alice"},
	}, nil)

	w := post(newRouter(svc), "/api/v1/user/login", map[string]string{
		"username": "alice", "password": "password123",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-token")
}

func TestHandler_Login_Failed(t *testing.T) {
	svc := new(mocks.MockUserService)
	svc.On("Login", mock.Anything, mock.Anything).Return(nil, errors.New(pkgerrors.Msg(pkgerrors.ErrPasswordWrong)))

	w := post(newRouter(svc), "/api/v1/user/login", map[string]string{
		"username": "alice", "password": "wrong",
	})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":10000`)
}

// ---- GetUserInfo ----

func TestHandler_GetUserInfo_Success(t *testing.T) {
	svc := new(mocks.MockUserService)
	svc.On("GetUserInfo", mock.Anything, uint(1)).Return(&domain.User{ID: 1, Username: "alice"}, nil)

	w := get(newRouter(svc), "/api/v1/user/1")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alice")
}

func TestHandler_GetUserInfo_NotFound(t *testing.T) {
	svc := new(mocks.MockUserService)
	svc.On("GetUserInfo", mock.Anything, uint(99)).Return(nil, errors.New(pkgerrors.Msg(pkgerrors.ErrUserNotFound)))

	w := get(newRouter(svc), "/api/v1/user/99")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":11001`)
}
