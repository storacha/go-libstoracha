package store

import "github.com/ipni/go-libipni/metadata"

// Option is an option configuring a store.
type Option func(cfg *options)

type options struct {
	metadataContext metadata.MetadataContext
}

// WithMetadataContext configues the IPNI metadata context, allowing custom
// metadata types to be stored. If not configured, the default context is used.
func WithMetadataContext(context metadata.MetadataContext) Option {
	return func(o *options) {
		o.metadataContext = context
	}
}
