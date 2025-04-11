package utils

import (
	"github.com/letronghoangminh/reproxy/pkg/interfaces"
	"github.com/letronghoangminh/reproxy/pkg/logger"
)

var Logger interfaces.Logger

func GetLogger() interfaces.Logger {
	if Logger == nil {
		Logger = logger.GetLogger()
	}
	return Logger
}
