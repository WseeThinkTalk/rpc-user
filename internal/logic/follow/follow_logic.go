package followlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	model "rpc-user/internal/model/follow"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/follow"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"gorm.io/gorm"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowLogic) Follow(in *user.FollowRequest) (resp *user.FollowResponse, err error) {
	resp = new(user.FollowResponse)
	resp.Code = 200
	resp.Msg = "success"

	if in.UserId == 0 {
		resp.Code = int64(code.FollowUserIdEmpty.Code())
		resp.Msg = code.FollowUserIdEmpty.Message()
		return resp, nil
	}
	if in.FollowedUserId == 0 {
		resp.Code = int64(code.FollowedUserIdEmpty.Code())
		resp.Msg = code.FollowedUserIdEmpty.Message()
		return resp, nil
	}
	if in.UserId == in.FollowedUserId {
		resp.Code = int64(code.CannotFollowSelf.Code())
		resp.Msg = code.CannotFollowSelf.Message()
		return resp, nil
	}
	follow, err := l.svcCtx.FollowModel.FindByUserIDAndFollowedUserID(l.ctx, in.UserId, in.FollowedUserId)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, nil
	}
	if follow != nil && follow.FollowStatus == types.FollowStatusFollow {
		return resp, nil
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if follow != nil {
			err = model.NewFollowModel(tx).UpdateFields(l.ctx, follow.ID, map[string]interface{}{
				"follow_status": types.FollowStatusFollow,
			})
		} else {
			err = model.NewFollowModel(tx).Insert(l.ctx, &model.Follow{
				UserID:         in.UserId,
				FollowedUserID: in.FollowedUserId,
				FollowStatus:   types.FollowStatusFollow,
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			})
		}

		if err != nil {
			return err
		}
		err = model.NewFollowCountModel(tx).IncrFollowCount(l.ctx, in.UserId)
		if err != nil {
			return err
		}
		return model.NewFollowCountModel(tx).IncrFansCount(l.ctx, in.FollowedUserId)
	})
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, nil
	}

	followExist, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, userFollowKey(in.UserId))
	if err == nil && followExist {
		_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, userFollowKey(in.UserId), time.Now().Unix(), strconv.FormatInt(in.FollowedUserId, 10))
		_, _ = l.svcCtx.BizRedis.ZremrangebyrankCtx(l.ctx, userFollowKey(in.UserId), 0, -(types.CacheMaxFollowCount + 1))
	}
	fansExist, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, userFansKey(in.FollowedUserId))
	if err == nil && fansExist {
		_, _ = l.svcCtx.BizRedis.ZaddCtx(l.ctx, userFansKey(in.FollowedUserId), time.Now().Unix(), strconv.FormatInt(in.UserId, 10))
		_, _ = l.svcCtx.BizRedis.ZremrangebyrankCtx(l.ctx, userFansKey(in.FollowedUserId), 0, -(types.CacheMaxFansCount + 1))
	}

	threading.GoSafe(func() {
		notif := map[string]interface{}{
			"userId":        in.FollowedUserId,
			"type":          int32(2),
			"title":         "关注通知",
			"content":       "关注了你",
			"refId":         in.UserId,
			"bizId":         fmt.Sprintf("follow:%d:%d", in.UserId, in.FollowedUserId),
			"triggerUserId": in.UserId,
		}
		data, err := json.Marshal(notif)
		if err != nil {
			return
		}
		_ = l.svcCtx.NotificationPusher.Push(context.Background(), string(data))
	})

	return resp, nil
}

func userFollowKey(userId int64) string {
	return fmt.Sprintf("biz#user#follow#%d", userId)
}

func userFansKey(userId int64) string {
	return fmt.Sprintf("biz#user#fans#%d", userId)
}
