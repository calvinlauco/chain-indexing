package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crypto-com/chainindex/internal/primptr"

	applogger "github.com/crypto-com/chainindex/internal/logger"

	"github.com/crypto-com/chainindex/infrastructure"

	"github.com/crypto-com/chainindex/internal/filereader/toml"
	"github.com/urfave/cli/v2"
)

func CliApp(args []string) error {
	cliApp := &cli.App{
		Name:                 filepath.Base(args[0]),
		Usage:                "Crypto.com Chain Indexing Service",
		Version:              "v0.0.1",
		Copyright:            "(c) 2020 Crypto.com",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Value: "./config/config.toml",
				Usage: "TOML `FILE` to load configuration from",
			},

			&cli.StringFlag{
				Name:  "logLevel,l",
				Usage: "Log level (Allowed values: fatal,error,info,debug)",
			},
			&cli.BoolFlag{
				Name:  "color",
				Usage: "Display colored log",
			},

			&cli.BoolFlag{
				Name:    "dbSSL",
				Usage:   "Enable Postgres SSL mode",
				EnvVars: []string{"DB_SSL"},
			},
			&cli.StringFlag{
				Name:    "dbHost",
				Usage:   "Postgres database hostname",
				EnvVars: []string{"DB_HOST"},
			},
			&cli.UintFlag{
				Name:    "dbPort",
				Usage:   "Postgres database port",
				EnvVars: []string{"DB_PORT"},
			},
			&cli.StringFlag{
				Name:    "dbUsername",
				Usage:   "Postgres username",
				EnvVars: []string{"DB_USERNAME"},
			},
			&cli.StringFlag{
				Name:     "dbPassword",
				Usage:    "Postgres password",
				EnvVars:  []string{"DB_PASSWORD"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "dbName",
				Usage:   "Postgres database name",
				EnvVars: []string{"DB_NAME"},
			},
			&cli.StringFlag{
				Name:    "dbSchema",
				Usage:   "Postgres schema name",
				EnvVars: []string{"DB_SCHEMA"},
			},

			&cli.StringFlag{
				Name:    "tendermintURL",
				Usage:   "Tendermint HTTP RPC URL",
				EnvVars: []string{"TENDERMINT_URL"},
			},
		},
		Action: func(ctx *cli.Context) error {
			if args := ctx.Args(); args.Len() > 0 {
				return fmt.Errorf("Unexpected arguments: %q", args.Get(0))
			}

			// Prepare FileConfig
			configPath := ctx.String("config")
			configReader, configFileErr := toml.FromFile(configPath)
			if configFileErr != nil {
				return configFileErr
			}
			var fileConfig FileConfig
			readConfigErr := configReader.Read(&fileConfig)
			if readConfigErr != nil {
				return readConfigErr
			}

			cliConfig := CLIConfig{
				LogLevel: ctx.String("logLevel"),

				DatabaseHost:     ctx.String("dbHost"),
				DatabaseUsername: ctx.String("dbUsername"),
				DatabasePassword: ctx.String("dbPassword"),
				DatabaseName:     ctx.String("dbName"),
				DatabaseSchema:   ctx.String("dbSchema"),

				TendermintHTTPRPCURL: ctx.String("tendermintURL"),
			}
			if ctx.IsSet("color") {
				cliConfig.LoggerColor = primptr.Bool(ctx.Bool("color"))
			}
			if ctx.IsSet("dbSSL") {
				cliConfig.DatabaseSSL = primptr.Bool(ctx.Bool("dbSSL"))
			}
			if ctx.IsSet("dgPort") {
				cliConfig.DatabasePort = primptr.Int32(int32(ctx.Int("dbPort")))
			}

			config := Config{
				fileConfig,
			}
			config.OverrideByCLIConfig(&cliConfig)

			fmt.Printf("%+v", config)

			// Create logger
			logLevel := parseLogLevel(config.Logger.Level)
			logger := infrastructure.NewZerologLogger(os.Stdout)
			logger.SetLogLevel(logLevel)

			// Setup system
			rdbConn, err := SetupRDbConn(&config, logger)
			if err != nil {
				logger.Panicf("error setting up RDb connection: %v", err)
			}

			httpAPIServer := NewHTTPAPIServer(logger, rdbConn, &config)
			go func() {
				if err := httpAPIServer.Run(); err != nil {
					logger.Panicf("%v", err)
				}
			}()

			indexService := NewIndexService(logger, rdbConn, &config)
			go func() {
				if err := indexService.Run(); err != nil {
					logger.Panicf("%v", err)
				}
			}()

			select {}
		},
	}

	err := cliApp.Run(args)
	if err != nil {
		return err
	}

	return nil
}

func parseLogLevel(level string) applogger.LogLevel {
	switch level {
	case "panic":
		return applogger.LOG_LEVEL_PANIC
	case "error":
		return applogger.LOG_LEVEL_ERROR
	case "info":
		return applogger.LOG_LEVEL_INFO
	case "debug":
		return applogger.LOG_LEVEL_DEBUG
	default:
		panic("Unsupported log level: " + level)
	}
}
