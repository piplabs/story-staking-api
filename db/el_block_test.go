package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

func TestELBlock(t *testing.T) {
	dbOperator, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	dbConn, err := dbOperator.DB()
	require.NoError(t, err)
	defer dbConn.Close()

	require.NoError(t, dbOperator.AutoMigrate(&db.ELBlock{}))
	require.NoError(t, dbOperator.AutoMigrate(&db.IndexPoint{}))

	indexerName := "el_block"
	require.NoError(t, db.SetupIndexPoint(dbOperator, &db.IndexPoint{
		Indexer:     indexerName,
		BlockHeight: 0,
	}))

	t.Run("TestCreateAndGetELBlocks", func(t *testing.T) {
		blocks := []*db.ELBlock{
			{Height: 1, Hash: "hash1", GasUsed: 100, GasLimit: 1000, Time: time.Unix(100, 0)},
			{Height: 2, Hash: "hash2", GasUsed: 200, GasLimit: 1000, Time: time.Unix(200, 0)},
		}
		require.NoError(t, db.BatchCreateELBlocks(dbOperator, indexerName, blocks, 2))

		latest, err := db.GetLatestELBlock(dbOperator, 1)
		require.NoError(t, err)
		require.Equal(t, 1, len(latest))
		require.Equal(t, int64(2), latest[0].Height)
		require.Equal(t, uint64(200), latest[0].GasUsed)
		require.Equal(t, uint64(1000), latest[0].GasLimit)
		require.Equal(t, time.Unix(200, 0).Unix(), latest[0].Time.Unix())

		blocks = []*db.ELBlock{
			{Height: 3, Hash: "hash3", GasUsed: 300, GasLimit: 1000, Time: time.Unix(300, 0)},
		}
		require.NoError(t, db.BatchCreateELBlocks(dbOperator, indexerName, blocks, 3))

		latest, err = db.GetLatestELBlock(dbOperator, 1)
		require.NoError(t, err)
		require.Equal(t, int64(3), latest[0].Height)
		require.Equal(t, uint64(300), latest[0].GasUsed)
		require.Equal(t, uint64(1000), latest[0].GasLimit)
		require.Equal(t, time.Unix(300, 0).Unix(), latest[0].Time.Unix())
	})
}
