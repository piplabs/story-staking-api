package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CLTotalStake struct {
	ID uint64 `gorm:"primarykey"`

	UpdateAt         int64 `gorm:"not null;column:update_at;index:idx_cl_total_stake_update_at,unique"`
	TotalStakeAmount int64 `gorm:"not null;column:total_stake_amount"`
}

func (CLTotalStake) TableName() string {
	return "cl_total_stakes"
}

func GetLatestCLTotalStake(db *gorm.DB) (*CLTotalStake, error) {
	var stake CLTotalStake

	if err := db.Order("update_at DESC").First(&stake).Error; err != nil {
		return nil, err
	}

	return &stake, nil
}

func GetCLTotalStakes(db *gorm.DB) ([]*CLTotalStake, error) {
	var stakes []*CLTotalStake

	if err := db.Order("update_at ASC").Find(&stakes).Error; err != nil {
		return nil, err
	}

	return stakes, nil
}

func GetLatestCLTotalStakeBefore(db *gorm.DB, timestamp int64) (*CLTotalStake, error) {
	var stake CLTotalStake

	if err := db.Where("update_at <= ?", timestamp).Order("update_at DESC").First(&stake).Error; err != nil {
		return nil, err
	}

	return &stake, nil
}

func GetCLTotalStakesAfter(db *gorm.DB, timestamp int64) ([]*CLTotalStake, error) {
	var stakes []*CLTotalStake

	if err := db.Where("update_at > ?", timestamp).Order("update_at ASC").Find(&stakes).Error; err != nil {
		return nil, err
	}

	return stakes, nil
}

func InsertCLTotalStake(db *gorm.DB, indexer string, stake *CLTotalStake) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "update_at"}},
			DoNothing: true,
		}).Create(stake).Error; err != nil {
			return err
		}

		return nil
	})
}

func BatchUpsertCLTotalStakes(db *gorm.DB, indexer string, stakes []*CLTotalStake, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "update_at"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_stake_amount": gorm.Expr("excluded.total_stake_amount"),
			}),
		}).CreateInBatches(stakes, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}
