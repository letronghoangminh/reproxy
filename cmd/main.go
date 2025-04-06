package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/letronghoangminh/reproxy/pkg/config"
	"github.com/letronghoangminh/reproxy/pkg/controllers"
	"github.com/letronghoangminh/reproxy/pkg/logger"
)

var (
	configPath = flag.String("config", "config/config.yaml", "path the config file")
)

func printLogo() {
	fmt.Println(`
  _____  ______ _____  _____   ______   ____     __
 |  __ \|  ____|  __ \|  __ \ / __ \ \ / /\ \   / /
 | |__) | |__  | |__) | |__) | |  | \ V /  \ \_/ / 
 |  _  /|  __| |  ___/|  _  /| |  | |> <    \   /  
 | | \ \| |____| |    | | \ \| |__| / . \    | |   
 |_|  \_\______|_|    |_|  \_\\____/_/ \_\   |_|`)
}

func main() {
	printLogo()
	flag.Parse()

	config.LoadConfig(*configPath)
	cfg := config.GetConfig()

	logger := logger.NewLogger(*cfg)

	defer logger.Sync()

	logger.Info("config loaded successfully for reproxy")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM | syscall.SIGINT | syscall.SIGQUIT,
	)
	defer stop()

	wg := &sync.WaitGroup{}

	go controllers.DefaultControllerServe(ctx, wg)
	go controllers.InitListenerControllers(ctx, wg)

	select {
	case <-ctx.Done():
		wg.Wait()
		logger.Info("all listeners have been shut down")
		stop()
	}
}
