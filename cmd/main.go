package main

import (
	"flag"
	"fmt"

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
	
	go controllers.DefaultControllerServe()
	go controllers.InitListenerControllers()

	select {}
}
