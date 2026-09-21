package userlogic

import (
	"context"

	"rpc-user/internal/model/user"
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
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = new(user.UserInfoData)

	u, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		if err == model.ErrNotFound {
			resp.Code = 404
			resp.Msg = "用户不存在"
			resp.Data = nil
			return resp, nil
		}
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = nil
		return resp, nil
	}

	if u.Avatar == "" {
		defaultAvatar := "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?auto=format&fit=crop&w=150&h=150"
		u.Avatar = defaultAvatar
		_ = l.svcCtx.UserModel.Update(l.ctx, u)
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
