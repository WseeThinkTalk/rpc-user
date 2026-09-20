package model

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Member struct {
	ID         int64 `gorm:"primary_key"`
	UserID     int64
	Level      int32
	ExpireTime time.Time
	Status     int32
	CreateTime time.Time
	UpdateTime time.Time
}

func (m *Member) TableName() string {
	return "member"
}

type MemberModel struct {
	db *gorm.DB
}

func NewMemberModel(db *gorm.DB) *MemberModel {
	return &MemberModel{db: db}
}

func (m *MemberModel) Insert(ctx context.Context, data *Member) error {
	return m.db.WithContext(ctx).Create(data).Error
}

func (m *MemberModel) FindByUserId(ctx context.Context, userId int64) (*Member, error) {
	var result Member
	err := m.db.WithContext(ctx).Where("user_id = ?", userId).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (m *MemberModel) UpsertMember(ctx context.Context, data *Member) error {
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"level", "expire_time", "status", "update_time"}),
	}).Create(data).Error
}

func (m *MemberModel) FindExpiredActive(ctx context.Context, before time.Time) ([]*Member, error) {
	var result []*Member
	err := m.db.WithContext(ctx).
		Where("status = ? AND expire_time < ?", 1, before).
		Find(&result).Error
	return result, err
}

func (m *MemberModel) UpdateStatus(ctx context.Context, id int64, status int32) error {
	return m.db.WithContext(ctx).Model(&Member{}).
		Where("id = ?", id).Update("status", status).Error
}
