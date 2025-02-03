package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

func TestCLValidatorUptime(t *testing.T) {
	dbOperator, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	dbConn, err := dbOperator.DB()
	require.NoError(t, err)
	defer dbConn.Close()

	require.NoError(t, dbOperator.AutoMigrate(&db.CLValidatorUptime{}))
	require.NoError(t, dbOperator.AutoMigrate(&db.IndexPoint{}))

	indexerName := "cl_validator_uptime"
	require.NoError(t, db.SetupIndexPoint(dbOperator, &db.IndexPoint{
		Indexer:     indexerName,
		BlockHeight: 0,
	}))

	t.Run("uptime update", func(t *testing.T) {
		initialUptimes := []*db.CLValidatorUptime{
			{
				EVMAddress: "0x1111",
				ActiveFrom: 1,
				ActiveTo:   10,
				VoteCount:  5,
			},
			{
				EVMAddress: "0x2222",
				ActiveFrom: 1,
				ActiveTo:   10,
				VoteCount:  10,
			},
		}
		require.NoError(t, db.BatchUpsertCLValidatorUptime(dbOperator, indexerName, initialUptimes, 10))

		uptimes, err := db.GetCLValidatorUptimes(dbOperator, "0x1111")
		require.NoError(t, err)
		require.Equal(t, 1, len(uptimes))
		require.Equal(t, int64(1), uptimes[0].ActiveFrom)
		require.Equal(t, int64(10), uptimes[0].ActiveTo)
		require.Equal(t, int64(5), uptimes[0].VoteCount)

		uptimes, err = db.GetCLValidatorUptimes(dbOperator, "0x2222")
		require.NoError(t, err)
		require.Equal(t, 1, len(uptimes))
		require.Equal(t, int64(1), uptimes[0].ActiveFrom)
		require.Equal(t, int64(10), uptimes[0].ActiveTo)
		require.Equal(t, int64(10), uptimes[0].VoteCount)

		followingUptimes := []*db.CLValidatorUptime{
			{
				EVMAddress: "0x1111",
				ActiveFrom: 11,
				ActiveTo:   20,
				VoteCount:  10,
			},
			{
				EVMAddress: "0x2222",
				ActiveFrom: 12,
				ActiveTo:   20,
				VoteCount:  8,
			},
			{
				EVMAddress: "0x3333",
				ActiveFrom: 11,
				ActiveTo:   20,
				VoteCount:  10,
			},
		}
		require.NoError(t, db.BatchUpsertCLValidatorUptime(dbOperator, indexerName, followingUptimes, 20))

		uptimes, err = db.GetCLValidatorUptimes(dbOperator, "0x1111")
		require.NoError(t, err)
		require.Equal(t, 1, len(uptimes))
		require.Equal(t, int64(1), uptimes[0].ActiveFrom)
		require.Equal(t, int64(20), uptimes[0].ActiveTo)
		require.Equal(t, int64(15), uptimes[0].VoteCount)

		uptimes, err = db.GetCLValidatorUptimes(dbOperator, "0x2222")
		require.NoError(t, err)
		require.Equal(t, 1, len(uptimes))
		require.Equal(t, int64(12), uptimes[0].ActiveFrom)
		require.Equal(t, int64(20), uptimes[0].ActiveTo)
		require.Equal(t, int64(8), uptimes[0].VoteCount)

		uptimes, err = db.GetCLValidatorUptimes(dbOperator, "0x3333")
		require.NoError(t, err)
		require.Equal(t, 1, len(uptimes))
		require.Equal(t, int64(11), uptimes[0].ActiveFrom)
		require.Equal(t, int64(20), uptimes[0].ActiveTo)
		require.Equal(t, int64(10), uptimes[0].VoteCount)
	})
}
