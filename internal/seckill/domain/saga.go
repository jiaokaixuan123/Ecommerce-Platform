package domain

import "time"

// SeckillSaga（状态机）：跟踪秒杀整个流程的 “事务状态”
type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"
	SagaStatusProcessing   SagaStatus = "processing"
	SagaStatusCompleted    SagaStatus = "completed"
	SagaStatusCompensating SagaStatus = "compensating"
	SagaStatusCompensated  SagaStatus = "compensated"
	SagaStatusFailed       SagaStatus = "failed"
)

type SagaStep string

const (
	SagaStepInit           SagaStep = "init"
	SagaStepOrderCreated   SagaStep = "order_created"
	SagaStepPaymentCreated SagaStep = "payment_created"
	SagaStepStockSynced    SagaStep = "stock_synced"
)

// SeckillSaga 秒杀编排实例（持久化状态机）
type SeckillSaga struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	SeckillID uint       `gorm:"not null;index:idx_seckill_user,unique" json:"seckill_id"`
	UserID    uint       `gorm:"not null;index:idx_seckill_user,unique" json:"user_id"`
	Amount    int64      `gorm:"not null" json:"amount"`
	OrderID   uint       `gorm:"not null;default:0" json:"order_id"`
	Status    SagaStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	Step      SagaStep   `gorm:"type:varchar(40);not null" json:"step"`
	LastError string     `gorm:"size:512" json:"last_error"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (SeckillSaga) TableName() string { return "seckill_sagas" }

// SeckillOutboxEvent（发件箱模式）

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"			// 待执行
	OutboxStatusProcessing OutboxStatus = "processing"		// 执行中
	OutboxStatusDone       OutboxStatus = "done"			// 完成
	OutboxStatusDead       OutboxStatus = "dead"			// 死信，人工处理
)

// SeckillOutboxEvent Outbox 事件表（持久化执行引擎入口）
type SeckillOutboxEvent struct {
	ID         uint         `gorm:"primaryKey" json:"id"`
	SagaID     uint         `gorm:"not null;index" json:"saga_id"`
	Topic      string       `gorm:"size:64;not null;index" json:"topic"`
	Status     OutboxStatus `gorm:"type:varchar(20);not null;index" json:"status"`
	RetryCount int          `gorm:"not null;default:0" json:"retry_count"`
	MaxRetry   int          `gorm:"not null;default:20" json:"max_retry"`
	NextRunAt  time.Time    `gorm:"not null;index" json:"next_run_at"`
	LastError  string       `gorm:"size:512" json:"last_error"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

func (SeckillOutboxEvent) TableName() string { return "seckill_outbox_events" }
