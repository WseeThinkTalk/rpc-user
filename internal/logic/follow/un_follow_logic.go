package followlogic

import (
	"context"
	"strconv"

	model "rpc-user/internal/model/follow"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/follow"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *UnFollowLogic) UnFollow(in *user.UnFollowRequest) (resp *user.UnFollowResponse, err error) {
	resp = new(user.UnFollowResponse)

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

	follow, err := l.svcCtx.FollowModel.FindByUserIDAndFollowedUserID(l.ctx, in.UserId, in.FollowedUserId)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}
	if follow == nil || follow.FollowStatus == types.FollowStatusUnfollow {
		return resp, nil
	}

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
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	followExist, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, userFollowKey(in.UserId))
	if err == nil && followExist {
		_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, userFollowKey(in.UserId), strconv.FormatInt(in.FollowedUserId, 10))
	}
	fansExist, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, userFansKey(in.FollowedUserId))
	if err == nil && fansExist {
		_, _ = l.svcCtx.BizRedis.ZremCtx(l.ctx, userFansKey(in.FollowedUserId), strconv.FormatInt(in.UserId, 10))
	}

	return resp, nil
}
