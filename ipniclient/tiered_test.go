package ipniclient_test

import (
	"context"
	"errors"
	"testing"
	"time"

	ipnifind "github.com/ipni/go-libipni/find/client"
	"github.com/ipni/go-libipni/find/model"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/ipniclient"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

type mockFinder struct {
	resp      *model.FindResponse
	err       error
	callCount int
}

func (m *mockFinder) Find(_ context.Context, _ multihash.Multihash) (*model.FindResponse, error) {
	m.callCount++
	return m.resp, m.err
}

var _ ipnifind.Finder = (*mockFinder)(nil)

func makeResponse(digest multihash.Multihash) *model.FindResponse {
	return &model.FindResponse{
		MultihashResults: []model.MultihashResult{
			{Multihash: digest},
		},
	}
}

func TestNewTieredFinder(t *testing.T) {
	t.Run("no options uses default", func(t *testing.T) {
		finder, err := ipniclient.NewTieredFinder(nil)
		require.NoError(t, err)
		require.NotNil(t, finder)
	})

	t.Run("WithTierFindTimeout sets timeout", func(t *testing.T) {
		finder, err := ipniclient.NewTieredFinder(nil, ipniclient.WithTierFindTimeout(42*time.Millisecond))
		require.NoError(t, err)
		require.NotNil(t, finder)
	})
}

func TestTieredFinder_Find(t *testing.T) {
	digest := testutil.RandomMultihash(t)

	t.Run("returns result from first tier", func(t *testing.T) {
		resp := makeResponse(digest)
		tier1 := &mockFinder{resp: resp}
		tier2 := &mockFinder{resp: makeResponse(digest)}

		finder, err := ipniclient.NewTieredFinder([]ipnifind.Finder{tier1, tier2})
		require.NoError(t, err)

		got, err := finder.Find(t.Context(), digest)
		require.NoError(t, err)
		require.Equal(t, resp, got)
		require.Equal(t, 1, tier1.callCount)
		require.Equal(t, 0, tier2.callCount, "second tier should not be queried")
	})

	t.Run("skips tier with no results and returns second tier", func(t *testing.T) {
		tier1 := &mockFinder{resp: &model.FindResponse{}}
		resp := makeResponse(digest)
		tier2 := &mockFinder{resp: resp}

		finder, err := ipniclient.NewTieredFinder([]ipnifind.Finder{tier1, tier2})
		require.NoError(t, err)

		got, err := finder.Find(t.Context(), digest)
		require.NoError(t, err)
		require.Equal(t, resp, got)
		require.Equal(t, 1, tier1.callCount)
		require.Equal(t, 1, tier2.callCount)
	})

	t.Run("skips erroring tier and returns result from next", func(t *testing.T) {
		tier1 := &mockFinder{err: errors.New("tier1 error")}
		resp := makeResponse(digest)
		tier2 := &mockFinder{resp: resp}

		finder, err := ipniclient.NewTieredFinder([]ipnifind.Finder{tier1, tier2})
		require.NoError(t, err)

		got, err := finder.Find(t.Context(), digest)
		require.NoError(t, err)
		require.Equal(t, resp, got)
		require.Equal(t, 1, tier1.callCount)
		require.Equal(t, 1, tier2.callCount)
	})

	t.Run("returns combined errors when all tiers fail", func(t *testing.T) {
		err1 := errors.New("tier1 error")
		err2 := errors.New("tier2 error")
		tier1 := &mockFinder{err: err1}
		tier2 := &mockFinder{err: err2}

		finder, err := ipniclient.NewTieredFinder([]ipnifind.Finder{tier1, tier2})
		require.NoError(t, err)

		_, err = finder.Find(t.Context(), digest)
		require.Error(t, err)
		require.ErrorIs(t, err, err1)
		require.ErrorIs(t, err, err2)
	})

	t.Run("returns empty response when all tiers have no results", func(t *testing.T) {
		tier1 := &mockFinder{resp: &model.FindResponse{}}
		tier2 := &mockFinder{resp: &model.FindResponse{}}

		finder, err := ipniclient.NewTieredFinder([]ipnifind.Finder{tier1, tier2})
		require.NoError(t, err)

		got, err := finder.Find(t.Context(), digest)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got.MultihashResults)
	})

	t.Run("returns empty response with no tiers", func(t *testing.T) {
		finder, err := ipniclient.NewTieredFinder(nil)
		require.NoError(t, err)

		got, err := finder.Find(t.Context(), digest)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got.MultihashResults)
	})

	t.Run("mixed errors and empty results returns error", func(t *testing.T) {
		tier1 := &mockFinder{err: errors.New("tier1 error")}
		tier2 := &mockFinder{resp: &model.FindResponse{}}

		finder, err := ipniclient.NewTieredFinder([]ipnifind.Finder{tier1, tier2})
		require.NoError(t, err)

		// tier1 errors, tier2 returns no results — findErr is non-nil so error is returned
		_, err = finder.Find(t.Context(), digest)
		require.Error(t, err)
	})
}
