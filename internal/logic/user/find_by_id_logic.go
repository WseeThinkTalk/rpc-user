package userlogic

import (
	"context"

	"rpc-user/internal/svc"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByIdLogic {
	return &FindByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindByIdLogic) FindById(in *user.FindByIdRequest) (*user.FindByIdResponse, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		logx.Errorf("用户不存在: %v", err)
		return nil, err
	}

	if u.Avatar == "" {
		defaultAvatar := "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=150&h=150"
		u.Avatar = defaultAvatar
		if err := l.svcCtx.UserModel.Update(l.ctx, u); err != nil {
			logx.Errorf("自动更新用户默认头像失败: %v", err)
		}
	}

	return &user.FindByIdResponse{
		Code: 200,
		Msg:  "success",
		Data: &user.UserInfoData{
			UserId:       int64(u.Id),
			Username:     u.Username,
			Avatar:       u.Avatar,
			Mobile:       u.Mobile,
			Role:         u.Role,
			DisplayId:    u.DisplayId,
			Bio:          u.Bio,
			Gender:       int32(u.Gender),
			ProfileCover: u.ProfileCover,
		},
	}, nil
}
