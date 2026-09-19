package svc

import (
	"rpc-user/user/internal/config"
	"rpc-user/user/internal/model"
	"rpc-user/pkg/orm"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
	DB        *orm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewPostgres(&orm.Config{
		DSN:          c.DataSource,
		MaxOpenConns: 10,
		MaxIdleConns: 100,
		MaxLifetime:  3600,
	})
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(db.DB),
		DB:        db,
	}
}
