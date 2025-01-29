package db

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CLValidatorUptime struct {
	ID         uint64 `gorm:"primarykey"`
	EVMAddress string `gorm:"not null;column:evm_address;uniqueIndex:idx_cl_uptime_evm_address"` // To lower case
	ActiveFrom int64  `gorm:"not null;column:active_from"`
	ActiveTo   int64  `gorm:"not null;column:active_to"`
	VoteCount  int64  `gorm:"not null;column:vote_count"`
}

func (CLValidatorUptime) TableName() string {
	return "cl_uptimes"
}

func BatchUpsertCLUptime(db *gorm.DB, indexer string, clUptimes []*CLValidatorUptime, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, uptime := range clUptimes {
			err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "evm_address"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"active_from": gorm.Expr(`
                        CASE
                            WHEN active_to + 1 != ? THEN ?
                            ELSE active_from
                        END`, uptime.ActiveFrom, uptime.ActiveFrom),
					"active_to": gorm.Expr(`
                        CASE
                            WHEN active_to + 1 = ? THEN ?
                            ELSE ?
                        END`, uptime.ActiveFrom, uptime.ActiveTo, uptime.ActiveTo),
					"vote_count": gorm.Expr(`
                        CASE
                            WHEN active_to + 1 = ? THEN vote_count + ?
                            ELSE ?
                        END`, uptime.ActiveFrom, uptime.VoteCount, uptime.VoteCount),
				}),
			}).Create(uptime).Error

			if err != nil {
				return err
			}
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}

func GetCLUptimes(db *gorm.DB, evmAddrs ...string) ([]*CLValidatorUptime, error) {
	var results []*CLValidatorUptime
	if err := db.Where("evm_address IN ?", evmAddrs).Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
