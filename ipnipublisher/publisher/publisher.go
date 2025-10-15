package publisher

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"iter"

	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipni/go-libipni/announce"
	"github.com/ipni/go-libipni/announce/httpsender"
	"github.com/ipni/go-libipni/dagsync/ipnisync/head"
	"github.com/ipni/go-libipni/ingest/schema"
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
	*options
	sender announce.Sender
	key    crypto.PrivKey
	store  store.PublisherStore
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
	o := &options{
		topic: "/indexer/ingest/mainnet",
	}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	peer, err := peer.IDFromPrivateKey(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get peer ID from private key: %w", err)
	}
	pub := &IPNIPublisher{key: id, store: store, options: o}
	if len(o.announceURLs) > 0 {
		sender, err := httpsender.New(o.announceURLs, peer)
		if err != nil {
			return nil, fmt.Errorf("cannot create http announce sender: %w", err)
		}
		log.Info("HTTP announcements enabled")
		pub.sender = sender
	}
	return pub, nil
}

func (p *IPNIPublisher) publishAdvForIndex(ctx context.Context, peer peer.ID, addrs []multiaddr.Multiaddr, contextID []byte, md metadata.Metadata, isRm bool, mhs iter.Seq[mh.Multihash]) (ipld.Link, error) {
	var err error

	log := log.With("providerID", peer).With("contextID", base64.StdEncoding.EncodeToString(contextID))

	chunkLink, err := p.store.ChunkLinkForProviderAndContextID(ctx, peer, contextID)
	if err != nil {
		if !store.IsNotFound(err) {
			return nil, fmt.Errorf("could not get entries cid by provider + context id: %s", err)
		}
	}

	// If not removing, then generate the link for the list of CIDs from the
	// contextID using the multihash lister, and store the relationship.
	if !isRm {
		log.Info("Creating advertisement")

		// If no previously-published ad for this context ID.
		if chunkLink == nil {
			log.Info("Generating entries linked list for advertisement")

			// Generate the linked list ipld.Link that is added to the
			// advertisement and used for ingestion.
			chunkLink, err = p.store.PutEntries(ctx, mhs)
			if err != nil {
				return nil, fmt.Errorf("could not generate entries list: %s", err)
			}
			if chunkLink == nil {
				log.Warnw("chunking for context ID resulted in no link", "contextID", contextID)
				chunkLink = schema.NoEntries
			}

			// Store the relationship between providerID, contextID and CID of the
			// advertised list of Cids.
			err = p.store.PutChunkLinkForProviderAndContextID(ctx, peer, contextID, chunkLink)
			if err != nil {
				return nil, fmt.Errorf("failed to write provider + context id to entries cid mapping: %s", err)
			}
		} else {
			// Lookup metadata for this providerID and contextID.
			prevMetadata, err := p.store.MetadataForProviderAndContextID(ctx, peer, contextID)
			if err != nil {
				if !store.IsNotFound(err) {
					return nil, fmt.Errorf("could not get metadata for provider + context id: %s", err)
				}
				log.Warn("No metadata for existing provider + context ID, generating new advertisement")
			}

			if md.Equal(prevMetadata) {
				// Metadata is the same; no change, no need for new
				// advertisement.
				return nil, ErrAlreadyAdvertised
			}

			// Linked list is the same, but metadata is different, so generate
			// new advertisement with same linked list, but new metadata.
		}

		if err = p.store.PutMetadataForProviderAndContextID(ctx, peer, contextID, md); err != nil {
			return nil, fmt.Errorf("failed to write provider + context id to metadata mapping: %s", err)
		}
	} else {
		log.Info("Creating removal advertisement")

		if chunkLink == nil {
			return nil, ErrContextIDNotFound
		}

		// If removing by context ID, it means the list of CIDs is not needed
		// anymore, so we can remove the entry from the datastore.
		err = p.store.DeleteChunkLinkForProviderAndContextID(ctx, peer, contextID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete provider + context id to entries cid mapping: %s", err)
		}
		err = p.store.DeleteMetadataForProviderAndContextID(ctx, peer, contextID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete provider + context id to metadata mapping: %s", err)
		}

		// Create an advertisement to delete content by contextID by specifying
		// that advertisement has no entries.
		chunkLink = schema.NoEntries

		// The advertisement still requires a valid metadata even though
		// metadata is not used for removal. Create a valid empty metadata.
		md = metadata.Default.New()
	}

	mdBytes, err := md.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var stringAddrs []string
	for _, addr := range addrs {
		stringAddrs = append(stringAddrs, addr.String())
	}

	adv := schema.Advertisement{
		Provider:  peer.String(),
		Addresses: stringAddrs,
		Entries:   chunkLink,
		ContextID: contextID,
		Metadata:  mdBytes,
		IsRm:      isRm,
	}

	// Get the previous advertisement that was generated.
	prevHead, err := p.store.Head(ctx)
	if err != nil {
		if !store.IsNotFound(err) {
			return nil, fmt.Errorf("could not get latest advertisement: %s", err)
		}
	}

	// Check for cid.Undef for the previous link. If this is the case, then
	// this means there are no previous advertisements.
	if prevHead == nil {
		log.Info("Latest advertisement CID was undefined - no previous advertisement")
	} else {
		adv.PreviousID = prevHead.Head
	}

	// Sign the advertisement.
	if err = adv.Sign(p.key); err != nil {
		return nil, err
	}

	adLink, err := p.publish(ctx, adv, prevHead)
	if err != nil {
		if !isRm {
			// rollback chunk links and metadata additions
			if rclErr := p.store.DeleteChunkLinkForProviderAndContextID(ctx, peer, contextID); rclErr != nil {
				err = errors.Join(err, fmt.Errorf("rolling back provider + context ID to entries CID mapping: %w", rclErr))
			}
			if rmdErr := p.store.DeleteMetadataForProviderAndContextID(ctx, peer, contextID); rmdErr != nil {
				err = errors.Join(err, fmt.Errorf("rolling back provider + context ID to metadata mapping: %w", rmdErr))
			}
		}
		return nil, err
	}
	return adLink, nil
}

func (p *IPNIPublisher) publish(ctx context.Context, adv schema.Advertisement, prevHead *head.SignedHead) (ipld.Link, error) {
	lnk, err := p.publishLocal(ctx, adv, prevHead)
	if err != nil {
		log.Errorw("Failed to store advertisement locally", "err", err)
		return nil, fmt.Errorf("failed to publish advertisement locally: %w", err)
	}
	if p.sender != nil {
		err = announce.Send(ctx, lnk.(cidlink.Link).Cid, p.pubHTTPAnnounceAddrs, p.sender)
		if err != nil {
			log.Warnw("Failed to announce advertisement", "err", err)
		}
	}
	return lnk, nil
}

func (p *IPNIPublisher) publishLocal(ctx context.Context, adv schema.Advertisement, prevHead *head.SignedHead) (ipld.Link, error) {
	if err := adv.Validate(); err != nil {
		return nil, err
	}

	lnk, err := p.store.PutAdvert(ctx, adv)
	if err != nil {
		return nil, err
	}
	log.Info("Stored ad in local link system")

	head, err := head.NewSignedHead(lnk.(cidlink.Link).Cid, p.topic, p.key)
	if err != nil {
		log.Errorw("Failed to generate signed head for the latest advertisement", "err", err)
		return nil, fmt.Errorf("failed to generate signed head for the latest advertisement: %w", err)
	}
	if _, err := p.store.ReplaceHead(ctx, prevHead, head); err != nil {
		log.Errorw("Failed to update reference to the latest advertisement", "err", err)
		return nil, fmt.Errorf("failed to update reference to latest advertisement: %w", err)
	}
	log.Info("Updated reference to the latest advertisement successfully")
	return lnk, nil
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
