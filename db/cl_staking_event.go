package db

import "gorm.io/gorm"

type CLStakingEvent struct {
	ID          uint64 `gorm:"primarykey"`
	ELTxHash    string `gorm:"not null;column:el_tx_hash;index:idx_cl_staking_event_el_tx_hash_event_type,priority:1"`
	EventType   string `gorm:"not null;column:event_type;index:idx_cl_staking_event_el_tx_hash_event_type,priority:2"`
	BlockHeight int64  `gorm:"not null;column:block_height;index:idx_cl_staking_event_block_height"`
	StatusOK    bool   `gorm:"not null;column:status_ok"`
	ErrorCode   string `gorm:"not null;column:error_code"`
	Amount      string `gorm:"not null;column:amount"`
}

func (CLStakingEvent) TableName() string {
	return "cl_staking_events"
}

func BatchCreateCLStakingEvents(db *gorm.DB, indexer string, events []*CLStakingEvent, height int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(events, 100).Error; err != nil {
			return err
		}

		return UpdateIndexPoint(tx, indexer, height)
	})
}
