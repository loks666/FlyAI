package service

import (
	"errors"
	"github.com/loks666/FlyAI/client"
	"github.com/loks666/FlyAI/model/entity"
	"time"
)

const (
	FeatureNews    = "news"
	FeatureWeekly  = "weekly"
	FeatureYearly  = "yearly"
	FeatureDormant = "dormant"

	ScheduleDaily     = "daily"
	ScheduleMonday    = "monday"
	ScheduleTuesday   = "tuesday"
	ScheduleWednesday = "wednesday"
	ScheduleThursday  = "thursday"
	ScheduleFriday    = "friday"
	ScheduleSaturday  = "saturday"
	ScheduleSunday    = "sunday"
	ScheduleMonthly   = "monthly"
	ScheduleYearly    = "yearly"
)

type featureDefault struct {
	SendTime         string
	Schedule         string
	MessageThreshold int
}

var featureDefaults = map[string]featureDefault{
	FeatureNews:    {SendTime: "08:00", Schedule: ScheduleDaily},
	FeatureWeekly:  {SendTime: "09:30", Schedule: ScheduleMonday},
	FeatureYearly:  {SendTime: "00:00", Schedule: ScheduleYearly},
	FeatureDormant: {SendTime: "09:00", Schedule: ScheduleMonday, MessageThreshold: 1},
}

var scheduleWeekdays = map[string]time.Weekday{
	ScheduleMonday:    time.Monday,
	ScheduleTuesday:   time.Tuesday,
	ScheduleWednesday: time.Wednesday,
	ScheduleThursday:  time.Thursday,
	ScheduleFriday:    time.Friday,
	ScheduleSaturday:  time.Saturday,
	ScheduleSunday:    time.Sunday,
}

var scheduleLabels = map[string]string{
	ScheduleDaily:     "每天",
	ScheduleMonday:    "每周一",
	ScheduleTuesday:   "每周二",
	ScheduleWednesday: "每周三",
	ScheduleThursday:  "每周四",
	ScheduleFriday:    "每周五",
	ScheduleSaturday:  "每周六",
	ScheduleSunday:    "每周日",
	ScheduleMonthly:   "每月 1 日",
	ScheduleYearly:    "每年 1 月 1 日",
}

func ValidateFeatureTime(value string) error {
	_, err := time.Parse("15:04", value)
	if err != nil || len(value) != 5 {
		return errors.New("时间格式必须为 HH:mm")
	}
	return nil
}

func ValidateFeatureSchedule(value string) error {
	if value == ScheduleDaily || value == ScheduleMonthly || value == ScheduleYearly {
		return nil
	}
	if _, ok := scheduleWeekdays[value]; !ok {
		return errors.New("推送周期必须为每天、每周一至周日、每月或每年")
	}
	return nil
}

func FeatureScheduleLabel(schedule string) string {
	if label, ok := scheduleLabels[schedule]; ok {
		return label
	}
	return scheduleLabels[ScheduleDaily]
}

func FeatureScheduleMatches(schedule string, now time.Time) bool {
	switch schedule {
	case ScheduleDaily:
		return true
	case ScheduleMonthly:
		return now.Day() == 1
	case ScheduleYearly:
		return now.Month() == time.January && now.Day() == 1
	}
	weekday, ok := scheduleWeekdays[schedule]
	return ok && now.Weekday() == weekday
}

func GetFeatureSetting(feature string) (entity.FeatureSetting, error) {
	defaults, ok := featureDefaults[feature]
	if !ok {
		return entity.FeatureSetting{}, errors.New("未知功能")
	}
	setting := entity.FeatureSetting{Feature: feature, SendTime: defaults.SendTime, Schedule: defaults.Schedule, MessageThreshold: defaults.MessageThreshold}
	result := client.MySQL.FirstOrCreate(&setting, entity.FeatureSetting{Feature: feature})
	err := result.Error
	if err == nil && setting.Schedule == "" {
		setting.Schedule = defaults.Schedule
		err = client.MySQL.Model(&setting).Update("schedule", setting.Schedule).Error
	}
	if err == nil && feature == FeatureYearly && (setting.Schedule != ScheduleYearly || setting.SendTime != "00:00") {
		setting.Schedule = ScheduleYearly
		setting.SendTime = "00:00"
		err = client.MySQL.Model(&setting).Updates(map[string]any{"schedule": setting.Schedule, "send_time": setting.SendTime}).Error
	}
	if err == nil && feature == FeatureDormant && setting.MessageThreshold < 1 {
		setting.MessageThreshold = defaults.MessageThreshold
		err = client.MySQL.Model(&setting).Update("message_threshold", setting.MessageThreshold).Error
	}
	return setting, err
}

func UpdateFeatureSetting(feature, schedule, sendTime string, messageThreshold *int) (entity.FeatureSetting, error) {
	if _, ok := featureDefaults[feature]; !ok {
		return entity.FeatureSetting{}, errors.New("未知功能")
	}
	if feature == FeatureYearly {
		return entity.FeatureSetting{}, errors.New("年排行榜固定于每年 1 月 1 日 00:00 推送，无需修改")
	}
	if err := ValidateFeatureTime(sendTime); err != nil {
		return entity.FeatureSetting{}, err
	}
	if err := ValidateFeatureSchedule(schedule); err != nil {
		return entity.FeatureSetting{}, err
	}
	if messageThreshold != nil {
		if feature != FeatureDormant {
			return entity.FeatureSetting{}, errors.New("只有潜水统计支持消息条数阈值")
		}
		if *messageThreshold < 1 || *messageThreshold > 100000 {
			return entity.FeatureSetting{}, errors.New("潜水消息阈值必须为 1～100000；配置 1 表示只统计 0 条")
		}
	}
	setting, err := GetFeatureSetting(feature)
	if err != nil {
		return setting, err
	}
	updates := map[string]any{"schedule": schedule, "send_time": sendTime}
	if messageThreshold != nil {
		updates["message_threshold"] = *messageThreshold
		setting.MessageThreshold = *messageThreshold
	}
	err = client.MySQL.Model(&entity.FeatureSetting{}).
		Where("feature = ?", feature).
		Updates(updates).Error
	setting.SendTime = sendTime
	setting.Schedule = schedule
	return setting, err
}

// ClaimFeatureRun atomically claims today's scheduled run for one feature.
func ClaimFeatureRun(feature string, now time.Time) (bool, error) {
	setting, err := GetFeatureSetting(feature)
	if err != nil {
		return false, err
	}
	if !FeatureScheduleMatches(setting.Schedule, now) || setting.SendTime != now.Format("15:04") {
		return false, nil
	}
	today := now.Format("2006-01-02")
	result := client.MySQL.Model(&entity.FeatureSetting{}).
		Where("feature = ?", feature).
		Where("last_run_date <> ? OR last_run_date = ''", today).
		Update("last_run_date", today)
	return result.RowsAffected == 1, result.Error
}
