package db

import (
	"time"

	"gorm.io/gorm"
)

type CLBlock struct {
	ID              uint64    `gorm:"primarykey"`
	Height          int64     `gorm:"not null;column:height;index:idx_cl_block_height,unique"`
	Hash            string    `gorm:"not null;column:hash;index:idx_cl_block_hash,unique"`
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

func GetCLBlocks(db *gorm.DB, heights []int64) ([]*CLBlock, error) {
	clBlocks := make([]*CLBlock, 0)

	if err := db.Where("height IN (?)", heights).Find(&clBlocks).Error; err != nil {
		return nil, err
	}

	return clBlocks, nil
}
