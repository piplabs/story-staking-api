package db

import "gorm.io/gorm"

type ELStakingEvent struct {
	ID                  uint64 `gorm:"primarykey"`
	TxHash              string `gorm:"not null;column:tx_hash;index:idx_el_staking_event_tx_hash_event_type,priority:1"`
	EventType           string `gorm:"not null;column:event_type;index:idx_el_staking_event_tx_hash_event_type,priority:2"`
	Address             string `gorm:"not null;column:address;index:idx_el_staking_event_address_block_height,priority:1"` // To lower case
	BlockHeight         int64  `gorm:"not null;column:block_height;index:idx_el_staking_event_address_block_height,priority:2"`
	SrcValidatorAddress string `gorm:"not null;column:src_validator_address"`
	DstValidatorAddress string `gorm:"not null;column:dst_validator_address"`
	DstAddress          string `gorm:"not null;column:dst_address"` // RewardAddrss | WithdrawAddress | OperatorAddress
}

func (ELStakingEvent) TableName() string {
	return "el_staking_events"
}

func BatchCreateELStakingEvents(db *gorm.DB, indexer string, events []*ELStakingEvent, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(events, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}
