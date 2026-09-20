package model

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is the interface for user operations
	UserModel interface {
		Insert(ctx context.Context, data *User) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*User, error)
		FindOneByDisplayId(ctx context.Context, displayId string) (*User, error)
		FindOneByMobile(ctx context.Context, mobile string) (*User, error)
		Update(ctx context.Context, data *User) error
		Delete(ctx context.Context, id uint64) error
		FindByMobile(ctx context.Context, mobile string) (*User, error)
		FindAll(ctx context.Context, keyword string, cursor, limit int64) ([]*User, error)
		UpdateRole(ctx context.Context, userId uint64, role int64) error
		UpdatePassword(ctx context.Context, userId uint64, password string) error
	}

	customUserModel struct {
		db *gorm.DB
	}

	User struct {
		Id           uint64    `gorm:"primaryKey;column:id"` // 主键ID
		Username     string    `gorm:"column:username"`       // 用户名
		Avatar       string    `gorm:"column:avatar"`         // 头像
		Mobile       string    `gorm:"column:mobile"`         // 手机号
		Password     string    `gorm:"column:password"`       // 密码
		Role         int64     `gorm:"column:role"`           // 角色 0:普通用户 1:管理员
		DisplayId    string    `gorm:"column:display_id"`     // 用户展示号
		Bio          string    `gorm:"column:bio"`            // 个人简介
		Gender       int64     `gorm:"column:gender"`         // 性别 0:不展示 1:男 2:女 3:其他
		ProfileCover string    `gorm:"column:profile_cover"`  // 主页背景图
		CreateTime   time.Time `gorm:"column:create_time;autoCreateTime"` // 创建时间
		UpdateTime   time.Time `gorm:"column:update_time;autoUpdateTime"` // 最后修改时间
	}
)

func (User) TableName() string {
	return "user"
}

// sqlResult implements sql.Result
type sqlResult struct {
	id       int64
	affected int64
}

func (r sqlResult) LastInsertId() (int64, error) { return r.id, nil }
func (r sqlResult) RowsAffected() (int64, error) { return r.affected, nil }

// NewUserModel returns a model for the database table.
func NewUserModel(db *gorm.DB) UserModel {
	return &customUserModel{
		db: db,
	}
}

func (m *customUserModel) Insert(ctx context.Context, data *User) (sql.Result, error) {
	err := m.db.WithContext(ctx).Create(data).Error
	if err != nil {
		return nil, err
	}
	return sqlResult{id: int64(data.Id), affected: 1}, nil
}

func (m *customUserModel) FindOne(ctx context.Context, id uint64) (*User, error) {
	var resp User
	err := m.db.WithContext(ctx).First(&resp, id).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customUserModel) FindOneByDisplayId(ctx context.Context, displayId string) (*User, error) {
	var resp User
	err := m.db.WithContext(ctx).Where("display_id = ?", displayId).First(&resp).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customUserModel) FindOneByMobile(ctx context.Context, mobile string) (*User, error) {
	var resp User
	err := m.db.WithContext(ctx).Where("mobile = ?", mobile).First(&resp).Error
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customUserModel) Update(ctx context.Context, data *User) error {
	return m.db.WithContext(ctx).Save(data).Error
}

func (m *customUserModel) Delete(ctx context.Context, id uint64) error {
	return m.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (m *customUserModel) FindByMobile(ctx context.Context, mobile string) (*User, error) {
	var resp User
	err := m.db.WithContext(ctx).Where("mobile = ?", mobile).First(&resp).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (m *customUserModel) FindAll(ctx context.Context, keyword string, cursor, limit int64) ([]*User, error) {
	var resp []*User
	query := m.db.WithContext(ctx).Where("id > ?", cursor)
	if keyword != "" {
		likeKW := "%" + keyword + "%"
		query = query.Where("username ILIKE ? OR display_id ILIKE ?", likeKW, likeKW)
	}
	err := query.Order("id asc").Limit(int(limit)).Find(&resp).Error
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *customUserModel) UpdateRole(ctx context.Context, userId uint64, role int64) error {
	return m.db.WithContext(ctx).Model(&User{}).Where("id = ?", userId).Update("role", role).Error
}

func (m *customUserModel) UpdatePassword(ctx context.Context, userId uint64, password string) error {
	return m.db.WithContext(ctx).Model(&User{}).Where("id = ?", userId).Update("password", password).Error
}
