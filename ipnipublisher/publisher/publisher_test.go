package publisher_test

import (
	"bytes"
	"context"
	"errors"
	"math/rand/v2"
	"slices"
	"sort"
	"testing"

	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/storacha/go-libstoracha/ipnipublisher/publisher"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-libstoracha/testutil"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	priv, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)

	pid, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	provInfo := peer.AddrInfo{ID: pid}

	ctx := context.Background()

	t.Run("single advert", func(t *testing.T) {
		dstore := dssync.MutexWrap(datastore.NewMapDatastore())
		st := store.FromDatastore(dstore)
		p, err := publisher.New(priv, st)
		require.NoError(t, err)

		digests := testutil.RandomMultihashes(t, rand.IntN(10)+1)
		adlnk, err := p.Publish(ctx, provInfo, testutil.RandomCID(t).String(), slices.Values(digests), metadata.Default.New())
		require.NoError(t, err)

		ad, err := st.Advert(ctx, adlnk)
		require.NoError(t, err)

		var ents []multihash.Multihash
		for e, err := range st.Entries(ctx, ad.Entries) {
			require.NoError(t, err)
			ents = append(ents, e)
		}

		require.Equal(t, digests, ents)
	})

	t.Run("single advert, chunked entries", func(t *testing.T) {
		dstore := dssync.MutexWrap(datastore.NewMapDatastore())
		st := store.FromDatastore(dstore)
		p, err := publisher.New(priv, st)
		require.NoError(t, err)

		digests := testutil.RandomMultihashes(t, store.MaxEntryChunkSize+1)
		adlnk, err := p.Publish(ctx, provInfo, testutil.RandomCID(t).String(), slices.Values(digests), metadata.Default.New())
		require.NoError(t, err)

		ad, err := st.Advert(ctx, adlnk)
		require.NoError(t, err)

		var estrs []string
		for e, err := range st.Entries(ctx, ad.Entries) {
			require.NoError(t, err)
			estrs = append(estrs, e.B58String())
		}
		sort.Strings(estrs)

		var dstrs []string
		for _, d := range digests {
			dstrs = append(dstrs, d.B58String())
		}
		sort.Strings(dstrs)

		require.Equal(t, len(digests), len(estrs))
		require.Equal(t, dstrs, estrs)
	})

	t.Run("multiple adverts", func(t *testing.T) {
		dstore := dssync.MutexWrap(datastore.NewMapDatastore())
		st := store.FromDatastore(dstore)
		p, err := publisher.New(priv, st)
		require.NoError(t, err)

		var adLinks []ipld.Link
		var contextIDs []string
		var digestLists [][]multihash.Multihash
		for range 1 + rand.IntN(100) {
			ctxid := testutil.RandomCID(t).String()
			digests := testutil.RandomMultihashes(t, 1+rand.IntN(100))

			l, err := p.Publish(ctx, provInfo, ctxid, slices.Values(digests), metadata.Default.New())
			require.NoError(t, err)

			adLinks = append(adLinks, l)
			contextIDs = append(contextIDs, ctxid)
			digestLists = append(digestLists, digests)
		}

		for i, adLink := range adLinks {
			ad, err := st.Advert(ctx, adLink)
			require.NoError(t, err)

			var digests []multihash.Multihash
			for e, err := range st.Entries(ctx, ad.Entries) {
				require.NoError(t, err)
				digests = append(digests, e)
			}

			require.Equal(t, contextIDs[i], string(ad.ContextID))
			require.Equal(t, digestLists[i], digests)
		}
	})

	t.Run("concurrent publish returns error", func(t *testing.T) {
		ms := mockStore{data: map[string][]byte{}}
		st := store.NewPublisherStore(
			&ms,
			store.NewDatastoreProviderContextTable(datastore.NewMapDatastore()),
			store.NewDatastoreProviderContextTable(datastore.NewMapDatastore()),
		)

		p, err := publisher.New(priv, st)
		require.NoError(t, err)

		ms.beforeReplace = func() {
			ms.beforeReplace = nil
			ctxid := testutil.RandomCID(t).String()
			digests := testutil.RandomMultihashes(t, 1+rand.IntN(100))
			l, err := p.Publish(ctx, provInfo, ctxid, slices.Values(digests), metadata.Default.New(&metadata.IpfsGatewayHttp{}))
			require.NoError(t, err)
			t.Logf("published new advert before another: %s", l)
		}

		ctxid := testutil.RandomCID(t).String()
		digests := testutil.RandomMultihashes(t, 1+rand.IntN(100))
		_, err = p.Publish(ctx, provInfo, ctxid, slices.Values(digests), metadata.Default.New(&metadata.IpfsGatewayHttp{}))
		require.ErrorIs(t, err, store.ErrPreconditionFailed)

		// subsequent publish should succeed
		l, err := p.Publish(ctx, provInfo, ctxid, slices.Values(digests), metadata.Default.New(&metadata.IpfsGatewayHttp{}))
		require.NoError(t, err)
		t.Logf("published new advert after retry: %s", l)
	})
}

type mockStore struct {
	data          map[string][]byte
	beforeReplace func()
}

func (ms *mockStore) Get(ctx context.Context, key string) ([]byte, error) {
	d, ok := ms.data[key]
	if !ok {
		return nil, store.NewErrNotFound(errors.New("key not found in map"))
	}
	return d, nil
}

func (ms *mockStore) Put(ctx context.Context, key string, data []byte) error {
	ms.data[key] = data
	return nil
}

func (ms *mockStore) Replace(ctx context.Context, key string, old []byte, new []byte) error {
	if ms.beforeReplace != nil {
		ms.beforeReplace()
	}
	d, ok := ms.data[key]
	if !ok {
		if len(old) > 0 {
			return store.ErrPreconditionFailed
		}
	}
	if !bytes.Equal(d, old) {
		return store.ErrPreconditionFailed
	}
	ms.data[key] = new
	return nil
}
