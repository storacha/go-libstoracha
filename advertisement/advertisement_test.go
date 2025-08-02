package advertisement_test

import (
	"net/url"
	"testing"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipni/go-libipni/maurl"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/storacha/go-libstoracha/advertisement"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/digestutil"
	"github.com/storacha/go-libstoracha/testutil"
)

func TestShardCID(t *testing.T) {
	curioPath := testutil.Must(multiaddr.NewMultiaddr("/http-path/" + url.PathEscape("piece/{blobCID}")))(t)
	blobPath := testutil.Must(multiaddr.NewMultiaddr("/http-path/" + url.PathEscape("blob/{blob}/{blob}")))(t)
	mixedPath := testutil.Must(multiaddr.NewMultiaddr("/http-path/" + url.PathEscape("blob/{blobCID}/{blob}")))(t)
	baseUrl := testutil.Must(url.Parse("https://node.com"))(t)
	base := testutil.Must(maurl.FromURL(baseUrl))(t)
	testCid := testutil.RandomCID(t)
	testMhs := testutil.RandomMultihashes(t, 3)
	testPieceLink := testutil.RandomPiece(t, 1<<16).Link()

	tests := []struct {
		name      string
		provider  peer.AddrInfo
		caveats   assert.LocationCaveats
		expected  *cid.Cid
		expectErr bool
	}{
		{
			name: "Valid shard CID from location url",
			provider: peer.AddrInfo{
				ID: testutil.RandomPeer(t),
				Addrs: []multiaddr.Multiaddr{
					multiaddr.Join(base, curioPath),
				},
			},
			caveats: assert.LocationCaveats{
				Content: types.FromHash(testMhs[0]),
				Location: []url.URL{
					*baseUrl.JoinPath("piece", testPieceLink.String()),
				},
			},
			expected: func() *cid.Cid {
				cid := testPieceLink.(cidlink.Link).Cid
				return &cid
			}(),
			expectErr: false,
		},
		{
			name: "Valid shard multihash from location url",
			provider: peer.AddrInfo{
				ID: testutil.RandomPeer(t),
				Addrs: []multiaddr.Multiaddr{
					multiaddr.Join(base, blobPath),
				},
			},
			caveats: assert.LocationCaveats{
				Content: types.FromHash(testMhs[0]),
				Location: []url.URL{
					*baseUrl.JoinPath("blob", digestutil.Format(testMhs[1]), digestutil.Format(testMhs[1])),
				},
			},
			expected: func() *cid.Cid {
				cid := cid.NewCidV1(cid.Raw, testMhs[1])
				return &cid
			}(),
			expectErr: false,
		},
		{
			name: "Valid CID in mixed location url",
			provider: peer.AddrInfo{
				ID: testutil.RandomPeer(t),
				Addrs: []multiaddr.Multiaddr{
					multiaddr.Join(base, mixedPath),
				},
			},
			caveats: assert.LocationCaveats{
				Content: types.FromHash(testMhs[0]),
				Location: []url.URL{
					*baseUrl.JoinPath("blob", testCid.String(), digestutil.Format(testCid.(cidlink.Link).Cid.Hash())),
				},
			},
			expected: func() *cid.Cid {
				cid := testCid.(cidlink.Link).Cid
				return &cid
			}(),
			expectErr: false,
		},
		// TODO: test for error on blob & cid not matching
		{
			name: "No matching location",
			provider: peer.AddrInfo{
				Addrs: []multiaddr.Multiaddr{
					multiaddr.Join(base, curioPath),
				},
			},
			caveats: assert.LocationCaveats{
				Location: []url.URL{
					*baseUrl.JoinPath("blob", digestutil.Format(testMhs[1]), digestutil.Format(testMhs[1])),
				},
				Content: types.FromHash(testMhs[0]),
			},
			expected:  nil,
			expectErr: false,
		},
		{
			name: "Invalid CID in location",
			provider: peer.AddrInfo{
				Addrs: []multiaddr.Multiaddr{
					multiaddr.Join(base, curioPath),
				},
			},
			caveats: assert.LocationCaveats{
				Location: []url.URL{
					*baseUrl.JoinPath("piece", "applesauce"),
				},
				Content: types.FromHash(testMhs[0]),
			},
			expected:  nil,
			expectErr: true,
		},
		// TODO: invalid multihash in location test
		// TODO: multiaddrs with blobs test

	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := advertisement.ShardCID(tt.provider, tt.caveats)
			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if result != nil && tt.expected != nil && !result.Equals(*tt.expected) {
				t.Errorf("expected: %v, got: %v", tt.expected, result)
			}
			if result == nil && tt.expected != nil {
				t.Errorf("expected: %v, got: nil", tt.expected)
			}
		})
	}
}
