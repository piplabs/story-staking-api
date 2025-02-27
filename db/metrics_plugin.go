package db

import (
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/pkg/metrics"
)

type MetricsPlugin struct{}

func (p *MetricsPlugin) Name() string {
	return "metrics_plugin"
}

func (p *MetricsPlugin) Initialize(db *gorm.DB) (err error) {
	db.Callback().Create().
		After("gorm:after_create").
		Register("metrics_plugin:after_create", afterOperation)

	db.Callback().Query().
		After("gorm:after_query").
		Register("metrics_plugin:after_query", afterOperation)

	db.Callback().Update().
		After("gorm:after_update").
		Register("metrics_plugin:after_update", afterOperation)

	db.Callback().Delete().
		After("gorm:after_delete").
		Register("metrics_plugin:after_delete", afterOperation)

	db.Callback().Row().
		After("gorm:row").
		Register("metrics_plugin:after_row", afterOperation)

	db.Callback().Raw().
		After("gorm:raw").
		Register("metrics_plugin:after_raw", afterOperation)

	return nil

}

func afterOperation(db *gorm.DB) {
	if db.Error != nil {
		metrics.DBErrorCounter.WithLabelValues().Inc()
	}
}
