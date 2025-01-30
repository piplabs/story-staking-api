package server

import (
	"context"
	"errors"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"

	"github.com/piplabs/story-indexer/cache"
	"github.com/piplabs/story-indexer/db"
	"github.com/piplabs/story-indexer/pkg/indexer"
)

type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	conf   *Config
	wg     *sync.WaitGroup
	sf     *singleflight.Group // TODO: use singleflight for all database queries

	rootDir string

	dbOperator    *gorm.DB
	cacheOperator *redis.Client

	ginService *gin.Engine
	httpServer *http.Server
	indexers   []indexer.Indexer
}

func NewServer(ctx context.Context, dir string, conf *Config) (*Server, error) {
	ctxS, cancelS := context.WithCancel(ctx)

	svr := &Server{
		ctx:    ctxS,
		cancel: cancelS,
		conf:   conf,
		wg:     &sync.WaitGroup{},
		sf:     &singleflight.Group{},

		rootDir: dir,
	}

	if err := svr.initServices(); err != nil {
		cancelS()
		return nil, err
	}

	return svr, nil
}

func (s *Server) Run() {
	switch s.conf.Server.IndexMode {
	case IndexModeReader:
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("story-indexer server stopped")
			}
		}()
	case IndexModeWriter:
		for _, indexer := range s.indexers {
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				indexer.Run()
			}()
		}
	default:
		log.Fatal().Str("index_mode", s.conf.Server.IndexMode).Msg("invalid index mode")
	}
}

func (s *Server) GracefulQuit() error {
	s.cancel()
	if s.conf.Server.IndexMode == IndexModeReader {
		_ = s.httpServer.Shutdown(context.Background())
	}
	s.wg.Wait()

	_ = s.cacheOperator.Close()
	connPool, err := s.dbOperator.DB()
	if err != nil {
		return err
	}
	_ = connPool.Close()

	return nil
}

func (s *Server) initServices() error { // TODO: get pwd from secret manager
	// Connect to database.
	postgresClient, err := db.NewPostgresClient(s.ctx, filepath.Join(s.rootDir, s.conf.Database.ConfigFile))
	if err != nil {
		return err
	}
	s.dbOperator = postgresClient

	// Connect to cache.
	redisClient, err := cache.NewRedisClient(s.ctx, filepath.Join(s.rootDir, s.conf.Cache.ConfigFile))
	if err != nil {
		return err
	}
	s.cacheOperator = redisClient

	// Setup gin service engine.
	s.setupGinService()

	// Setup indexers.
	s.setupIndexers()

	// Setup database states for `writer` mode.
	if s.conf.Server.IndexMode == IndexModeWriter {
		s.dbOperator.AutoMigrate(&db.CLBlock{})
		s.dbOperator.AutoMigrate(&db.CLStakingEvent{})
		s.dbOperator.AutoMigrate(&db.CLValidatorUptime{})
		s.dbOperator.AutoMigrate(&db.ELBlock{})
		s.dbOperator.AutoMigrate(&db.ELReward{})
		s.dbOperator.AutoMigrate(&db.ELStakingEvent{})
		s.dbOperator.AutoMigrate(&db.IndexPoint{})

		// Initialize genesis index points.
		for _, indexer := range s.indexers {
			if err := db.SetupIndexPoint(s.dbOperator, &db.IndexPoint{
				Indexer:     indexer.Name(),
				BlockHeight: 0,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Server) setupGinService() {
	gin.SetMode(s.conf.Server.ServiceMode)

	s.ginService = gin.New()
	s.ginService.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	s.ginService.Use(gin.Logger())
	s.ginService.Use(gin.Recovery())

	apiGroup := s.ginService.Group("/api")
	{
		// Indexer APIs.
		apiGroup.GET("/network_status", s.NetworkStatusHandler())
		apiGroup.GET("/estimated_apr", s.EstimatedAPRHandler())
		apiGroup.GET("/operations/:evm_address", s.OperationsHandler())
		apiGroup.GET("/rewards/:evm_address", s.RewardsHandler())
		// Proxy to Story API.
		apiGroup.GET("/staking/pool", s.StakingPoolHandler())

		apiGroup.GET("/staking/validators", s.StakingValidatorsHandler())
		apiGroup.GET("/staking/validators/:validator_address", s.StakingValidatorHandler())
		apiGroup.GET("/staking/validators/:validator_address/delegations", s.StakingValidatorDelegationsHandler())
		apiGroup.GET("/staking/validators/:validator_address/delegations/:delegator_address", s.StakingDelegationHandler())
		apiGroup.GET("/staking/validators/:validator_address/delegators/:delegator_address/period_delegations", s.StakingValidatorDelegatorPeriodDelegationsHandler())
		apiGroup.GET("/staking/validators/:validator_address/delegators/:delegator_address/period_delegations/:period_delegation_id", s.StakingValidatorDelegatorPeriodDelegationHandler())

		apiGroup.GET("/staking/delegations/:delegator_address", s.StakingDelegatorDelegationsHandler())

		apiGroup.GET("/staking/delegators/:delegator_address/unbonding_delegations", s.StakingDelegatorUnbondingDelegationsHandler())
	}

	s.httpServer = &http.Server{
		Addr:    s.conf.Server.ServicePort,
		Handler: s.ginService,
	}
}

func (s *Server) setupIndexers() error {
	clBlockIndexer, err := indexer.NewCLBlockIndexer(s.ctx, s.dbOperator, s.cacheOperator, s.conf.Blockchain.ConsensusChainID, s.conf.Blockchain.CometbftRPCEndpoint)
	if err != nil {
		return err
	}
	s.indexers = append(s.indexers, clBlockIndexer)

	clStakingEventIndexer, err := indexer.NewCLStakingEventIndexer(s.ctx, s.dbOperator, s.cacheOperator, s.conf.Blockchain.ConsensusChainID, s.conf.Blockchain.CometbftRPCEndpoint)
	if err != nil {
		return err
	}
	s.indexers = append(s.indexers, clStakingEventIndexer)

	clValidatorUptimeIndexer, err := indexer.NewCLValidatorUptimeIndexer(s.ctx, s.dbOperator, s.cacheOperator, s.conf.Blockchain.ConsensusChainID, s.conf.Blockchain.CometbftRPCEndpoint)
	if err != nil {
		return err
	}
	s.indexers = append(s.indexers, clValidatorUptimeIndexer)

	elBlockIndexer, err := indexer.NewELBlockIndexer(s.ctx, s.dbOperator, s.cacheOperator, s.conf.Blockchain.GethRPCEndpoint)
	if err != nil {
		return err
	}
	s.indexers = append(s.indexers, elBlockIndexer)

	elRewardIndexer, err := indexer.NewELRewardIndexer(s.ctx, s.dbOperator, s.cacheOperator, s.conf.Blockchain.GethRPCEndpoint)
	if err != nil {
		return err
	}
	s.indexers = append(s.indexers, elRewardIndexer)

	elStakingEventIndexer, err := indexer.NewELStakingEventIndexer(s.ctx, s.dbOperator, s.cacheOperator, s.conf.Blockchain.GethRPCEndpoint)
	if err != nil {
		return err
	}
	s.indexers = append(s.indexers, elStakingEventIndexer)

	return nil
}
