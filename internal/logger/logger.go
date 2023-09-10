package logger

import (
	"errors"
	"fmt"
	stdlog "log"
	"os"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate options-gen -out-filename=logger_options.gen.go -from-struct=Options
type Options struct {
	level          string `option:"mandatory" validate:"required,oneof=debug info warn error"`
	productionMode bool
}

var globalLogLevel = zap.NewAtomicLevel()

func MustInit(opts Options) {
	if err := Init(opts); err != nil {
		panic(err)
	}
}

func Init(opts Options) error {
	if err := setLogLevel(opts); err != nil {
		return fmt.Errorf("set log level error: %v", err)
	}

	encoderConfig := zapcore.EncoderConfig{
		LevelKey:    "level",
		MessageKey:  "msg",
		NameKey:     "component",
		TimeKey:     "T",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		EncodeLevel: zapcore.CapitalColorLevelEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if opts.productionMode {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(encoder, os.Stdout, globalLogLevel),
	}

	l := zap.New(zapcore.NewTee(cores...))
	zap.ReplaceGlobals(l)

	return nil
}

func setLogLevel(opts Options) error {
	if err := opts.Validate(); err != nil {
		return fmt.Errorf("validation logger options error: %v", err)
	}

	lvl, err := zapcore.ParseLevel(opts.level)
	if err != nil {
		return fmt.Errorf("parse logger level error: %v", err)
	}

	globalLogLevel.SetLevel(lvl)
	return nil
}

func SetLogLevel(opts Options) error {
	return setLogLevel(opts)
}

func LogLevel() string {
	return strings.ToUpper(globalLogLevel.String())
}

func Sync() {
	if err := zap.L().Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
		stdlog.Printf("cannot sync logger: %v", err)
	}
}
