package logic

import (
	"context"

	"rpc-user/user/internal/svc"
	"rpc-user/user/service"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProfileLogic) UpdateProfile(in *service.UpdateProfileRequest) (*service.UpdateProfileResponse, error) {
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		logx.Errorf("Find user error: %v", err)
		return nil, err
	}

	user.Username = in.Username
	if in.Avatar != "" {
		user.Avatar = in.Avatar
	}
	user.Bio = in.Bio
	user.Gender = int64(in.Gender)
	user.ProfileCover = in.ProfileCover

	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		logx.Errorf("Update user error: %v", err)
		return nil, err
	}

	return &service.UpdateProfileResponse{}, nil
}
