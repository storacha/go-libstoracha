package testutil

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipni/go-libipni/find/model"
	ipnimeta "github.com/ipni/go-libipni/metadata"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/blobindex"
	cassert "github.com/storacha/go-libstoracha/capabilities/assert"
	ctypes "github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/metadata"
	"github.com/storacha/go-libstoracha/piece/digest"
	"github.com/storacha/go-libstoracha/piece/piece"
	"github.com/storacha/go-ucanto/core/car"
	"github.com/storacha/go-ucanto/core/delegation"
	"github.com/storacha/go-ucanto/core/ipld/block"
	"github.com/storacha/go-ucanto/principal"
	"github.com/storacha/go-ucanto/principal/ed25519/signer"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/stretchr/testify/require"
)

func RandomBytes(t *testing.T, size int) []byte {
	bytes := make([]byte, size)
	_, _ = crand.Read(bytes)
	return bytes
}

var seedSeq int64

func RandomPeer(t *testing.T) peer.ID {
	src := rand.NewSource(seedSeq)
	seedSeq++
	r := rand.New(src)
	_, publicKey := Must2(crypto.GenerateEd25519Key(r))(t)
	return Must(peer.IDFromPublicKey(publicKey))(t)
}

func RandomPrincipal(t *testing.T) ucan.Principal {
	return RandomSigner(t)
}

func RandomSigner(t *testing.T) principal.Signer {
	return Must(signer.Generate())(t)
}

func RandomMultiaddr(t *testing.T) multiaddr.Multiaddr {
	// generate a random ipv4 address
	addr := &net.TCPAddr{IP: net.IPv4(byte(rand.Intn(255)), byte(rand.Intn(255)), byte(rand.Intn(255)), byte(rand.Intn(255))), Port: rand.Intn(65535)}
	maddr := Must(manet.FromIP(addr.IP))(t)
	port := Must(multiaddr.NewComponent(multiaddr.ProtocolWithCode(multiaddr.P_TCP).Name, strconv.Itoa(addr.Port)))(t)
	return multiaddr.Join(maddr, port)
}

func RandomCID(t *testing.T) datamodel.Link {
	return cidlink.Link{Cid: cid.NewCidV1(cid.Raw, RandomMultihash(t))}
}

func RandomMultihash(t *testing.T) mh.Multihash {
	bytes := RandomBytes(t, 10)
	return Must(mh.Sum(bytes, mh.SHA2_256, -1))(t)
}

func MultihashFromBytes(t *testing.T, b []byte) mh.Multihash {
	return Must(mh.Sum(b, mh.SHA2_256, -1))(t)
}

func RandomMultihashes(t *testing.T, count int) []mh.Multihash {
	require.Greater(t, count, 0, "count must be greater than 0")
	mhs := make([]mh.Multihash, 0, count)
	for range count {
		mhs = append(mhs, RandomMultihash(t))
	}
	return mhs
}

// RandomPiece is a helper that produces a piece with the given unpadded size.
func RandomPiece(t testing.TB, unpaddedSize int64) piece.PieceLink {
	t.Helper()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	dataReader := io.LimitReader(r, unpaddedSize)

	calc := &commp.Calc{}
	n, err := io.Copy(calc, dataReader)
	require.NoError(t, err, "failed copying data into commp.Calc")
	require.Equal(t, unpaddedSize, n)

	commP, paddedSize, err := calc.Digest()
	require.NoError(t, err, "failed to compute commP")

	pieceDigest, err := digest.FromCommitmentAndSize(commP, uint64(unpaddedSize))
	require.NoError(t, err, "failed building piece digest from commP")

	p := piece.FromPieceDigest(pieceDigest)
	// Ensure our piece’s PaddedSize matches the commp library’s reported paddedSize.
	require.Equal(t, paddedSize, p.PaddedSize())

	t.Logf("Created test piece: %s from unpadded size: %d", pieceLinkString(p), unpaddedSize)
	return p
}

// pieceLinkString is a helper to display piece metadata in logs.
func pieceLinkString(p piece.PieceLink) string {
	return fmt.Sprintf("Piece: %s, Height: %d, Padding: %d, PaddedSize: %d",
		p.Link(), p.Height(), p.Padding(), p.PaddedSize())
}

// RandomCAR creates a CAR with a single block of random bytes of the specified
// size. It returns the link of the root block, the hash of the CAR itself and
// the bytes of the CAR.
func RandomCAR(t *testing.T, size int) (datamodel.Link, mh.Multihash, []byte) {
	bytes := RandomBytes(t, size)
	c := Must(cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.SHA2_256,
		MhLength: -1,
	}.Sum(bytes))(t)

	root := cidlink.Link{Cid: c}
	r := car.Encode([]datamodel.Link{root}, func(yield func(block.Block, error) bool) {
		yield(block.NewBlock(root, bytes), nil)
	})
	carBytes := Must(io.ReadAll(r))(t)
	carDigest := Must(mh.Sum(carBytes, mh.SHA2_256, -1))(t)
	return root, carDigest, carBytes
}

func delegateCapability[Caveats ucan.CaveatBuilder](t *testing.T, can ucan.Capability[Caveats]) delegation.Delegation {
	did := Must(signer.Generate())(t)
	return Must(delegation.Delegate(Service, did, []ucan.Capability[Caveats]{can}))(t)
}

func RandomLocationClaim(t *testing.T) ucan.Capability[cassert.LocationCaveats] {
	return cassert.Location.New(Service.DID().String(), cassert.LocationCaveats{
		Content:  ctypes.FromHash(RandomMultihash(t)),
		Location: []url.URL{*TestURL},
	})
}

func RandomLocationDelegation(t *testing.T) delegation.Delegation {
	return delegateCapability(t, RandomLocationClaim(t))
}

func RandomIndexClaim(t *testing.T) ucan.Capability[cassert.IndexCaveats] {
	return cassert.Index.New(Service.DID().String(), cassert.IndexCaveats{
		Content: RandomCID(t),
		Index:   RandomCID(t),
	})
}

func RandomIndexDelegation(t *testing.T) delegation.Delegation {
	return delegateCapability(t, RandomIndexClaim(t))
}

func RandomEqualsClaim(t *testing.T) ucan.Capability[cassert.EqualsCaveats] {
	return cassert.Equals.New(Service.DID().String(), cassert.EqualsCaveats{
		Content: ctypes.FromHash(RandomMultihash(t)),
		Equals:  RandomCID(t),
	})
}

func RandomEqualsDelegation(t *testing.T) delegation.Delegation {
	return delegateCapability(t, RandomEqualsClaim(t))
}

func RandomProviderResult(t *testing.T) model.ProviderResult {
	return model.ProviderResult{
		ContextID: RandomBytes(t, 10),
		Metadata:  RandomBytes(t, 10),
		Provider: &peer.AddrInfo{
			ID: RandomPeer(t),
			Addrs: []multiaddr.Multiaddr{
				RandomMultiaddr(t),
				RandomMultiaddr(t),
			},
		},
	}
}

func RandomBitswapProviderResult(t *testing.T) model.ProviderResult {
	pr := RandomProviderResult(t)
	bitswapMeta := Must(ipnimeta.Bitswap{}.MarshalBinary())(t)
	pr.Metadata = bitswapMeta
	return pr
}

func RandomIndexClaimProviderResult(t *testing.T) model.ProviderResult {
	indexMeta := metadata.IndexClaimMetadata{
		Index:      RandomCID(t).(cidlink.Link).Cid,
		Expiration: 0,
		Claim:      RandomCID(t).(cidlink.Link).Cid,
	}
	metaBytes := Must(indexMeta.MarshalBinary())(t)

	pr := RandomProviderResult(t)
	pr.Metadata = metaBytes
	return pr
}

func RandomLocationCommitmentProviderResult(t *testing.T) model.ProviderResult {
	shard := RandomCID(t).(cidlink.Link).Cid
	locationMeta := metadata.LocationCommitmentMetadata{
		Shard:      &shard,
		Range:      &metadata.Range{Offset: 128},
		Expiration: 0,
		Claim:      RandomCID(t).(cidlink.Link).Cid,
	}
	metaBytes := Must(locationMeta.MarshalBinary())(t)

	pr := RandomProviderResult(t)
	pr.Metadata = metaBytes
	return pr
}

func RandomShardedDagIndexView(t *testing.T, size int) (mh.Multihash, blobindex.ShardedDagIndexView) {
	root, digest, bytes := RandomCAR(t, size)
	shard := Must(blobindex.FromShardArchives(root, [][]byte{bytes}))(t)
	return digest, shard
}
