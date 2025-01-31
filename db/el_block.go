package db

import (
	"time"

	"gorm.io/gorm"
)

type ELBlock struct {
	ID       uint64    `gorm:"primarykey"`
	Height   int64     `gorm:"not null;column:height;index:idx_el_block_height,unique"`
	Hash     string    `gorm:"not null;column:hash;index:idx_el_block_hash,unique"`
	GasUsed  uint64    `gorm:"not null"`
	GasLimit uint64    `gorm:"not null"`
	Time     time.Time `gorm:"not null;column:time"`
}

func (ELBlock) TableName() string {
	return "el_blocks"
}

func BatchCreateELBlocks(db *gorm.DB, indexer string, blocks []*ELBlock, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(blocks, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}

func GetLatestELBlock(db *gorm.DB, n int) ([]*ELBlock, error) {
	var elBlks []*ELBlock
	if err := db.Order("height DESC").Limit(n).Find(&elBlks).Error; err != nil {
		return nil, err
	}

	return elBlks, nil
}
