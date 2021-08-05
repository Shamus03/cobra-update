package update

import (
	"log"

	"github.com/google/go-github/v37/github"
)

type optionCtx struct {
	logger                *log.Logger
	debugLogger           *log.Logger
	errorLogger           *log.Logger
	assetIsCompatibleFunc func(*github.ReleaseAsset) bool
}

type Option interface {
	apply(ctx *optionCtx)
}

func SetAssetIsCompatibleFunc(f func(*github.ReleaseAsset) bool) Option {
	return setAssetIsCompatibleFunc{f}
}

type setAssetIsCompatibleFunc struct {
	f func(*github.ReleaseAsset) bool
}

func (o setAssetIsCompatibleFunc) apply(ctx *optionCtx) {
	ctx.assetIsCompatibleFunc = o.f
}

func SetLoggerFlags(flags int) Option {
	return setLoggerFlags{flags}
}

type setLoggerFlags struct {
	flags int
}

func (o setLoggerFlags) apply(ctx *optionCtx) {
	ctx.logger.SetFlags(o.flags)
}

func SetDebugLoggerFlags(flags int) Option {
	return setDebugLoggerFlags{flags}
}

type setDebugLoggerFlags struct {
	flags int
}

func (o setDebugLoggerFlags) apply(ctx *optionCtx) {
	ctx.debugLogger.SetFlags(o.flags)
}

func SetErrorLoggerFlags(flags int) Option {
	return setErrorLoggerFlags{flags}
}

type setErrorLoggerFlags struct {
	flags int
}

func (o setErrorLoggerFlags) apply(ctx *optionCtx) {
	ctx.errorLogger.SetFlags(o.flags)
}
