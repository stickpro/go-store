package repository

import "github.com/jackc/pgx/v5"

type Option func(*Options)

type Options struct {
	Tx pgx.Tx
}

func parseOptions(opts ...Option) Options {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return options
}

func WithTx(tx pgx.Tx) Option {
	return func(o *Options) {
		o.Tx = tx
	}
}
