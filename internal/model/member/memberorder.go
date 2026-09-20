package model

import (
	"context"
	"errors"
	"time"

	types "rpc-user/internal/types/member"

	"gorm.io/gorm"
)

var ErrOrderAlreadyProcessed = errors.New("order already processed")

type MemberOrder struct {
	ID            int64 `gorm:"primary_key"`
	UserID        int64
	OrderSN       string `gorm:"column:order_sn"`
	Level         int32
	DurationDays  int32
	Amount        int64
	PayChannel    string
	TransactionID string
	Status        int32
	CreateTime    time.Time
	UpdateTime    time.Time
}

func (m *MemberOrder) TableName() string {
	return "member_order"
}

type MemberOrderModel struct {
	db *gorm.DB
}

func NewMemberOrderModel(db *gorm.DB) *MemberOrderModel {
	return &MemberOrderModel{db: db}
}

func (m *MemberOrderModel) Insert(ctx context.Context, data *MemberOrder) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *MemberOrderModel) FindByTransactionId(ctx context.Context, transactionId string) (*MemberOrder, error) {
	var result MemberOrder
	err := m.db.WithContext(ctx).Where("transaction_id = ?", transactionId).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *MemberOrderModel) FindByOrderSn(ctx context.Context, orderSn string) (*MemberOrder, error) {
	var result MemberOrder
	err := m.db.WithContext(ctx).Where("order_sn = ?", orderSn).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *MemberOrderModel) FindByUserId(ctx context.Context, userId int64, cursor, limit int64) ([]*MemberOrder, error) {
	var result []*MemberOrder
	query := m.db.WithContext(ctx).Where("user_id = ?", userId)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	err := query.Order("id desc").Limit(int(limit)).Find(&result).Error
	return result, err
}

func (m *MemberOrderModel) UpdateStatus(ctx context.Context, id int64, status int32) error {
	return m.db.WithContext(ctx).Model(&MemberOrder{}).
		Where("id = ?", id).Update("status", status).Error
}

func (m *MemberOrderModel) UpdatePaySuccess(ctx context.Context, orderSn string, transactionId string, status int32) error {
	result := m.db.WithContext(ctx).Model(&MemberOrder{}).
		Where("order_sn = ? AND status = ?", orderSn, types.OrderStatusPending).
		Updates(map[string]interface{}{
			"status":         status,
			"transaction_id": transactionId,
			"update_time":    time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrderAlreadyProcessed
	}
	return nil
}
