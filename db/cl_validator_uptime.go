package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CLValidatorUptime struct {
	ID         uint64 `gorm:"primarykey"`
	EVMAddress string `gorm:"not null;column:evm_address;index:idx_cl_validator_uptime_evm_address,unique"` // To lower case
	ActiveFrom int64  `gorm:"not null;column:active_from"`
	ActiveTo   int64  `gorm:"not null;column:active_to"`
	VoteCount  int64  `gorm:"not null;column:vote_count"`
}

func (CLValidatorUptime) TableName() string {
	return "cl_validator_uptimes"
}

func BatchUpsertCLValidatorUptime(db *gorm.DB, indexer string, clUptimes []*CLValidatorUptime, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, uptime := range clUptimes {
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "evm_address"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"active_from": gorm.Expr(`
						CASE
							WHEN cl_validator_uptimes.active_to + 1 = excluded.active_from THEN cl_validator_uptimes.active_from
							ELSE excluded.active_from
						END`),
					"active_to": gorm.Expr(`excluded.active_to`),
					"vote_count": gorm.Expr(`
						CASE
							WHEN cl_validator_uptimes.active_to + 1 = excluded.active_from THEN cl_validator_uptimes.vote_count + excluded.vote_count
							ELSE excluded.vote_count
						END`),
				}),
			}).Create(&uptime).Error

			if err != nil {
				return err
			}
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}

func GetCLValidatorUptimes(db *gorm.DB, evmAddrs ...string) ([]*CLValidatorUptime, error) {
	var results []*CLValidatorUptime
	if err := db.Where("evm_address IN ?", evmAddrs).Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
