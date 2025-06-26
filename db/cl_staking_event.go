package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type CLStakingEvent struct {
	ID          uint64 `gorm:"primarykey"`
	ELTxHash    string `gorm:"not null;column:el_tx_hash;index:idx_cl_staking_event_el_tx_hash_event_type,priority:1"`
	EventType   string `gorm:"not null;column:event_type;index:idx_cl_staking_event_el_tx_hash_event_type,priority:2"`
	BlockHeight int64  `gorm:"not null;column:block_height;index:idx_cl_staking_event_block_height"`
	StatusOK    bool   `gorm:"not null;column:status_ok"`
	ErrorCode   string `gorm:"not null;column:error_code"`
	Amount      string `gorm:"not null;column:amount"`
}

type CLSuccessfulStakingEvent struct {
	CLStakingEvent
	BlockTime time.Time `gorm:"not null;column:block_time"`
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

func GetSuccessfulCLStakingEventsAfter(db *gorm.DB, eventTypes []string, blockHeight int64) ([]*CLSuccessfulStakingEvent, error) {
	var events []*CLSuccessfulStakingEvent

	if err := db.
		Table("cl_staking_events AS e").
		Select("e.*, b.time AS block_time").
		Joins("JOIN cl_blocks AS b ON e.block_height = b.height").
		Where("e.event_type IN ?", eventTypes).
		Where("e.status_ok = ?", true).
		Where("e.block_height > ?", blockHeight).
		Scan(&events).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return events, nil
	} else if err != nil {
		return nil, err
	}

	return events, nil
}
