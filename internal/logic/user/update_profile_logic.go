package userlogic

import (
	"rpc-user/pkg/code"
	"context"

	"rpc-user/internal/model/user"
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

func (l *UpdateProfileLogic) UpdateProfile(in *user.UpdateProfileRequest) (resp *user.UpdateProfileResponse, err error) {
	resp = new(user.UpdateProfileResponse)

	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		if err == model.ErrNotFound {
			resp.Code = int64(code.NotFound.Code())
			resp.Msg = "用户不存在"
			return resp, nil
		}
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
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
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	return resp, nil
}
