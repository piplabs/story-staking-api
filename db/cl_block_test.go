package db_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/piplabs/story-staking-api/db"
)

func TestCLBlock(t *testing.T) {
	dbOperator, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	dbConn, err := dbOperator.DB()
	require.NoError(t, err)
	defer dbConn.Close()

	require.NoError(t, dbOperator.AutoMigrate(&db.CLBlock{}))
	require.NoError(t, dbOperator.AutoMigrate(&db.IndexPoint{}))

	indexerName := "cl_block"
	require.NoError(t, db.SetupIndexPoint(dbOperator, &db.IndexPoint{
		Indexer:     indexerName,
		BlockHeight: 0,
	}))

	t.Run("TestCreateAndGetCLBlocks", func(t *testing.T) {
		blocks := []*db.CLBlock{
			{Height: 1, Hash: "hash1", ProposerAddress: "address1", Time: time.Unix(100, 0)},
			{Height: 2, Hash: "hash2", ProposerAddress: "address2", Time: time.Unix(200, 0)},
		}
		require.NoError(t, db.BatchCreateCLBlocks(dbOperator, indexerName, blocks, 2))

		latest, err := db.GetLatestCLBlock(dbOperator)
		require.NoError(t, err)
		require.Equal(t, int64(2), latest.Height)
		require.Equal(t, "address2", latest.ProposerAddress)
		require.Equal(t, time.Unix(200, 0).Unix(), latest.Time.Unix())

		blocks = []*db.CLBlock{
			{Height: 3, Hash: "hash3", ProposerAddress: "address3", Time: time.Unix(300, 0)},
		}
		require.NoError(t, db.BatchCreateCLBlocks(dbOperator, indexerName, blocks, 3))

		latest, err = db.GetLatestCLBlock(dbOperator)
		require.NoError(t, err)
		require.Equal(t, int64(3), latest.Height)
		require.Equal(t, "address3", latest.ProposerAddress)
		require.Equal(t, time.Unix(300, 0).Unix(), latest.Time.Unix())
	})
}
