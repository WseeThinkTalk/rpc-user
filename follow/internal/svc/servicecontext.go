package svc

import (
	"rpc-user/follow/internal/config"
	"rpc-user/follow/internal/model"
	"rpc-user/pkg/orm"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *orm.DB
	FollowModel        *model.FollowModel
	FollowCountModel   *model.FollowCountModel
	BizRedis           *redis.Redis
	NotificationPusher *kq.Pusher
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
		Config:             c,
		DB:                 db,
		FollowModel:        model.NewFollowModel(db.DB),
		FollowCountModel:   model.NewFollowCountModel(db.DB),
		BizRedis:           rds,
		NotificationPusher: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
	}
}
