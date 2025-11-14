package publisher

import (
	"context"
	"fmt"
	"iter"

	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
)

var log = logging.Logger("publisher")

type Publisher interface {
	// Publish creates, signs and publishes an advert. It then announces the new
	// advert to other indexers.
	Publish(ctx context.Context, provider peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) (ipld.Link, error)
}

type AsyncPublisher interface {
	// Publish creates, signs and publishes an advert but does so asynchronously, so no advert CID is returned.
	Publish(ctx context.Context, provider peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) error
}

type IPNIPublisher struct {
	batchPublisher *AdvertisementPublisher
	store          store.PublisherStore
}

// Publish creates a new advertisement from the latest head, signs it, and publishes it.
// Publish is not safe for concurrent use and advertisements may be lost if called concurrently. A mutex or any other
// synchronization mechanism must be used around Publish if it will be called from concurrent goroutines.
func (p *IPNIPublisher) Publish(ctx context.Context, providerInfo peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) (ipld.Link, error) {
	link, err := p.publishAdvForIndex(ctx, providerInfo.ID, providerInfo.Addrs, []byte(contextID), meta, false, digests)
	if err != nil {
		return nil, fmt.Errorf("publishing IPNI advert: %w", err)
	}
	return link, nil
}

var _ Publisher = (*IPNIPublisher)(nil)

// New creates a new IPNI publisher.
// IPNIPublisher is not safe for concurrent use. There is the risk of losing advertisements if Publish is called
// from concurrent goroutines. If you will be publishing from multiple goroutines concurrently, a synchronization
// mechanism (such as sync.Mutex) must be used to ensure that Publish is called serially.
func New(id crypto.PrivKey, store store.PublisherStore, opts ...Option) (*IPNIPublisher, error) {
	bp, err := NewAdvertisementPublisher(id, store, opts...)
	if err != nil {
		return nil, err
	}
	return &IPNIPublisher{
		batchPublisher: bp,
		store:          store,
	}, nil
}

func (p *IPNIPublisher) publishAdvForIndex(ctx context.Context, peer peer.ID, addrs []multiaddr.Multiaddr, contextID []byte, md metadata.Metadata, isRm bool, mhs iter.Seq[mh.Multihash]) (ipld.Link, error) {

	adv, err := GenerateAd(ctx, p.store, peer, addrs, contextID, md, isRm, mhs)
	if err != nil {
		return nil, err
	}

	p.batchPublisher.AddToBatch(adv)

	return p.batchPublisher.Commit(ctx)
}

type simpleAsyncPublisher struct {
	publisher Publisher
}

func AsyncFrom(p Publisher) AsyncPublisher {
	return &simpleAsyncPublisher{
		publisher: p,
	}
}

func (s *simpleAsyncPublisher) Publish(ctx context.Context, provider peer.AddrInfo, contextID string, digests iter.Seq[mh.Multihash], meta metadata.Metadata) error {
	_, err := s.publisher.Publish(ctx, provider, contextID, digests, meta)
	return err
}
