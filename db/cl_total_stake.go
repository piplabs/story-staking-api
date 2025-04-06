package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CLTotalStake struct {
	ID uint64 `gorm:"primarykey"`

	CreatedAt        int64 `gorm:"not null;column:created_at;index:idx_cl_created_at_total_stake_amount,priority:1"`
	TotalStakeAmount int64 `gorm:"not null;column:total_stake_amount;index:idx_cl_created_at_total_stake_amount,priority:2"`
}

func (CLTotalStake) TableName() string {
	return "cl_total_stakes"
}

func GetLatestCLTotalStake(db *gorm.DB) (*CLTotalStake, error) {
	var stake CLTotalStake

	if err := db.Order("created_at DESC").First(&stake).Error; err != nil {
		return nil, err
	}

	return &stake, nil
}

func InsertCLTotalStake(db *gorm.DB, indexer string, stake *CLTotalStake) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "created_at"}},
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
			Columns: []clause.Column{{Name: "created_at"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_stake_amount": gorm.Expr("excluded.total_stake_amount"),
			}),
		}).CreateInBatches(stakes, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}
