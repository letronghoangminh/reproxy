// Package main provides the entry point for the Reproxy application.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/controllers"
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/logger"
	"github.com/letronghoangminh/reproxy/pkg/utils"
)

var (
	configPath = flag.String("config", "config/config.yaml", "path to the config file")
	// logFormat is currently unused but kept for future use
	_       = flag.String("log-format", "json", "log format (json or console)")
	version = flag.Bool("version", false, "print version information and exit")

	buildVersion = "dev"
	buildDate    = "unknown"

	appLogger interfaces.Logger
)

func printLogo() {
	fmt.Println(`
  _____  ______ _____  _____   ______   ____     __
 |  __ \|  ____|  __ \|  __ \ / __ \ \ / /\ \   / /
 | |__) | |__  | |__) | |__) | |  | \ V /  \ \_/ / 
 |  _  /|  __| |  ___/|  _  /| |  | |> <    \   /  
 | | \ \| |____| |    | | \ \| |__| / . \    | |   
 |_|  \_\______|_|    |_|  \_\\____/_/ \_\   |_|`)
	fmt.Printf("\nVersion: %s (Built on: %s)\n\n", buildVersion, buildDate)
}

func main() {
	printLogo()
	flag.Parse()

	if *version {
		fmt.Printf("Reproxy version %s (Built on: %s)\n", buildVersion, buildDate)
		os.Exit(0)
	}

	if err := loadConfig(); err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	cfg := config.GetConfig()

	appLogger = logger.NewLogger(*cfg)
	utils.Logger = appLogger

	defer func() {
		_ = appLogger.Sync()
	}()

	appLogger.Info("Starting Reproxy",
		"version", buildVersion,
		"build_date", buildDate,
		"config_path", *configPath)

	ctx, stop := setupSignalHandling()
	defer stop()

	wg := &sync.WaitGroup{}

	startTime := time.Now()

	go controllers.DefaultControllerServe(ctx, wg)
	go controllers.InitListenerControllers(ctx, wg)

	appLogger.Info("Reproxy started successfully",
		"startup_time_ms", time.Since(startTime).Milliseconds())

	<-ctx.Done()

	shutdownStart := time.Now()
	appLogger.Info("Shutdown initiated")

	wg.Wait()

	appLogger.Info("Shutdown complete",
		"shutdown_time_ms", time.Since(shutdownStart).Milliseconds())
	stop()
}

func loadConfig() error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic during configuration loading: %v\n", r)
			os.Exit(1)
		}
	}()

	config.LoadConfig(*configPath)
	return nil
}

func setupSignalHandling() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
}
