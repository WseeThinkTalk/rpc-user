package followlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	model "rpc-user/internal/model/follow"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/follow"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/threading"
	"gorm.io/gorm"
)

// checkAndTrimZSetScript 仅当缓存 Key 存在时，原子追加新成员并按容量修剪多余成员
// 解决传统 Exists -> Zadd -> Zremrangebyrank 非原子 3 次 RTT 导致的冷缓存截断竞态与残缺缓存污染
var checkAndTrimZSetScript = redis.NewScript(`
local key = KEYS[1]
if redis.call("EXISTS", key) == 1 then
    local score = tonumber(ARGV[1])
    local member = ARGV[2]
    local maxCount = tonumber(ARGV[3])
    redis.call("ZADD", key, score, member)
    local trimEnd = -(maxCount + 1)
    redis.call("ZREMRANGEBYRANK", key, 0, trimEnd)
    return 1
end
return 0
`)

func atomicAddAndTrimZSet(ctx context.Context, rds *redis.Redis, key string, score int64, member string, maxCount int) error {
	if rds == nil {
		return nil
	}
	_, err := rds.ScriptRunCtx(ctx, checkAndTrimZSetScript, []string{key}, score, member, maxCount)
	return err
}

// UpdateFollowCache 更新关注与粉丝缓存（提供给业务逻辑与基准测试调用）
func UpdateFollowCache(ctx context.Context, rds *redis.Redis, userId, followedUserId int64) error {
	nowUnix := time.Now().Unix()
	_ = atomicAddAndTrimZSet(ctx, rds, userFollowKey(userId), nowUnix, strconv.FormatInt(followedUserId, 10), types.CacheMaxFollowCount)
	_ = atomicAddAndTrimZSet(ctx, rds, userFansKey(followedUserId), nowUnix, strconv.FormatInt(userId, 10), types.CacheMaxFansCount)
	return nil
}

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Follow 关注指定用户
func (l *FollowLogic) Follow(in *user.FollowRequest) (resp *user.FollowResponse, err error) {
	resp = new(user.FollowResponse)

	if in.UserId == 0 {
		resp.Code = int64(code.FollowUserIdEmpty.Code())
		resp.Msg = code.FollowUserIdEmpty.Message()
		return resp, nil
	}
	if in.FollowedUserId == 0 {
		resp.Code = int64(code.FollowedUserIdEmpty.Code())
		resp.Msg = code.FollowedUserIdEmpty.Message()
		return resp, nil
	}
	if in.UserId == in.FollowedUserId {
		resp.Code = int64(code.CannotFollowSelf.Code())
		resp.Msg = code.CannotFollowSelf.Message()
		return resp, nil
	}

	// 查询已有关注记录
	follow, err := l.svcCtx.FollowModel.FindByUserIDAndFollowedUserID(l.ctx, in.UserId, in.FollowedUserId)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}
	if follow != nil && follow.FollowStatus == types.FollowStatusFollow {
		return resp, nil
	}

	// 事务更新关注关系与计数
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		if follow != nil {
			err = model.NewFollowModel(tx).UpdateFields(l.ctx, follow.ID, map[string]interface{}{
				"follow_status": types.FollowStatusFollow,
			})
		} else {
			err = model.NewFollowModel(tx).Insert(l.ctx, &model.Follow{
				UserID:         in.UserId,
				FollowedUserID: in.FollowedUserId,
				FollowStatus:   types.FollowStatusFollow,
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			})
		}

		if err != nil {
			return err
		}
		err = model.NewFollowCountModel(tx).IncrFollowCount(l.ctx, in.UserId)
		if err != nil {
			return err
		}
		return model.NewFollowCountModel(tx).IncrFansCount(l.ctx, in.FollowedUserId)
	})
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	// 原子更新关注与粉丝缓存（仅当缓存存在时原子追加并修剪，防止残缺冷缓存污染与 3 次 RTT 竞态）
	_ = UpdateFollowCache(l.ctx, l.svcCtx.BizRedis, in.UserId, in.FollowedUserId)

	// 异步发送关注通知
	threading.GoSafe(func() {
		notif := map[string]interface{}{
			"userId":        in.FollowedUserId,
			"type":          int32(2),
			"title":         "关注通知",
			"content":       "关注了你",
			"refId":         in.UserId,
			"bizId":         fmt.Sprintf("follow:%d:%d", in.UserId, in.FollowedUserId),
			"triggerUserId": in.UserId,
		}
		data, err := json.Marshal(notif)
		if err != nil {
			return
		}
		_ = l.svcCtx.NotificationPusher.Push(context.Background(), string(data))
	})

	return resp, nil
}

func userFollowKey(userId int64) string {
	return fmt.Sprintf("biz#user#follow#%d", userId)
}

func userFansKey(userId int64) string {
	return fmt.Sprintf("biz#user#fans#%d", userId)
}
