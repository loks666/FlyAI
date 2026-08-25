package watergroup

import (
	"github.com/loks666/FlyAI/client"
	"gorm.io/gorm"
	"time"
)

// rankUser
// @description: 排行榜用户
type rankUser struct {
	GroupUser     string // 微信Id
	Nickname      string // 昵称
	Count         int64  // 消息数
	TextCount     int64  // 文本
	ImageCount    int64  // 图片
	VoiceCount    int64  // 语音
	VideoCount    int64  // 视频和小视频
	EmoticonCount int64  // 表情包
	ShareCount    int64  // 分享、转发、名片
	OtherCount    int64  // 位置等其他用户内容
}

var rankMessageTypes = []int{1, 3, 34, 42, 43, 47, 48, 49, 62}

// getRankData
// @description: 获取消息排行榜
// @param groupId string 群Id
// @param d string 模式(yesterday | week | month)
// @return rank
// @return err
func getRankData(groupId, date string, reference ...time.Time) (rank []rankUser, err error) {
	currentMemberQuery := `(SELECT group_id, wxid, MAX(nickname) AS nickname
		FROM t_group_user
		GROUP BY group_id, wxid
		HAVING MAX(is_member) = 1 AND MAX(skip_chat_rank) = 0) AS tgu`
	tx := client.MySQL.Table("t_message AS tm").
		Joins("JOIN "+currentMemberQuery+" ON tgu.wxid = tm.group_user AND tm.from_user = tgu.group_id").
		Select(`tm.group_user, tgu.nickname, COUNT(1) AS count,
			SUM(CASE WHEN tm.type = 1 THEN 1 ELSE 0 END) AS text_count,
			SUM(CASE WHEN tm.type = 3 THEN 1 ELSE 0 END) AS image_count,
			SUM(CASE WHEN tm.type = 34 THEN 1 ELSE 0 END) AS voice_count,
			SUM(CASE WHEN tm.type IN (43, 62) THEN 1 ELSE 0 END) AS video_count,
			SUM(CASE WHEN tm.type = 47 THEN 1 ELSE 0 END) AS emoticon_count,
			SUM(CASE WHEN tm.type IN (42, 49) THEN 1 ELSE 0 END) AS share_count,
			SUM(CASE WHEN tm.type = 48 THEN 1 ELSE 0 END) AS other_count`).
		Where("tm.from_user = ?", groupId).
		Where("tm.type IN (?)", rankMessageTypes).
		Group("tm.group_user, tgu.nickname").
		Order("`count` DESC")

	now := time.Now()
	if len(reference) > 0 {
		now = reference[0]
	}
	tx = applyRankDateFilter(tx, date, now)

	// 查询指定时间段全部数据
	err = tx.Find(&rank).Error
	return
}

func applyRankDateFilter(tx *gorm.DB, date string, now time.Time) *gorm.DB {
	switch date {
	case "yesterday":
		return tx.Where("DATEDIFF(tm.create_at,NOW()) = -1")
	case "week":
		start, end := rollingWeekRange(now)
		return tx.Where("tm.create_at >= ? AND tm.create_at < ?", start, end)
	case "month":
		return tx.Where("PERIOD_DIFF(date_format(now(), '%Y%m'), date_format(tm.create_at, '%Y%m')) = 1")
	case "year":
		return tx.Where("YEAR(tm.create_at) = YEAR(NOW()) - 1")
	default:
		return tx
	}
}

func rollingWeekRange(now time.Time) (start, end time.Time) {
	end = now.In(time.Local)
	return end.AddDate(0, 0, -7), end
}
