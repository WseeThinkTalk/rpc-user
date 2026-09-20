package userlogic

import (
	"context"

	"rpc-user/internal/svc"
	"rpc-user/user"

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

func (l *UpdateProfileLogic) UpdateProfile(in *user.UpdateProfileRequest) (*user.UpdateProfileResponse, error) {
	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		logx.Errorf("Find user error: %v", err)
		return nil, err
	}

	u.Username = in.Username
	if in.Avatar != "" {
		u.Avatar = in.Avatar
	}
	u.Bio = in.Bio
	u.Gender = int64(in.Gender)
	u.ProfileCover = in.ProfileCover

	err = l.svcCtx.UserModel.Update(l.ctx, u)
	if err != nil {
		logx.Errorf("Update user error: %v", err)
		return nil, err
	}

	return &user.UpdateProfileResponse{
		Code: 200,
		Msg:  "success",
	}, nil
}
