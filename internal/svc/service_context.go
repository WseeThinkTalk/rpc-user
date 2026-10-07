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
	db := orm.MustNewPostgres(&orm.Config{
		DSN:          c.DB.DataSource,
		MaxOpenConns: c.DB.MaxOpenConns,
		MaxIdleConns: c.DB.MaxIdleConns,
		MaxLifetime:  c.DB.MaxLifetime,
	})

	rds := redis.MustNewRedis(c.BizRedis)

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
