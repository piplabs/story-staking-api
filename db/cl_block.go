package db

import (
	"time"

	"gorm.io/gorm"
)

type CLBlock struct {
	ID              uint64    `gorm:"primarykey"`
	Height          int64     `gorm:"not null;column:height;uniqueIndex:idx_cl_block_height"`
	Hash            string    `gorm:"not null;column:hash;uniqueIndex:idx_cl_block_hash"`
	ProposerAddress string    `gorm:"not null;column:proposer_address"`
	Time            time.Time `gorm:"not null;column:time"`
}

func (CLBlock) TableName() string {
	return "cl_blocks"
}

func BatchCreateCLBlocks(db *gorm.DB, indexer string, blocks []*CLBlock, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(blocks, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}

func GetLatestCLBlock(db *gorm.DB) (*CLBlock, error) {
	var clBlk CLBlock
	if err := db.Order("height DESC").First(&clBlk).Error; err != nil {
		return nil, err
	}

	return &clBlk, nil
}
