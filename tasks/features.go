package tasks

import (
	"github.com/loks666/FlyAI/service"
	"github.com/loks666/FlyAI/tasks/dormant"
	"github.com/loks666/FlyAI/tasks/news"
	"github.com/loks666/FlyAI/tasks/watergroup"
	"log"
	"time"
)

func RunTimedFeatures() {
	now := time.Now()
	features := []struct {
		key string
		run func()
	}{
		{key: service.FeatureNews, run: news.DailyNews},
		{key: service.FeatureWeekly, run: watergroup.Week},
		{key: service.FeatureDormant, run: dormant.Weekly},
	}
	for _, feature := range features {
		claimed, err := service.ClaimFeatureRun(feature.key, now)
		if err != nil {
			log.Printf("检查定时功能[%s]失败: %v", feature.key, err)
			continue
		}
		if claimed {
			feature.run()
		}
	}
}
