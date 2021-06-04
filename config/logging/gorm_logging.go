package logging

import (
	"github.com/apsdehal/go-logger"
)

// GormLogger struct
type GormLogger struct{
	Logger *logger.Logger
}

// Print - Log Formatter
func (n *GormLogger) Printf(s string, v ...interface{}) {
	if len(v) >= 4 {
		n.Logger.InfoF("[DAL] [%.2fms] %s", v[1], v[3])
	}
}