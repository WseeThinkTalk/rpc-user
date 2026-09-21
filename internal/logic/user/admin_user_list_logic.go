package userlogic

import (
	"context"

	"rpc-user/internal/svc"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserListLogic {
	return &AdminUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminUserListLogic) AdminUserList(in *user.AdminUserListRequest) (resp *user.AdminUserListResponse, err error) {
	resp = new(user.AdminUserListResponse)
	resp.Data = new(user.AdminUserListData)
	resp.Data.Users = make([]*user.UserInfoData, 0)

	users, err := l.svcCtx.UserModel.FindAll(l.ctx, in.Keyword, in.Cursor, in.PageSize)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = nil
		return resp, nil
	}

	var isEnd bool
	if len(users) < int(in.PageSize) {
		isEnd = true
	}

	nextCursor := in.Cursor
	for _, u := range users {
		resp.Data.Users = append(resp.Data.Users, &user.UserInfoData{
			UserId:       int64(u.Id),
			Username:     u.Username,
			Mobile:       u.Mobile,
			Avatar:       u.Avatar,
			Role:         u.Role,
			DisplayId:    u.DisplayId,
			Bio:          u.Bio,
			Gender:       int32(u.Gender),
			ProfileCover: u.ProfileCover,
		})
		nextCursor = int64(u.Id)
	}

	resp.Data.Cursor = nextCursor
	resp.Data.IsEnd = isEnd

	return resp, nil
}
