package store

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"sync"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	logging "github.com/ipfs/go-log/v2"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/ipni/go-libipni/dagsync/ipnisync/head"
	"github.com/ipni/go-libipni/ingest/schema"
	"github.com/ipni/go-libipni/metadata"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/core/ipld/block"
	"github.com/storacha/go-ucanto/core/ipld/codec/cbor"
	"github.com/storacha/go-ucanto/core/ipld/codec/json"
	"github.com/storacha/go-ucanto/core/ipld/hash/sha256"
)

var log = logging.Logger("store")

type ErrNotFound struct {
	underlying error
}

func NewErrNotFound(underlying error) ErrNotFound {
	return ErrNotFound{underlying: underlying}
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("unable to find underlying: %s", e.underlying.Error())
}

func (e ErrNotFound) Unwrap() error {
	return e.underlying
}

// ErrPreconditionFailed is returned when a conditional write fails.
var ErrPreconditionFailed = errors.New("precondition failed")

const (
	keyToMetadataMapPrefix  = "map/keyMD/"
	keyToChunkLinkMapPrefix = "map/keyChunkLink/"
	headKey                 = "head"
)

// MaxEntryChunkSize is the maximum number of multihashes each advertisement
// entry chunk may contain.
var MaxEntryChunkSize = 16384

type SimpleStore interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Put(ctx context.Context, key string, len uint64, data io.Reader) error
}

type Store interface {
	SimpleStore
	// Replace allows conditional writes to the store. It returns
	// [ErrPreconditionFailed] when a write fails due to the old data not matching
	// the passed value. Note: old may be nil, when no old data exists.
	Replace(ctx context.Context, key string, old io.Reader, len uint64, new io.Reader) error
}

type ProviderContextTable interface {
	Get(ctx context.Context, p peer.ID, contextID []byte) ([]byte, error)
	Put(ctx context.Context, p peer.ID, contextID []byte, data []byte) error
	Delete(ctx context.Context, p peer.ID, contextID []byte) error
}

type EncodeableStore interface {
	Encode(ctx context.Context, id ipld.Link, w io.Writer) error
	EncodeHead(ctx context.Context, w io.Writer) error
}

type AdvertReadable interface {
	// Advert retrieves an existing advert from the store.
	Advert(ctx context.Context, id ipld.Link) (schema.Advertisement, error)
}

type AdvertWritable interface {
	PutAdvert(ctx context.Context, ad schema.Advertisement) (ipld.Link, error)
}

type AdvertStore interface {
	AdvertReadable
	AdvertWritable
}

type EntriesReadable interface {
	// Entries returns an iterable of multihashes from the store for the
	// given root of an existing advertisement entries chain.
	Entries(ctx context.Context, root ipld.Link) iter.Seq2[multihash.Multihash, error]
}

type EntriesWritable interface {
	// PutEntries writes a given set of multihash entries to do the store and returns the root cid
	PutEntries(ctx context.Context, entries iter.Seq[multihash.Multihash]) (ipld.Link, error)
}

type EntriesStore interface {
	EntriesReadable
	EntriesWritable
}

type HeadStore interface {
	Head(ctx context.Context) (*head.SignedHead, error)
	// ReplaceHead swaps the previous head for the provided new head. Note the
	// previous head may be nil when no head exists yet.
	//
	// It allows an implementation to perform conditional writes to mitigate
	// against losing an advert when 2 or more are published concurrently.
	//
	// [ErrPreconditionFailed] is returned when the write fails due to the current
	// head not matching the passed old head.
	ReplaceHead(ctx context.Context, oldHead *head.SignedHead, newHead *head.SignedHead) (ipld.Link, error)
}

type ChunkLinkStore interface {
	ChunkLinkForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) (ipld.Link, error)
	PutChunkLinkForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte, adCid ipld.Link) error
	DeleteChunkLinkForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) error
}

type MetadataStore interface {
	MetadataForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) (metadata.Metadata, error)
	PutMetadataForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte, md metadata.Metadata) error
	DeleteMetadataForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) error
}

type PublisherStore interface {
	AdvertStore
	EntriesStore
	HeadStore
	ChunkLinkStore
	MetadataStore
}

type FullStore interface {
	EncodeableStore
	PublisherStore
}

type AdStore struct {
	store           Store
	chunkLinks      ProviderContextTable
	metadata        ProviderContextTable
	metadataContext metadata.MetadataContext
}

var _ FullStore = (*AdStore)(nil)

func (s *AdStore) PutAdvert(ctx context.Context, ad schema.Advertisement) (ipld.Link, error) {
	return PutAdvert(ctx, s.store, ad)
}

func (s *AdStore) Advert(ctx context.Context, id ipld.Link) (schema.Advertisement, error) {
	return Advert(ctx, s.store, id)
}

func (s *AdStore) Entries(ctx context.Context, root ipld.Link) iter.Seq2[multihash.Multihash, error] {
	return Entries(ctx, s.store, root)
}

func (s *AdStore) PutEntries(ctx context.Context, mhs iter.Seq[multihash.Multihash]) (ipld.Link, error) {
	return PutEntries(ctx, s.store, mhs, MaxEntryChunkSize)
}

func (s *AdStore) Encode(ctx context.Context, id datamodel.Link, w io.Writer) error {
	return Encode(ctx, s.store, id, w)
}

func (s *AdStore) ChunkLinkForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) (datamodel.Link, error) {
	return ChunkLink(ctx, s.chunkLinks, p, contextID)
}

func (s *AdStore) PutChunkLinkForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte, chunkLink datamodel.Link) error {
	return PutChunkLink(ctx, s.chunkLinks, p, contextID, chunkLink)
}

func (s *AdStore) DeleteChunkLinkForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) error {
	return s.chunkLinks.Delete(ctx, p, contextID)
}

func (s *AdStore) Head(ctx context.Context) (*head.SignedHead, error) {
	return Head(ctx, s.store)
}

func (s *AdStore) ReplaceHead(ctx context.Context, oldHead *head.SignedHead, newHead *head.SignedHead) (datamodel.Link, error) {
	return ReplaceHead(ctx, s.store, oldHead, newHead)
}

func (s *AdStore) EncodeHead(ctx context.Context, w io.Writer) error {
	return EncodeHead(ctx, s.store, w)
}

func (s *AdStore) MetadataForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) (metadata.Metadata, error) {
	return Metadata(ctx, s.metadataContext, s.metadata, p, contextID)
}

func (s *AdStore) PutMetadataForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte, md metadata.Metadata) error {
	return PutMetadata(ctx, s.metadata, p, contextID, md)
}

func (s *AdStore) DeleteMetadataForProviderAndContextID(ctx context.Context, p peer.ID, contextID []byte) error {
	return s.metadata.Delete(ctx, p, contextID)
}

func NewPublisherStore(store Store, chunkLinks, metadataTable ProviderContextTable, opts ...Option) *AdStore {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	mctx := o.metadataContext
	if mctx == nil {
		mctx = metadata.Default
	}
	return &AdStore{store, chunkLinks, metadataTable, mctx}
}

func Advert(ctx context.Context, ds SimpleStore, id ipld.Link) (schema.Advertisement, error) {
	var ad schema.Advertisement
	r, err := ds.Get(ctx, id.String())
	if err != nil {
		return ad, err
	}
	defer r.Close()
	v, err := io.ReadAll(r)
	if err != nil {
		return ad, err
	}
	ad, err = schema.BytesToAdvertisement(asCID(id), v)
	if err != nil {
		return ad, err
	}
	return ad, nil
}

func PutAdvert(ctx context.Context, ds SimpleStore, adv schema.Advertisement) (ipld.Link, error) {
	return store(ctx, ds, &adv, schema.AdvertisementPrototype.Type())
}

func PutEntries(ctx context.Context, ds SimpleStore, entries iter.Seq[multihash.Multihash], chunkSize int) (next ipld.Link, err error) {
	mhs := make([]multihash.Multihash, 0, chunkSize)
	var mhCount, chunkCount int
	for mh := range entries {
		mhs = append(mhs, mh)
		mhCount++
		if len(mhs) >= chunkSize {
			next, err = store(ctx, ds, toChunk(mhs, next), schema.EntryChunkPrototype.Type())
			if err != nil {
				return nil, err
			}
			chunkCount++
			// NewLinkedListOfMhs makes it own copy, so safe to reuse mhs
			mhs = mhs[:0]
		}
	}
	if len(mhs) != 0 {
		next, err = store(ctx, ds, toChunk(mhs, next), schema.EntryChunkPrototype.Type())
		if err != nil {
			return nil, err
		}
		chunkCount++
	}

	log.Infow("Generated linked chunks of multihashes", "totalMhCount", mhCount, "chunkCount", chunkCount)
	return next, nil
}

func Entries(ctx context.Context, ds SimpleStore, root ipld.Link) iter.Seq2[multihash.Multihash, error] {
	return func(yield func(multihash.Multihash, error) bool) {
		cur := root
		for cur != nil && cur != schema.NoEntries {
			r, err := ds.Get(ctx, cur.String())
			if err != nil {
				yield(nil, err)
				return
			}
			defer r.Close()
			v, err := io.ReadAll(r)
			if err != nil {
				yield(nil, err)
				return
			}
			ent, err := schema.BytesToEntryChunk(asCID(cur), v)
			if err != nil {
				yield(nil, err)
				return
			}

			for _, d := range ent.Entries {
				if !yield(d, nil) {
					return
				}
			}

			cur = ent.Next
		}
	}
}

func Encode(ctx context.Context, ds SimpleStore, lnk ipld.Link, w io.Writer) error {
	r, err := ds.Get(ctx, lnk.String())
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(w, r)
	return err
}

func Head(ctx context.Context, ds SimpleStore) (*head.SignedHead, error) {
	r, err := ds.Get(ctx, headKey)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return head.Decode(r)
}

func EncodeHead(ctx context.Context, ds SimpleStore, w io.Writer) error {
	r, err := ds.Get(ctx, headKey)
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(w, r)
	return err
}

func ReplaceHead(ctx context.Context, ds Store, oldHead *head.SignedHead, newHead *head.SignedHead) (datamodel.Link, error) {
	blk, err := block.Encode(newHead, head.SignedHeadPrototype.Type(), json.Codec, sha256.Hasher)
	if err != nil {
		return nil, err
	}
	newBytes := blk.Bytes()
	var oldBytesReader io.Reader
	if oldHead != nil {
		blk, err := block.Encode(oldHead, head.SignedHeadPrototype.Type(), json.Codec, sha256.Hasher)
		if err != nil {
			return nil, err
		}
		oldBytesReader = bytes.NewReader(blk.Bytes())
	}
	err = ds.Replace(ctx, headKey, oldBytesReader, uint64(len(newBytes)), bytes.NewReader(newBytes))
	if err != nil {
		return nil, err
	}
	return blk.Link(), nil
}

func ChunkLink(ctx context.Context, ds ProviderContextTable, p peer.ID, contextID []byte) (datamodel.Link, error) {
	data, err := ds.Get(ctx, p, contextID)
	if err != nil {
		return nil, err
	}
	_, c, err := cid.CidFromBytes(data)
	if err != nil {
		return nil, err
	}
	return cidlink.Link{Cid: c}, nil
}

func PutChunkLink(ctx context.Context, ds ProviderContextTable, p peer.ID, contextID []byte, lnk datamodel.Link) error {
	return ds.Put(ctx, p, contextID, []byte(lnk.Binary()))
}

func Metadata(ctx context.Context, mctx metadata.MetadataContext, ds ProviderContextTable, p peer.ID, contextID []byte) (metadata.Metadata, error) {
	md := mctx.New()
	data, err := ds.Get(ctx, p, contextID)
	if err != nil {
		return md, err
	}
	if err := md.UnmarshalBinary(data); err != nil {
		return md, err
	}
	return md, nil
}

func PutMetadata(ctx context.Context, ds ProviderContextTable, p peer.ID, contextID []byte, md metadata.Metadata) error {
	data, err := md.MarshalBinary()
	if err != nil {
		return err
	}
	return ds.Put(ctx, p, contextID, data)
}

func store(ctx context.Context, ds SimpleStore, value any, typ ipldschema.Type) (ipld.Link, error) {
	blk, err := block.Encode(value, typ, cbor.Codec, sha256.Hasher)
	if err != nil {
		return nil, err
	}
	err = ds.Put(ctx, blk.Link().String(), uint64(len(blk.Bytes())), bytes.NewReader(blk.Bytes()))
	if err != nil {
		return nil, err
	}
	return blk.Link(), nil
}

func toChunk(mhs []multihash.Multihash, next ipld.Link) *schema.EntryChunk {
	chunk := schema.EntryChunk{
		Entries: mhs,
	}
	if next != nil {
		chunk.Next = next
	}
	return &chunk
}

func IsNotFound(err error) bool {
	// solve for the unfortuante lack of standards on not found errors
	var errNotFound ErrNotFound
	return errors.Is(err, datastore.ErrNotFound) || errors.As(err, &errNotFound)
}

func providerContextKey(provider peer.ID, contextID []byte) datastore.Key {
	contextKey, _ := multibase.Encode(multibase.Base58BTC, contextID)
	return datastore.NewKey(provider.String() + "/" + contextKey)
}

type dsProviderContextTable struct {
	ds datastore.Datastore
}

func NewDatastoreProviderContextTable(ds datastore.Datastore) ProviderContextTable {
	return &dsProviderContextTable{ds}
}

func (d *dsProviderContextTable) Delete(ctx context.Context, p peer.ID, contextID []byte) error {
	return d.ds.Delete(ctx, providerContextKey(p, contextID))
}

func (d *dsProviderContextTable) Get(ctx context.Context, p peer.ID, contextID []byte) ([]byte, error) {
	return d.ds.Get(ctx, providerContextKey(p, contextID))
}

func (d *dsProviderContextTable) Put(ctx context.Context, p peer.ID, contextID []byte, data []byte) error {
	return d.ds.Put(ctx, providerContextKey(p, contextID), data)
}

var _ ProviderContextTable = (*dsProviderContextTable)(nil)

type dsStoreAdapter struct {
	ds    datastore.Datastore
	mutex sync.Mutex
}

// Get implements Store.
func (d *dsStoreAdapter) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	data, err := d.ds.Get(ctx, datastore.NewKey(key))
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

// Put implements Store.
func (d *dsStoreAdapter) Put(ctx context.Context, key string, len uint64, r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.ds.Put(ctx, datastore.NewKey(key), data)
}

func (d *dsStoreAdapter) Replace(ctx context.Context, key string, old io.Reader, newLen uint64, new io.Reader) error {
	var oldBytes []byte
	if old != nil {
		b, err := io.ReadAll(old)
		if err != nil {
			return err
		}
		oldBytes = b
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	cur, err := d.ds.Get(ctx, datastore.NewKey(key))
	if err != nil {
		if !errors.Is(err, datastore.ErrNotFound) {
			return err
		}
		// if error was not found, but old is non-nil then something went bad
		if old != nil {
			return ErrPreconditionFailed
		}
	}

	if !bytes.Equal(cur, oldBytes) {
		return ErrPreconditionFailed
	}

	newBytes, err := io.ReadAll(new)
	if err != nil {
		return err
	}

	return d.ds.Put(ctx, datastore.NewKey(key), newBytes)
}

var _ Store = (*dsStoreAdapter)(nil)

type directoryStore struct {
	directory string
	mutex     sync.Mutex
}

// Get implements Store.
func (d *directoryStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	path, err := filepath.Abs(filepath.Join(d.directory, key))
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

// Put implements Store.
func (d *directoryStore) Put(ctx context.Context, key string, len uint64, data io.Reader) error {
	path, err := filepath.Abs(filepath.Join(d.directory, key))
	if err != nil {
		return err
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, data)
	return err
}

func (d *directoryStore) Replace(ctx context.Context, key string, old io.Reader, newLen uint64, new io.Reader) error {
	var oldBytes []byte
	if old != nil {
		b, err := io.ReadAll(old)
		if err != nil {
			return err
		}
		oldBytes = b
	}

	path, err := filepath.Abs(filepath.Join(d.directory, key))
	if err != nil {
		return err
	}
	d.mutex.Lock()
	defer d.mutex.Unlock()

	cur, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		// if error was not found, but old is non-nil then something went bad
		if old != nil {
			return ErrPreconditionFailed
		}
	}
	if !bytes.Equal(cur, oldBytes) {
		return ErrPreconditionFailed
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, new)
	return err
}

var _ Store = (*directoryStore)(nil)

func asCID(link ipld.Link) cid.Cid {
	if cl, ok := link.(cidlink.Link); ok {
		return cl.Cid
	}
	return cid.MustParse(link.String())
}

func SimpleStoreFromDatastore(ds datastore.Datastore) SimpleStore {
	return &dsStoreAdapter{ds: ds}
}

func FromDatastore(ds datastore.Datastore, opts ...Option) FullStore {
	return NewPublisherStore(
		&dsStoreAdapter{ds: ds},
		&dsProviderContextTable{namespace.Wrap(ds, datastore.NewKey(keyToChunkLinkMapPrefix))},
		&dsProviderContextTable{namespace.Wrap(ds, datastore.NewKey(keyToMetadataMapPrefix))},
		opts...,
	)
}

func FromLocalStore(storagePath string, ds datastore.Datastore, opts ...Option) FullStore {
	store := &directoryStore{directory: storagePath}
	chunkLinksStore := &dsProviderContextTable{namespace.Wrap(ds, datastore.NewKey(keyToChunkLinkMapPrefix))}
	mdStore := &dsProviderContextTable{namespace.Wrap(ds, datastore.NewKey(keyToMetadataMapPrefix))}
	return NewPublisherStore(store, chunkLinksStore, mdStore, opts...)
}
