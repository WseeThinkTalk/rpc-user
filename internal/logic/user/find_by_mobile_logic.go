package userlogic

import (
	"context"

	model "rpc-user/internal/model/user"
	"rpc-user/internal/svc"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindByMobileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindByMobileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByMobileLogic {
	return &FindByMobileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindByMobileLogic) FindByMobile(in *user.FindByMobileRequest) (resp *user.FindByMobileResponse, err error) {
	resp = new(user.FindByMobileResponse)
	resp.Data = new(user.UserAuthData)

	u, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		if err == model.ErrNotFound {
			resp.Data = nil
			return resp, nil
		}
		logx.Errorf("FindByMobile mobile: %s error: %v", in.Mobile, err)
		return nil, err
	}

	resp.Data.UserId = int64(u.Id)
	resp.Data.Username = u.Username
	resp.Data.Avatar = u.Avatar
	resp.Data.Password = u.Password
	resp.Data.Role = u.Role

	return resp, nil
}
