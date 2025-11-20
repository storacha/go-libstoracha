package publisher

import (
	"context"
	"encoding/base64"
	"fmt"
	"iter"

	"github.com/ipni/go-libipni/ingest/schema"
	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
)

// GenerateAd generates an advertisement for the given parameters.
func GenerateAd(ctx context.Context, publisherStore store.PublisherStore, peer peer.ID, addrs []multiaddr.Multiaddr, contextID []byte, md metadata.Metadata, isRm bool, mhs iter.Seq[mh.Multihash]) (schema.Advertisement, error) {
	var err error

	log := log.With("providerID", peer).With("contextID", base64.StdEncoding.EncodeToString(contextID))

	chunkLink, err := publisherStore.ChunkLinkForProviderAndContextID(ctx, peer, contextID)
	if err != nil {
		if !store.IsNotFound(err) {
			return schema.Advertisement{}, fmt.Errorf("could not get entries cid by provider + context id: %s", err)
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
			chunkLink, err = publisherStore.PutEntries(ctx, mhs)
			if err != nil {
				return schema.Advertisement{}, fmt.Errorf("could not generate entries list: %s", err)
			}
			if chunkLink == nil {
				log.Warnw("chunking for context ID resulted in no link", "contextID", contextID)
				chunkLink = schema.NoEntries
			}

			// Store the relationship between providerID, contextID and CID of the
			// advertised list of Cids.
			err = publisherStore.PutChunkLinkForProviderAndContextID(ctx, peer, contextID, chunkLink)
			if err != nil {
				return schema.Advertisement{}, fmt.Errorf("failed to write provider + context id to entries cid mapping: %s", err)
			}
		} else {
			// Lookup metadata for this providerID and contextID.
			prevMetadata, err := publisherStore.MetadataForProviderAndContextID(ctx, peer, contextID)
			if err != nil {
				if !store.IsNotFound(err) {
					return schema.Advertisement{}, fmt.Errorf("could not get metadata for provider + context id: %s", err)
				}
				log.Warn("No metadata for existing provider + context ID, generating new advertisement")
			}

			if md.Equal(prevMetadata) {
				// Metadata is the same; no change, no need for new
				// advertisement.
				return schema.Advertisement{}, ErrAlreadyAdvertised
			}

			// Linked list is the same, but metadata is different, so generate
			// new advertisement with same linked list, but new metadata.
		}

		if err = publisherStore.PutMetadataForProviderAndContextID(ctx, peer, contextID, md); err != nil {
			return schema.Advertisement{}, fmt.Errorf("failed to write provider + context id to metadata mapping: %s", err)
		}
	} else {
		log.Info("Creating removal advertisement")

		if chunkLink == nil {
			return schema.Advertisement{}, ErrContextIDNotFound
		}

		// If removing by context ID, it means the list of CIDs is not needed
		// anymore, so we can remove the entry from the datastore.
		err = publisherStore.DeleteChunkLinkForProviderAndContextID(ctx, peer, contextID)
		if err != nil {
			return schema.Advertisement{}, fmt.Errorf("failed to delete provider + context id to entries cid mapping: %s", err)
		}
		err = publisherStore.DeleteMetadataForProviderAndContextID(ctx, peer, contextID)
		if err != nil {
			return schema.Advertisement{}, fmt.Errorf("failed to delete provider + context id to metadata mapping: %s", err)
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
		return schema.Advertisement{}, err
	}

	var stringAddrs []string
	for _, addr := range addrs {
		stringAddrs = append(stringAddrs, addr.String())
	}

	return schema.Advertisement{
		Provider:  peer.String(),
		Addresses: stringAddrs,
		Entries:   chunkLink,
		ContextID: contextID,
		Metadata:  mdBytes,
		IsRm:      isRm,
	}, nil

}
