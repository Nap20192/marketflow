package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"marketflow/internal"
	"marketflow/internal/adapters/primary/ui"
	"marketflow/pkg/logger"
)

func usage() {
	fmt.Println(`Usage:
  marketflow [--port <N>]
  marketflow --help

Options:
  --port N     Port number`)
}

func main() {
	port := flag.String("port", "8080", "")
	debug := flag.Bool("debug", false, "")
	flag.Usage = usage
	flag.Parse()

	if *debug {
		logger.InitLogger("debug")
	} else {
		logger.InitLogger("info")
	}

	serverConfig, err := ui.NewServerConfig(port)
	if err != nil {
		slog.Error("invalid port", "error", err)
		os.Exit(1)
	}

	slog.Info("MarketFlow starting up")
	app := internal.NewApp(serverConfig)
	app.Run(context.Background())
}
