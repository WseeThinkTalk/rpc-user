package logic

import (
	"context"

	"rpc-user/user/internal/svc"
	"rpc-user/user/service"

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

func (l *FindByIdLogic) FindById(in *service.FindByIdRequest) (*service.FindByIdResponse, error) {

	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		logx.Errorf("用户不存在")
		return nil, err
	}

	// 动态为没有头像的用户分配并持久化默认头像
	if user.Avatar == "" {
		defaultAvatar := "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=150&h=150"
		user.Avatar = defaultAvatar
		if err := l.svcCtx.UserModel.Update(l.ctx, user); err != nil {
			logx.Errorf("自动更新用户默认头像失败: %v", err)
		}
	}

	return &service.FindByIdResponse{
		UserId:       int64(user.Id),
		Username:     user.Username,
		Avatar:       user.Avatar,
		Mobile:       user.Mobile,
		Role:         user.Role,
		DisplayId:    user.DisplayId,
		Bio:          user.Bio,
		Gender:       int32(user.Gender),
		ProfileCover: user.ProfileCover,
	}, nil
}
