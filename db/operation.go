package db

import (
	"gorm.io/gorm"
)

type Operation struct {
	TxHash              string `gorm:"column:tx_hash" json:"tx_hash"`
	BlockHeight         int64  `gorm:"column:block_height" json:"block_height"`
	EventType           string `gorm:"column:event_type" json:"event_type"`
	Address             string `gorm:"column:address" json:"address"`
	SrcValidatorAddress string `gorm:"column:src_validator_address" json:"src_validator_address"`
	DstValidatorAddress string `gorm:"column:dst_validator_address" json:"dst_validator_address"`
	DstAddress          string `gorm:"column:dst_address" json:"dst_address"`

	StatusOK  bool   `gorm:"column:status_ok" json:"status_ok"`
	ErrorCode string `gorm:"column:error_code" json:"error_code"`
	Amount    string `gorm:"column:amount;type:numeric" json:"amount"`
}

func GetOperations(db *gorm.DB, evmAddr string, page, perPage int) ([]*Operation, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	query := db.Table("el_staking_events AS el").
		Joins("INNER JOIN cl_staking_events AS cl ON el.tx_hash = cl.el_tx_hash").
		Where("el.address = ?", evmAddr)

	// Perform the count query
	var totalOperations int64
	if err := query.Count(&totalOperations).Error; err != nil {
		return nil, 0, err
	}

	// Perform the paginated query
	var operations []*Operation
	if err := query.
		Select(`
			el.tx_hash AS tx_hash,
			el.block_height AS block_height,
			el.event_type,
			el.address,
			el.src_validator_address,
			el.dst_validator_address,
			el.dst_address,
			cl.status_ok,
			cl.error_code,
			cl.amount
		`).
		Order("el.block_height DESC").
		Limit(perPage).
		Offset(offset).
		Scan(&operations).Error; err != nil {
		return nil, 0, err
	}

	return operations, totalOperations, nil
}
