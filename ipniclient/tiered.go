package ipniclient

import (
	"context"
	"errors"
	"time"

	ipnifind "github.com/ipni/go-libipni/find/client"
	"github.com/ipni/go-libipni/find/model"
	multihash "github.com/multiformats/go-multihash"
)

type TieredFinder struct {
	Tiers   []ipnifind.Finder
	timeout time.Duration
}

type TieredFinderOption func(*TieredFinder) error

// WithTierFindTimeout sets the timeout for each individual tier find operation.
func WithTierFindTimeout(timeout time.Duration) TieredFinderOption {
	return func(f *TieredFinder) error {
		f.timeout = timeout
		return nil
	}
}

// NewTieredFinder creates a new [ipnifind.Finder] with the provided tiers and
// options.
//
// The finders are tried in order until a successful find is made or all finders
// have been tried. If all finders fail, an error is returned that combines all
// errors from each finder.
//
// A successful request with no results is considered a fail and will cause the
// next tier to be tried.
func NewTieredFinder(tiers []ipnifind.Finder, opts ...TieredFinderOption) (*TieredFinder, error) {
	finder := TieredFinder{Tiers: tiers, timeout: 5 * time.Second}
	for _, opt := range opts {
		if err := opt(&finder); err != nil {
			return nil, err
		}
	}
	return &finder, nil
}

func (t *TieredFinder) Find(ctx context.Context, digest multihash.Multihash) (*model.FindResponse, error) {
	var findErr error
	for _, tier := range t.Tiers {
		ctx, cancel := context.WithTimeout(ctx, t.timeout)
		resp, err := tier.Find(ctx, digest)
		cancel()
		if err != nil {
			findErr = errors.Join(findErr, err)
			continue
		}
		if len(resp.MultihashResults) == 0 {
			continue // No results found, try next tier.
		}
		return resp, nil
	}
	// if no error, all tiers were tried but no results found, return empty
	// response without error
	if findErr == nil {
		return &model.FindResponse{}, nil
	}
	return nil, findErr
}

var _ ipnifind.Finder = (*TieredFinder)(nil)
