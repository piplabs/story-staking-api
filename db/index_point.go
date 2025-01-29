package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IndexPoint struct {
	ID          uint64 `gorm:"primarykey"`
	Indexer     string `gorm:"not null;column:indexer;uniqueIndex:idx_index_point_indexer"`
	BlockHeight int64  `gorm:"not null;column:block_height"`
}

func (IndexPoint) TableName() string {
	return "index_points"
}

func SetupIndexPoint(db *gorm.DB, indexPoint *IndexPoint) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "indexer"}},
		DoNothing: true,
	}).Create(indexPoint).Error
}

func GetIndexPoint(db *gorm.DB, indexer string) (*IndexPoint, error) {
	var indexPoint IndexPoint
	if err := db.Where("indexer = ?", indexer).First(&indexPoint).Error; err != nil {
		return nil, err
	}

	return &indexPoint, nil
}

func UpdateIndexPoint(db *gorm.DB, indexer string, blockHeight int64) error {
	return db.Model(&IndexPoint{}).Where("indexer = ?", indexer).Update("block_height", blockHeight).Error
}
