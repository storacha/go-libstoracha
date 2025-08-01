package testutil

import (
	crand "crypto/rand"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	commp "github.com/filecoin-project/go-fil-commp-hashhash"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	peer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/piece/digest"
	"github.com/storacha/go-libstoracha/piece/piece"
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
