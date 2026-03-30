package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ecommerce-platform/internal/seckill/domain"
	pkgerrors "github.com/ecommerce-platform/pkg/errors"
	"gorm.io/gorm"
)

const outboxTopicSeckillSaga = "seckill.saga.execute"

type SeckillSagaRepository interface {
	// 同一个事务保存 saga + outbox
	CreateSagaWithOutbox(ctx context.Context, saga *domain.SeckillSaga, outbox *domain.SeckillOutboxEvent) error
	GetByID(ctx context.Context, id uint) (*domain.SeckillSaga, error)
	Save(ctx context.Context, saga *domain.SeckillSaga) error
}

// Outbox 后台 worker 工作流
type SeckillOutboxRepository interface {
	ListDueIDs(ctx context.Context, now time.Time, limit int) ([]uint, error)
	ClaimByID(ctx context.Context, id uint) (*domain.SeckillOutboxEvent, bool, error)
	MarkDone(ctx context.Context, id uint) error
	MarkRetry(ctx context.Context, id uint, nextRunAt time.Time, lastErr string) error
	MarkDead(ctx context.Context, id uint, lastErr string) error
}

type seckillSagaRepository struct {
	db *gorm.DB
}

type seckillOutboxRepository struct {
	db *gorm.DB
}

func NewSeckillSagaRepository(db *gorm.DB) SeckillSagaRepository {
	return &seckillSagaRepository{db: db}
}

func NewSeckillOutboxRepository(db *gorm.DB) SeckillOutboxRepository {
	return &seckillOutboxRepository{db: db}
}

// CreateSagaWithOutbox：同库事务创建 saga + outbox（同一事务边界）
func (r *seckillSagaRepository) CreateSagaWithOutbox(ctx context.Context, saga *domain.SeckillSaga, outbox *domain.SeckillOutboxEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建 saga 
		if err := tx.Create(saga).Error; err != nil {
			if isDuplicateEntry(err) {
				return pkgerrors.New(pkgerrors.ErrSeckillRepeat)
			}
			return err
		}

		// 根据 saga 创建 outbox
		outbox.SagaID = saga.ID
		outbox.Topic = outboxTopicSeckillSaga
		outbox.Status = domain.OutboxStatusPending
		if outbox.NextRunAt.IsZero() {
			outbox.NextRunAt = time.Now()
		}
		if outbox.MaxRetry <= 0 {
			outbox.MaxRetry = 20
		}

		if err := tx.Create(outbox).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *seckillSagaRepository) GetByID(ctx context.Context, id uint) (*domain.SeckillSaga, error) {
	var saga domain.SeckillSaga
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&saga).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.New(pkgerrors.ErrNotFound)
		}
		return nil, err
	}
	return &saga, nil
}

func (r *seckillSagaRepository) Save(ctx context.Context, saga *domain.SeckillSaga) error {
	return r.db.WithContext(ctx).Save(saga).Error
}

// ListDueIDs：扫描待执行事件，取出到期可执行的任务
func (r *seckillOutboxRepository) ListDueIDs(ctx context.Context, now time.Time, limit int) ([]uint, error) {
	if limit <= 0 {
		limit = 100
	}

	var ids []uint
	err := r.db.WithContext(ctx).
		Model(&domain.SeckillOutboxEvent{}).
		Where("topic = ? AND status = ? AND next_run_at <= ?", outboxTopicSeckillSaga, domain.OutboxStatusPending, now).
		Order("id asc").
		Limit(limit).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// ClaimByID：原子抢占事件（pending -> processing）
func (r *seckillOutboxRepository) ClaimByID(ctx context.Context, id uint) (*domain.SeckillOutboxEvent, bool, error) {
	// 改变目标时间的状态 pending -> processing
	result := r.db.WithContext(ctx).
		Model(&domain.SeckillOutboxEvent{}).
		Where("id = ? AND status = ?", id, domain.OutboxStatusPending).
		Update("status", domain.OutboxStatusProcessing)

	if result.Error != nil {
		return nil, false, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, false, nil
	}

	// 找到这个事件并返回
	var event domain.SeckillOutboxEvent
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&event).Error; err != nil {
		return nil, false, err
	}
	return &event, true, nil
}

// MarkDone：完成
func (r *seckillOutboxRepository) MarkDone(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.SeckillOutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     domain.OutboxStatusDone,
			"last_error": "",
		}).Error
}

// MarkRetry：重试
func (r *seckillOutboxRepository) MarkRetry(ctx context.Context, id uint, nextRunAt time.Time, lastErr string) error {
	return r.db.WithContext(ctx).
		Model(&domain.SeckillOutboxEvent{}).
		Where("id = ? AND status = ?", id, domain.OutboxStatusProcessing).
		Updates(map[string]any{
			"status":      domain.OutboxStatusPending,
			"retry_count": gorm.Expr("retry_count + 1"),
			"next_run_at": nextRunAt,
			"last_error":  truncateErr(lastErr),
		}).Error
}

// MarkDead：死信
func (r *seckillOutboxRepository) MarkDead(ctx context.Context, id uint, lastErr string) error {
	return r.db.WithContext(ctx).
		Model(&domain.SeckillOutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":     domain.OutboxStatusDead,
			"last_error": truncateErr(lastErr),
		}).Error
}

// 判断数据库错误是不是 “唯一键冲突”（重复插入）
func isDuplicateEntry(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "unique")
}

// 截断错误信息，防止超长字符串存不进数据库
func truncateErr(msg string) string {
	const maxLen = 500
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen]
}
