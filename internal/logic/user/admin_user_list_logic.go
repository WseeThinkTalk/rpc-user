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

func (l *AdminUserListLogic) AdminUserList(in *user.AdminUserListRequest) (*user.AdminUserListResponse, error) {
	users, err := l.svcCtx.UserModel.FindAll(l.ctx, in.Keyword, in.Cursor, in.PageSize)
	if err != nil {
		return nil, err
	}

	var isEnd bool
	if len(users) < int(in.PageSize) {
		isEnd = true
	}

	var items []*user.UserInfoData
	var nextCursor int64 = in.Cursor
	for _, u := range users {
		items = append(items, &user.UserInfoData{
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

	return &user.AdminUserListResponse{
		Code: 200,
		Msg:  "success",
		Data: &user.AdminUserListData{
			Users:  items,
			Cursor: nextCursor,
			IsEnd:  isEnd,
		},
	}, nil
}
