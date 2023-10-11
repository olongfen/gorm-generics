package achieve

import "gorm.io/gorm/logger"

type options struct {
	opentracingPlugin *OpentracingPlugin
	autoMigrate       bool
	autoMigrateDst    []any
	logger            logger.Interface
	tablePrefix       string
}

type Option interface {
	apply(*options)
}

type loggerOption struct {
	log logger.Interface
}

func (l loggerOption) apply(opts *options) {
	opts.logger = l.log
}

func WithLogger(log logger.Interface) Option {
	return loggerOption{log: log}
}

type autoMigrateOption bool

func (a autoMigrateOption) apply(opts *options) {
	opts.autoMigrate = bool(a)
}

func WithAutoMigrate(a bool) Option {
	return autoMigrateOption(a)
}

type opentracingPluginOption struct {
	opentracingPlugin *OpentracingPlugin
}

func (o opentracingPluginOption) apply(opts *options) {
	opts.opentracingPlugin = o.opentracingPlugin
}

func WithOpentracingPlugin(op *OpentracingPlugin) Option {
	return &opentracingPluginOption{opentracingPlugin: op}
}

type autoMigrateDstOption []any

func (a autoMigrateDstOption) apply(opts *options) {
	opts.autoMigrateDst = a
}

func WithAutoMigrateDst(models []any) Option {
	return autoMigrateDstOption(models)
}

type tablePrefixOption string

func (a tablePrefixOption) apply(opts *options) {
	opts.tablePrefix = string(a)
}

func WithTablePrefix(s string) Option {
	return tablePrefixOption(s)
}
