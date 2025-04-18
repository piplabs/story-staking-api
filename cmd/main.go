package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/rs/zerolog/log"
	_ "go.uber.org/automaxprocs"

	"github.com/piplabs/story-staking-api/pkg/server"
)

var (
	home   = kingpin.Flag("home", "Home directory").Default(".").String()
	config = kingpin.Flag("config", "Config file path").Default("config.toml").String()
)

func main() {
	kingpin.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configFile, err := os.Open(*config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open config file")
	}
	defer configFile.Close()

	var svrConfig server.Config
	if _, err := toml.NewDecoder(configFile).Decode(&svrConfig); err != nil {
		log.Fatal().Err(err).Msg("Failed to decode config file")
	}

	if err := svrConfig.Validate(); err != nil {
		log.Fatal().Err(err).Msg("invalid config")
	}

	svr, err := server.NewServer(ctx, *home, &svrConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("new story-staking-api server failed")
	}

	svr.Run()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-ch
	log.Error().Msgf("received signal %v, quiting gracefully", sig)
	_ = svr.GracefulQuit()
}
