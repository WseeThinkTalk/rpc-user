package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"rpc-user/follow/code"
	"rpc-user/follow/internal/model"
	"rpc-user/follow/internal/svc"
	"rpc-user/follow/internal/types"
	"rpc-user/follow/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
	"gorm.io/gorm"
)

type UnFollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnFollowLogic {
	return &UnFollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UnFollow 取消关注
func (l *UnFollowLogic) UnFollow(in *pb.UnFollowRequest) (*pb.UnFollowResponse, error) {
	if in.UserId == 0 {
		return nil, code.FollowUserIdEmpty
	}
	if in.FollowedUserId == 0 {
		return nil, code.FollowedUserIdEmpty
	}

	follow, err := l.svcCtx.FollowModel.FindByUserIDAndFollowedUserID(l.ctx, in.UserId, in.FollowedUserId)
	if err != nil {
		l.Logger.Errorf("[UnFollow] FollowModel.FindByUserIDAndFollowedUserID err: %v req: %v", err, in)
		return nil, err
	}
	if follow == nil {
		return &pb.UnFollowResponse{}, nil
	}
	if follow.FollowStatus == types.FollowStatusUnfollow {
		return &pb.UnFollowResponse{}, nil
	}

	// 事务
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		err := model.NewFollowModel(tx).UpdateFields(l.ctx, follow.ID, map[string]interface{}{
			"follow_status": types.FollowStatusUnfollow,
		})
		if err != nil {
			return err
		}
		err = model.NewFollowCountModel(tx).DecrFollowCount(l.ctx, in.UserId)
		if err != nil {
			return err
		}
		return model.NewFollowCountModel(tx).DecrFansCount(l.ctx, in.FollowedUserId)
	})
	if err != nil {
		l.Logger.Errorf("[UnFollow] Transaction error: %v", err)
		return nil, err
	}
	_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, userFollowKey(in.UserId), strconv.FormatInt(in.FollowedUserId, 10))
	if err != nil {
		l.Logger.Errorf("[UnFollow] BizRedis.ZremCtx error: %v", err)
		return nil, err
	}
	_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, userFansKey(in.FollowedUserId), strconv.FormatInt(in.UserId, 10))
	if err != nil {
		l.Logger.Errorf("[UnFollow] BizRedis.ZremCtx error: %v", err)
		return nil, err
	}

	// 异步发送取消关注通知
	threading.GoSafe(func() {
		notif := map[string]interface{}{
			"userId":        in.FollowedUserId,
			"type":          int32(2),
			"title":         "取消关注",
			"content":       "取消了对你的关注",
			"refId":         in.UserId,
			"bizId":         fmt.Sprintf("follow:%d:%d", in.UserId, in.FollowedUserId),
			"triggerUserId": in.UserId,
		}
		data, err := json.Marshal(notif)
		if err != nil {
			l.Logger.Errorf("[UnFollow] marshal notification err: %v", err)
			return
		}
		if err := l.svcCtx.NotificationPusher.Push(context.Background(), string(data)); err != nil {
			l.Logger.Errorf("[UnFollow] push notification err: %v", err)
		}
	})

	return &pb.UnFollowResponse{}, nil
}
