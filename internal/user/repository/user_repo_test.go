package repository_test

import (
	"context"
	"testing"

	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/ecommerce-platform/internal/user/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setuoDB：使用SQLite内存数据库
func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.User{}))
	return db
}

// TestCreate_And_GetByID：测试用户创建和根据ID获取用户信息
func TestCreate_And_GetByID(t *testing.T) {
	repo := repository.NewUserRepository(setupDB(t))
	ctx := context.Background()

	user := &domain.User{Username: "alice", Password: "hashed", Email: "alice@example.com", Status: 1}
	require.NoError(t, repo.Create(ctx, user))
	assert.NotZero(t, user.ID)

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "alice", found.Username)
}

func TestGetByUsername(t *testing.T) {
	repo := repository.NewUserRepository(setupDB(t))
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.User{Username: "bob", Password: "hashed", Status: 1}))

	found, err := repo.GetByUsername(ctx, "bob")
	require.NoError(t, err)
	assert.Equal(t, "bob", found.Username)
}

func TestGetByUsername_NotFound(t *testing.T) {
	repo := repository.NewUserRepository(setupDB(t))
	ctx := context.Background()

	_, err := repo.GetByUsername(ctx, "nobody")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestGetByEmail(t *testing.T) {
	repo := repository.NewUserRepository(setupDB(t))
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.User{Username: "carol", Password: "hashed", Email: "carol@example.com", Status: 1}))

	found, err := repo.GetByEmail(ctx, "carol@example.com")
	require.NoError(t, err)
	assert.Equal(t, "carol", found.Username)
}

func TestGetByPhone(t *testing.T) {
	repo := repository.NewUserRepository(setupDB(t))
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &domain.User{Username: "dave", Password: "hashed", Phone: "13800138000", Status: 1}))

	found, err := repo.GetByPhone(ctx, "13800138000")
	require.NoError(t, err)
	assert.Equal(t, "dave", found.Username)
}

func TestUpdate(t *testing.T) {
	repo := repository.NewUserRepository(setupDB(t))
	ctx := context.Background()

	user := &domain.User{Username: "eve", Password: "hashed", Status: 1}
	require.NoError(t, repo.Create(ctx, user))

	user.Nickname = "Eve Nickname"
	require.NoError(t, repo.Update(ctx, user))

	found, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, "Eve Nickname", found.Nickname)
}
