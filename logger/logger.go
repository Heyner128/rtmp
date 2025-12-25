package logger

import (
	"sync"

	"go.uber.org/zap"
)

var once sync.Once

var sugar *zap.SugaredLogger

func Get() *zap.SugaredLogger {
	once.Do(func() {
		logger, _ := zap.NewDevelopment()
		defer logger.Sync()
		sugar = logger.Sugar()
	})
	return sugar
}
