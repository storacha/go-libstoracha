package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipni/go-libipni/find/model"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/storacha/go-libstoracha/ipnipublisher/notifier"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func mockIpniApi(t *testing.T, id peer.ID) (*httptest.Server, []ipld.Link) {
	var ads []ipld.Link
	for range 10 {
		ads = append(ads, testutil.RandomCID(t))
	}

	n := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if n >= len(ads) {
			w.WriteHeader(500)
			return
		}
		resp := model.ProviderInfo{
			AddrInfo:          peer.AddrInfo{ID: id},
			LastAdvertisement: ads[n].(cidlink.Link).Cid,
		}
		bytes, _ := json.Marshal(resp)
		w.Write(bytes)
		n++
	}))

	return ts, ads
}

func TestNotifier(t *testing.T) {
	notifier.NotifierPollInterval = time.Millisecond

	priv, _, err := crypto.GenerateEd25519Key(nil)
	require.NoError(t, err)

	pid, err := peer.IDFromPrivateKey(priv)
	require.NoError(t, err)

	t.Run("notifies all CIDs", func(t *testing.T) {
		ts, ads := mockIpniApi(t, pid)
		defer ts.Close()

		notif, err := notifier.NewRemoteSyncNotifier(ts.URL, priv, &mockHead{})
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(len(ads))

		var notifications []ipld.Link
		notif.Notify(func(ctx context.Context, head, prev ipld.Link) {
			notifications = append(notifications, head)
			wg.Done()
		})

		notif.Start(context.Background())
		wg.Wait()
		notif.Stop()

		require.Equal(t, ads, notifications)
	})

	t.Run("notifies all CIDs with known head", func(t *testing.T) {
		ts, chain := mockIpniApi(t, pid)
		defer ts.Close()

		notif, err := notifier.NewRemoteSyncNotifier(ts.URL, priv, &mockHead{chain[0]})
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(len(chain) - 1)

		var notifications []ipld.Link
		notif.Notify(func(ctx context.Context, head, prev ipld.Link) {
			notifications = append(notifications, head)
			wg.Done()
		})

		notif.Start(context.Background())
		wg.Wait()
		notif.Stop()

		require.Equal(t, chain[1:], notifications)
	})
}

type mockHead struct {
	head ipld.Link
}

func (m *mockHead) Get(context.Context) ipld.Link {
	return m.head
}

func (m *mockHead) Set(_ context.Context, head ipld.Link) error {
	m.head = head
	return nil
}
