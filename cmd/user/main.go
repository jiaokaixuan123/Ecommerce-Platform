package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecommerce-platform/internal/user/domain"
	"github.com/ecommerce-platform/internal/user/handler"
	"github.com/ecommerce-platform/internal/user/repository"
	"github.com/ecommerce-platform/internal/user/service"
	"github.com/ecommerce-platform/pkg/cache"
	"github.com/ecommerce-platform/pkg/config"
	"github.com/ecommerce-platform/pkg/database"
	"github.com/ecommerce-platform/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// 初始化日志
	logger.Init(cfg.App.Mode)
	defer logger.Sync()

	// 初始化 MySQL
	db, err := database.NewMySQL(&cfg.MySQL)
	if err != nil {
		logger.Fatal("failed to connect mysql", zap.Error(err))
	}
	logger.Info("mysql connected")

	// 初始化 Redis
	rdb, err := cache.NewRedis(&cfg.Redis)
	if err != nil {
		logger.Fatal("failed to connect redis", zap.Error(err))
	}
	logger.Info("redis connected")

	// 自动迁移数据库表
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		logger.Fatal("failed to migrate database", zap.Error(err))
	}

	// 初始化依赖
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)
	userHandler := handler.NewUserHandler(userService)

	_ = rdb

	// 初始化 Gin
	gin.SetMode(cfg.App.Mode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 注册路由
	handler.RegisterRoutes(r, userHandler, cfg.JWT.Secret)

	// 启动服务（优雅退出）
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.App.Port),
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("user service started", zap.Int("port", cfg.App.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
