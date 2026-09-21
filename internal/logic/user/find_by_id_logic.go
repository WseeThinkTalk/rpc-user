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

func (l *FindByIdLogic) FindById(in *user.FindByIdRequest) (resp *user.FindByIdResponse, err error) {
	resp = new(user.FindByIdResponse)
	resp.Data = new(user.UserInfoData)

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

	resp.Data.UserId = int64(u.Id)
	resp.Data.Username = u.Username
	resp.Data.Avatar = u.Avatar
	resp.Data.Mobile = u.Mobile
	resp.Data.Role = u.Role
	resp.Data.DisplayId = u.DisplayId
	resp.Data.Bio = u.Bio
	resp.Data.Gender = int32(u.Gender)
	resp.Data.ProfileCover = u.ProfileCover

	return resp, nil
}
