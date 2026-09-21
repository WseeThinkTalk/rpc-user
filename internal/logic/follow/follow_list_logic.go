package followlogic

import (
	"context"
	"strconv"
	"time"

	model "rpc-user/internal/model/follow"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/follow"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

const userFollowExpireTime = 3600 * 24 * 2

type FollowListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowListLogic {
	return &FollowListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowListLogic) FollowList(in *user.FollowListRequest) (resp *user.FollowListResponse, err error) {
	resp = new(user.FollowListResponse)
	resp.Data = new(user.FollowListData)
	resp.Data.Items = make([]*user.FollowItem, 0)

	if in.UserId == 0 {
		resp.Code = int64(code.UserIdEmpty.Code())
		resp.Msg = code.UserIdEmpty.Message()
		return resp, nil
	}
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}
	if in.Cursor == 0 {
		in.Cursor = time.Now().Unix()
	}

	var (
		isCache, isEnd  bool
		lastId, cursor  int64
		followedUserIds []int64
		follows         []*model.Follow
		curPage         []*user.FollowItem
	)

	followUserIds, _ := l.cacheFollowUserIds(l.ctx, in.UserId, in.Cursor, in.PageSize)
	if len(followUserIds) > 0 {
		isCache = true
		if followUserIds[len(followUserIds)-1] == -1 {
			followUserIds = followUserIds[:len(followUserIds)-1]
			isEnd = true
		}
		if len(followUserIds) == 0 {
			return resp, nil
		}
		follows, err = l.svcCtx.FollowModel.FindByFollowedUserIds(l.ctx, in.UserId, followUserIds)
		if err != nil {
			resp.Code = int64(code.ServerErr.Code())
			resp.Msg = err.Error()
			return resp, nil
		}
		// 组装缓存中的关注列表
		for _, v := range follows {
			followedUserIds = append(followedUserIds, v.FollowedUserID)
			curPage = append(curPage, &user.FollowItem{
				Id:             v.ID,
				FollowedUserId: v.FollowedUserID,
				CreateTime:     v.CreateTime.Unix(),
			})
		}
	} else {
		follows, err = l.svcCtx.FollowModel.FindByUserId(l.ctx, in.UserId, types.CacheMaxFollowCount)
		if err != nil {
			resp.Code = int64(code.ServerErr.Code())
			resp.Msg = err.Error()
			return resp, nil
		}
		if len(follows) == 0 {
			return resp, nil
		}
		var firstPageFollows []*model.Follow
		if len(follows) > int(in.PageSize) {
			firstPageFollows = follows[:in.PageSize]
		} else {
			firstPageFollows = follows
			isEnd = true
		}
		// 组装第一页关注列表数据
		for _, v := range firstPageFollows {
			followedUserIds = append(followedUserIds, v.FollowedUserID)
			curPage = append(curPage, &user.FollowItem{
				Id:             v.ID,
				FollowedUserId: v.FollowedUserID,
				CreateTime:     v.CreateTime.Unix(),
			})
		}
	}
	if len(curPage) > 0 {
		pageLast := curPage[len(curPage)-1]
		lastId = pageLast.Id
		cursor = pageLast.CreateTime
		if cursor < 0 {
			cursor = 0
		}
		// 根据游标匹配当前分页起始位置
		for k, v := range curPage {
			if v.CreateTime == in.Cursor && v.Id == in.Id {
				curPage = curPage[k:]
				break
			}
		}
	}
	fc, err := l.svcCtx.FollowCountModel.FindByUserIds(l.ctx, followedUserIds)
	_ = err
	uidFansCount := make(map[int64]int)
	// 汇总各关注用户的粉丝数
	for _, v := range fc {
		uidFansCount[v.UserID] = v.FansCount
	}
	// 回填各关注项的粉丝数统计
	for _, v := range curPage {
		v.FansCount = int64(uidFansCount[v.FollowedUserId])
	}
	resp.Data.IsEnd = isEnd
	resp.Data.Cursor = cursor
	resp.Data.Id = lastId
	if curPage != nil {
		resp.Data.Items = curPage
	}

	if !isCache {
		threading.GoSafe(func() {
			if len(follows) < types.CacheMaxFollowCount && len(follows) > 0 {
				follows = append(follows, &model.Follow{FollowedUserID: -1})
			}
			_ = l.addCacheFollow(context.Background(), in.UserId, follows)
		})
	}

	return resp, nil
}

func (l *FollowListLogic) cacheFollowUserIds(ctx context.Context, userId, cursor, pageSize int64) ([]int64, error) {
	key := userFollowKey(userId)
	b, err := l.svcCtx.BizRedis.ExistsCtx(ctx, key)
	if err == nil && b {
		_ = l.svcCtx.BizRedis.ExpireCtx(ctx, key, userFollowExpireTime)
	}
	pairs, err := l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, 0, cursor, 0, int(pageSize))
	if err != nil {
		return nil, err
	}
	var uids []int64
	// 解析 Redis ZSet 元素中的关注用户 ID
	for _, v := range pairs {
		uid, err := strconv.ParseInt(v.Key, 10, 64)
		if err != nil {
			continue
		}
		uids = append(uids, uid)
	}

	return uids, nil
}

func (l *FollowListLogic) addCacheFollow(ctx context.Context, userId int64, follows []*model.Follow) error {
	if len(follows) == 0 {
		return nil
	}
	key := userFollowKey(userId)
	// 批量写入关注缓存到 Redis ZSet
	for _, v := range follows {
		var score int64
		if v.FollowedUserID == -1 {
			score = 0
		} else {
			score = v.CreateTime.Unix()
		}
		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.FormatInt(v.FollowedUserID, 10))
		if err != nil {
			return err
		}
	}

	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, userFollowExpireTime)
}
