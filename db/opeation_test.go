package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
	"github.com/piplabs/story-staking-api/pkg/indexer"
)

func TestOperation(t *testing.T) {
	dbOperator, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	dbConn, err := dbOperator.DB()
	require.NoError(t, err)
	defer dbConn.Close()

	require.NoError(t, dbOperator.AutoMigrate(&db.CLStakingEvent{}))
	require.NoError(t, dbOperator.AutoMigrate(&db.ELStakingEvent{}))
	require.NoError(t, dbOperator.AutoMigrate(&db.IndexPoint{}))

	clIndexerName := "cl_staking_event"
	require.NoError(t, db.SetupIndexPoint(dbOperator, &db.IndexPoint{
		Indexer:     clIndexerName,
		BlockHeight: 0,
	}))

	elIndexerName := "el_staking_event"
	require.NoError(t, db.SetupIndexPoint(dbOperator, &db.IndexPoint{
		Indexer:     elIndexerName,
		BlockHeight: 0,
	}))

	t.Run("success event", func(t *testing.T) {
		elStakingEvents := []*db.ELStakingEvent{
			{
				TxHash:              "tx_hash1",
				BlockHeight:         1,
				EventType:           indexer.TypeStake,
				Address:             "address1",
				SrcValidatorAddress: "",
				DstValidatorAddress: "dst_validator_address1",
				DstAddress:          "",
			},
		}
		require.NoError(t, db.BatchCreateELStakingEvents(dbOperator, elIndexerName, elStakingEvents, 1))

		clStakingEvents := []*db.CLStakingEvent{
			{
				ELTxHash:    "tx_hash1",
				BlockHeight: 3,
				StatusOK:    true,
				ErrorCode:   "",
				Amount:      "1000000000000000000",
			},
		}
		require.NoError(t, db.BatchCreateCLStakingEvents(dbOperator, clIndexerName, clStakingEvents, 3))

		events, total, err := db.GetOperations(dbOperator, "address1", 1, 100)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, 1, len(events))

		require.Equal(t, "tx_hash1", events[0].TxHash)
		require.Equal(t, int64(1), events[0].BlockHeight)
		require.Equal(t, true, events[0].StatusOK)
		require.Equal(t, "", events[0].ErrorCode)
		require.Equal(t, "1000000000000000000", events[0].Amount)
	})

	t.Run("failure event", func(t *testing.T) {
		elStakingEvents := []*db.ELStakingEvent{
			{
				TxHash:              "tx_hash2",
				BlockHeight:         2,
				EventType:           indexer.TypeStake,
				Address:             "address2",
				SrcValidatorAddress: "",
				DstValidatorAddress: "dst_validator_address1",
				DstAddress:          "",
			},
		}
		require.NoError(t, db.BatchCreateELStakingEvents(dbOperator, elIndexerName, elStakingEvents, 2))

		clStakingEvents := []*db.CLStakingEvent{
			{
				ELTxHash:    "tx_hash2",
				BlockHeight: 4,
				StatusOK:    false,
				ErrorCode:   "Unspecified",
				Amount:      "1000000000000000000",
			},
		}
		require.NoError(t, db.BatchCreateCLStakingEvents(dbOperator, clIndexerName, clStakingEvents, 4))

		events, total, err := db.GetOperations(dbOperator, "address2", 1, 100)
		require.NoError(t, err)
		require.Equal(t, int64(1), total)
		require.Equal(t, 1, len(events))

		require.Equal(t, "tx_hash2", events[0].TxHash)
		require.Equal(t, int64(2), events[0].BlockHeight)
		require.Equal(t, false, events[0].StatusOK)
		require.Equal(t, "Unspecified", events[0].ErrorCode)
		require.Equal(t, "1000000000000000000", events[0].Amount)
	})
}
