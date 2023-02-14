package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// root logger
var log *zap.SugaredLogger

// Init the root logger.
func Init() {
	zapCfg := zap.NewDevelopmentConfig()
	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}
	withOptions := zapLogger.WithOptions(zap.AddCallerSkip(1)) // skip wrapper func
	log = withOptions.Sugar()
}

func getDefaultLog() *zap.SugaredLogger {
	if log != nil {
		return log
	}
	Init()
	return log
}

func Sync() error {
	return getDefaultLog().Sync()
}

// Debug calls log.Debug on the root Logger.
func Debug(args ...interface{}) {
	getDefaultLog().Debug(args...)
}

// Info calls log.Info on the root Logger.
func Info(args ...interface{}) {
	getDefaultLog().Info(args...)
}

// Warn calls log.Warn on the root Logger.
func Warn(args ...interface{}) {
	getDefaultLog().Warn(args...)
}

// Error calls log.Error on the root Logger.
func Error(args ...interface{}) {
	getDefaultLog().Error(args...)
}

// Fatal calls log.Fatal on the root Logger.
func Fatal(args ...interface{}) {
	getDefaultLog().Fatal(args...)
}

// Debugf calls log.Debugf on the root Logger.
func Debugf(template string, args ...interface{}) {
	getDefaultLog().Debugf(template, args...)
}

// Infof calls log.Infof on the root Logger.
func Infof(template string, args ...interface{}) {
	getDefaultLog().Infof(template, args...)
}

// Warnf calls log.Warnf on the root Logger.
func Warnf(template string, args ...interface{}) {
	getDefaultLog().Warnf(template, args...)
}

// Fatalf calls log.Fatalf on the root Logger.
func Fatalf(template string, args ...interface{}) {
	getDefaultLog().Fatalf(template, args...)
}

// Errorf calls log.Errorf on the root logger and stores the error message into
// the ErrorFile.
func Errorf(template string, args ...interface{}) {
	getDefaultLog().Errorf(template, args...)
}
