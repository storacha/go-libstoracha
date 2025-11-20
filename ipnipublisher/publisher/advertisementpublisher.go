package publisher

import (
	"context"
	"fmt"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipni/go-libipni/announce"
	"github.com/ipni/go-libipni/announce/httpsender"
	"github.com/ipni/go-libipni/dagsync/ipnisync/head"
	"github.com/ipni/go-libipni/ingest/schema"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-ucanto/core/ipld"
)

type AdvertisementPublisher struct {
	*options
	pendingAds []schema.Advertisement
	sender     announce.Sender
	key        crypto.PrivKey
	store      store.PublisherStore
}

func NewAdvertisementPublisher(id crypto.PrivKey, store store.PublisherStore, opts ...Option) (*AdvertisementPublisher, error) {
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
	batchPublisher := &AdvertisementPublisher{
		options: o,
		key:     id,
		store:   store,
	}
	if len(o.announceURLs) > 0 {
		sender, err := httpsender.New(o.announceURLs, peer)
		if err != nil {
			return nil, fmt.Errorf("cannot create http announce sender: %w", err)
		}
		log.Info("HTTP announcements enabled")
		batchPublisher.sender = sender
	}
	return batchPublisher, nil
}

func (p *AdvertisementPublisher) AddToBatch(adv schema.Advertisement) error {
	p.pendingAds = append(p.pendingAds, adv)
	return nil
}

func (p *AdvertisementPublisher) Commit(ctx context.Context) (ipld.Link, error) {
	pendingAds := p.pendingAds
	p.pendingAds = nil
	lnk, err := p.commit(ctx, pendingAds)
	if err != nil {
		for _, adv := range pendingAds {
			if !adv.IsRm {
				peer, err := peer.Decode(adv.Provider)
				if err == nil {
					_ = p.store.DeleteChunkLinkForProviderAndContextID(ctx, peer, adv.ContextID)
				}
			}
		}
		return nil, err
	}
	return lnk, nil
}
func (p *AdvertisementPublisher) commit(ctx context.Context, pendingAds []schema.Advertisement) (ipld.Link, error) {

	// Get the previous advertisement that was generated.
	prevHead, err := p.store.Head(ctx)
	if err != nil {
		if !store.IsNotFound(err) {
			return nil, fmt.Errorf("could not get latest advertisement: %s", err)
		}
	}
	var prevLink ipld.Link
	// Check for cid.Undef for the previous link. If this is the case, then
	// this means there are no previous advertisements.
	if prevHead == nil {
		log.Info("Latest advertisement CID was undefined - no previous advertisement")
	} else {
		prevLink = prevHead.Head
	}

	if len(pendingAds) == 0 {
		log.Info("No pending advertisements to commit")
		return prevLink, nil
	}

	// Store all pending advertisements in order, linking each to the previous.
	for _, adv := range pendingAds {
		adv.PreviousID = prevLink

		// Sign the advertisement.
		if err = adv.Sign(p.key); err != nil {
			return nil, err
		}

		if err := adv.Validate(); err != nil {
			return nil, err
		}

		lnk, err := p.store.PutAdvert(ctx, adv)
		if err != nil {
			return nil, err
		}
		log.Info("Stored ad in local link system")
		prevLink = lnk
	}

	lnk := prevLink
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

	if p.sender != nil {
		err = announce.Send(ctx, lnk.(cidlink.Link).Cid, p.pubHTTPAnnounceAddrs, p.sender)
		if err != nil {
			log.Warnw("Failed to announce advertisement", "err", err)
		}
	}

	return lnk, nil
}
