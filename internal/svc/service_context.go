package svc

import (
	"rpc-user/internal/config"
	followmodel "rpc-user/internal/model/follow"
	membermodel "rpc-user/internal/model/member"
	usermodel "rpc-user/internal/model/user"
	"rpc-user/pkg/orm"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *orm.DB
	BizRedis           *redis.Redis
	NotificationPusher *kq.Pusher

	UserModel        usermodel.UserModel
	MemberModel      *membermodel.MemberModel
	MemberOrderModel *membermodel.MemberOrderModel
	FollowModel      *followmodel.FollowModel
	FollowCountModel *followmodel.FollowCountModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dsn := c.DB.DataSource
	if dsn == "" {
		dsn = c.DataSource
	}

	db := orm.MustNewPostgres(&orm.Config{
		DSN:          dsn,
		MaxOpenConns: c.DB.MaxOpenConns,
		MaxIdleConns: c.DB.MaxIdleConns,
		MaxLifetime:  c.DB.MaxLifetime,
	})

	var rds *redis.Redis
	if c.BizRedis.Host != "" {
		rds = redis.MustNewRedis(redis.RedisConf{
			Host:        c.BizRedis.Host,
			Pass:        c.BizRedis.Pass,
			Type:        c.BizRedis.Type,
			PingTimeout: 10000000000,
		})
	}

	var pusher *kq.Pusher
	if len(c.KqPusherConf.Brokers) > 0 && c.KqPusherConf.Topic != "" {
		pusher = kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic)
	}

	return &ServiceContext{
		Config:             c,
		DB:                 db,
		BizRedis:           rds,
		NotificationPusher: pusher,

		UserModel:        usermodel.NewUserModel(db.DB),
		MemberModel:      membermodel.NewMemberModel(db.DB),
		MemberOrderModel: membermodel.NewMemberOrderModel(db.DB),
		FollowModel:      followmodel.NewFollowModel(db.DB),
		FollowCountModel: followmodel.NewFollowCountModel(db.DB),
	}
}
