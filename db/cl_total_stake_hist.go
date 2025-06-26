package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CLTotalStakeHist struct {
	ID uint64 `gorm:"primarykey"`

	TotalStakeAmount int64 `gorm:"not null;column:total_stake_amount"`
	UpdatedAtBlock   int64 `gorm:"not null;column:updated_at_block;index:idx_cl_total_stake_hist_updated_at_block,unique"`
	UpdatedAtTime    int64 `gorm:"not null;column:updated_at_time;index:idx_cl_total_stake_hist_updated_at_time,unique"`
}

func (CLTotalStakeHist) TableName() string {
	return "cl_total_stake_hists"
}

func GetLatestCLTotalStakeHist(db *gorm.DB) (*CLTotalStakeHist, error) {
	var stake CLTotalStakeHist

	if err := db.Order("updated_at_time DESC").First(&stake).Error; err != nil {
		return nil, err
	}

	return &stake, nil
}

func GetCLTotalStakeHists(db *gorm.DB) ([]*CLTotalStakeHist, error) {
	var stakes []*CLTotalStakeHist

	if err := db.Order("updated_at_time ASC").Find(&stakes).Error; err != nil {
		return nil, err
	}

	return stakes, nil
}

func GetLatestCLTotalStakeHistBefore(db *gorm.DB, timestamp int64) (*CLTotalStakeHist, error) {
	var stake CLTotalStakeHist

	if err := db.Where("updated_at_time <= ?", timestamp).Order("updated_at_time DESC").First(&stake).Error; err != nil {
		return nil, err
	}

	return &stake, nil
}

func GetCLTotalStakeHistsAfter(db *gorm.DB, timestamp int64) ([]*CLTotalStakeHist, error) {
	var stakes []*CLTotalStakeHist

	if err := db.Where("updated_at_time > ?", timestamp).Order("updated_at_time ASC").Find(&stakes).Error; err != nil {
		return nil, err
	}

	return stakes, nil
}

func UpsertCLGenesisTotalStakeHist(db *gorm.DB, indexer string, stake *CLTotalStakeHist) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "updated_at_time"}},
			DoNothing: true,
		}).Create(stake).Error; err != nil {
			return err
		}

		return nil
	})
}

func BatchUpsertCLTotalStakeHists(db *gorm.DB, indexer string, stakes []*CLTotalStakeHist) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "updated_at_time"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"total_stake_amount": gorm.Expr("excluded.total_stake_amount"),
			}),
		}).CreateInBatches(stakes, 100).Error; err != nil {
			return err
		}

		return nil
	})
}
