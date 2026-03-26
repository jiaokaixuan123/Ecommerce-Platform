package repository_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/payment/domain"
	"github.com/ecommerce-platform/internal/payment/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Payment{}))
	return db
}

func TestRepoCreate_Success(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{
		PaymentNo: "PAY001",
		OrderID:   1,
		UserID:    1,
		Amount:    5000,
		Channel:   domain.PaymentChannelMock,
		Status:    domain.PaymentStatusPending,
	}
	err := repo.Create(ctx, p)
	assert.NoError(t, err)
	assert.NotZero(t, p.ID)
}

func TestRepoCreate_DuplicatePaymentNo(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{PaymentNo: "PAY_DUP", OrderID: 2, UserID: 1, Amount: 1000, Channel: domain.PaymentChannelMock, Status: domain.PaymentStatusPending}
	require.NoError(t, repo.Create(ctx, p))

	p2 := &domain.Payment{PaymentNo: "PAY_DUP", OrderID: 3, UserID: 1, Amount: 2000, Channel: domain.PaymentChannelMock, Status: domain.PaymentStatusPending}
	err := repo.Create(ctx, p2)
	assert.Error(t, err)
}

func TestRepoGetByPaymentNo_Success(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{PaymentNo: "PAY002", OrderID: 10, UserID: 1, Amount: 3000, Channel: domain.PaymentChannelMock, Status: domain.PaymentStatusPending}
	require.NoError(t, repo.Create(ctx, p))

	got, err := repo.GetByPaymentNo(ctx, "PAY002")
	assert.NoError(t, err)
	assert.Equal(t, uint(10), got.OrderID)
	assert.Equal(t, int64(3000), got.Amount)
}

func TestRepoGetByPaymentNo_NotFound(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	got, err := repo.GetByPaymentNo(ctx, "NONEXISTENT")
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestRepoGetByOrderID_Success(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{PaymentNo: "PAY003", OrderID: 20, UserID: 2, Amount: 8000, Channel: domain.PaymentChannelAlipay, Status: domain.PaymentStatusPending}
	require.NoError(t, repo.Create(ctx, p))

	got, err := repo.GetByOrderID(ctx, 20)
	assert.NoError(t, err)
	assert.Equal(t, "PAY003", got.PaymentNo)
}

func TestRepoGetByOrderID_NotFound(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	got, err := repo.GetByOrderID(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestRepoUpdateStatus_Success(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{PaymentNo: "PAY004", OrderID: 30, UserID: 1, Amount: 1500, Channel: domain.PaymentChannelWechat, Status: domain.PaymentStatusPending}
	require.NoError(t, repo.Create(ctx, p))

	err := repo.UpdateStatus(ctx, p.ID, domain.PaymentStatusPending, domain.PaymentStatusSuccess, "TP_12345")
	assert.NoError(t, err)

	updated, err := repo.GetByPaymentNo(ctx, "PAY004")
	require.NoError(t, err)
	assert.Equal(t, domain.PaymentStatusSuccess, updated.Status)
	assert.Equal(t, "TP_12345", updated.ThirdPartyNo)
}

func TestRepoUpdateStatus_WrongFromStatus(t *testing.T) {
	db := setupDB(t)
	repo := repository.NewPaymentRepository(db)
	ctx := context.Background()

	p := &domain.Payment{PaymentNo: "PAY005", OrderID: 40, UserID: 1, Amount: 2500, Channel: domain.PaymentChannelMock, Status: domain.PaymentStatusSuccess}
	require.NoError(t, repo.Create(ctx, p))

	// 尝试从 Pending → Success，但实际状态是 Success（CAS 失败）
	err := repo.UpdateStatus(ctx, p.ID, domain.PaymentStatusPending, domain.PaymentStatusSuccess, "TP_99")
	assert.Error(t, err)
}
