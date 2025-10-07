package notifier

import (
	"context"
	"fmt"

	"github.com/ipfs/go-cid"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
	"github.com/storacha/go-ucanto/core/ipld"
)

const remoteHeadPrefix = "head/remote/"

type HeadState struct {
	ds     store.Store
	hdkey  string
	cached ipld.Link
}

func NewHeadState(ds store.Store, hostname string) (*HeadState, error) {
	var hd ipld.Link
	hdkey := remoteHeadPrefix + hostname
	v, err := ds.Get(context.Background(), hdkey)
	if err != nil {
		if !store.IsNotFound(err) {
			return nil, fmt.Errorf("getting remote IPNI head CID from datastore: %w", err)
		}
	} else {
		c, err := cid.Cast(v)
		if err != nil {
			return nil, fmt.Errorf("parsing remote IPNI head CID: %w", err)
		}
		hd = cidlink.Link{Cid: c}
	}
	return &HeadState{ds: ds, hdkey: hdkey, cached: hd}, nil
}

func (h *HeadState) Get(ctx context.Context) ipld.Link {
	return h.cached
}

func (h *HeadState) Set(ctx context.Context, head ipld.Link) error {
	err := h.ds.Put(ctx, h.hdkey, []byte(head.Binary()))
	if err != nil {
		return fmt.Errorf("saving remote IPNI sync'd head: %w", err)
	}
	h.cached = head
	return nil
}
