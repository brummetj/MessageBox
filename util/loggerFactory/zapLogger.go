package loggerFactory

import (
	"MessageBox/config"
	"MessageBox/util/logger"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)
var logLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "ALERT",
	zapcore.FatalLevel:  "EMERGENCY",
}
func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("Jan 01, 2006  15:04:05"))
}
func CustomLevelFileEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + logLevelSeverity[level] + "]")
}

func RegisterLogger(c config.LogConfig) error {
	var levelMap = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info": zapcore.InfoLevel,
		"warning": zapcore.WarnLevel,
	}
	var cfg = zap.Config{
		Level:             zap.NewAtomicLevelAt(levelMap[c.Level]),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:          c.Encoding,
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       c.OutputPaths,
		ErrorOutputPaths:  c.ErrorPaths,
		InitialFields:     nil,
	}
	cfg.EncoderConfig.EncodeTime = SyslogTimeEncoder
	cfg.EncoderConfig.EncodeLevel = CustomLevelFileEncoder
	zLogger, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return errors.Wrap(err, "unable to build zap logger")
	}
	defer zLogger.Sync()
	zSugar := zLogger.Sugar()
	logger.SetLogger(zSugar)
	logger.Log.Info("Logger construction completed")
	return nil
}

func ZapEchoHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			req := c.Request()
			res := c.Response()
			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			var fieldStr = "{STATUS: %v} - {LATENCY: %v} - {METHOD: %v} - {PATH: %v} - {REMOTE_IP: %v}"
			var fields = fmt.Sprintf(fieldStr, res.Status, time.Since(start).String(), req.Method, req.RequestURI, c.RealIP())
			n := res.Status
			switch {
				case n >= 500:
					logger.Log.Errorf("Server error: %v - {Error: %v}", fields, err)
				case n >= 400:
					logger.Log.Warnf("Client error: %v- {Error: %v}", fields, err)
				case n >= 300:
					logger.Log.Infof("Redirection: %v", fields)
				default:
					logger.Log.Infof("Success: %v", fields)
			}

			return nil
		}
	}
}