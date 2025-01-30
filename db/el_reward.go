package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ELReward struct {
	ID               uint64 `gorm:"primarykey"`
	Address          string `gorm:"not null;column:address;index:idx_el_reward_address,unique"` // To lower case
	Amount           string `gorm:"not null;column:amount;type:numeric"`
	LastUpdateHeight int64  `gorm:"not null;column:last_update_height"`
}

func (ELReward) TableName() string {
	return "el_rewards"
}

func BatchUpsertELRewards(db *gorm.DB, indexer string, rewards []*ELReward, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "address"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"amount":             gorm.Expr("el_rewards.amount + excluded.amount"),
				"last_update_height": gorm.Expr("GREATEST(el_rewards.last_update_height, excluded.last_update_height)"),
			}),
		}).CreateInBatches(rewards, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}

func GetRewards(db *gorm.DB, evmAddr string) (*ELReward, error) {
	var reward ELReward
	if err := db.Where("address = ?", evmAddr).First(&reward).Error; err != nil {
		return nil, err
	}

	return &reward, nil
}
