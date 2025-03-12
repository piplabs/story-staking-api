package db

import (
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/pkg/util"
)

type CLValidatorVote struct {
	ID          uint64 `gorm:"primarykey"`
	Validator   string `gorm:"not null;column:validator;index:idx_cl_validator_vote_validator_block_height,priority:1,unique"` // To lower case
	BlockHeight int64  `gorm:"not null;column:block_height;index:idx_cl_validator_vote_validator_block_height,priority:2,unique"`
}

func (CLValidatorVote) TableName() string {
	return "cl_validator_votes"
}

func BatchUpdateCLValidatorVotes(db *gorm.DB, indexer string, validatorVotes []*CLValidatorVote, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(validatorVotes, 100).Error; err != nil {
			return err
		}

		if err := tx.Where("block_height < ?", height-util.UptimeWindow+1).Delete(&CLValidatorVote{}).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}

func GetCLValidatorsVotes(db *gorm.DB, validators ...string) (map[string]int64, error) {
	var results []struct {
		Validator string
		Count     int64
	}

	if err := db.Table("cl_validator_votes").
		Select("validator, COUNT(*) as count").
		Where("validator IN ?", validators).
		Group("validator").
		Find(&results).Error; err != nil {
		return nil, err
	}

	validatorVoteCounts := make(map[string]int64)
	for _, res := range results {
		validatorVoteCounts[res.Validator] = res.Count
	}

	return validatorVoteCounts, nil
}
