package svc

import (
	"rpc-user/member/internal/config"
	"rpc-user/member/internal/model"
	"rpc-user/pkg/orm"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config           config.Config
	DB               *orm.DB
	MemberModel      *model.MemberModel
	MemberOrderModel *model.MemberOrderModel
	BizRedis         *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := orm.MustNewMysql(&orm.Config{
		DSN:          c.DB.DataSource,
		MaxOpenConns: c.DB.MaxOpenConns,
		MaxIdleConns: c.DB.MaxIdleConns,
		MaxLifetime:  c.DB.MaxLifetime,
	})

	rds := redis.MustNewRedis(redis.RedisConf{
		Host:        c.BizRedis.Host,
		Pass:        c.BizRedis.Pass,
		Type:        c.BizRedis.Type,
		PingTimeout: 10000000000,
	})

	return &ServiceContext{
		Config:           c,
		DB:               db,
		MemberModel:      model.NewMemberModel(db.DB),
		MemberOrderModel: model.NewMemberOrderModel(db.DB),
		BizRedis:         rds,
	}
}
