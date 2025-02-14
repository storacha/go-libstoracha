package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipld/go-ipld-prime"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/storacha/ipni-publisher/pkg/notifier"
	"github.com/storacha/ipni-publisher/pkg/publisher"
	"github.com/storacha/ipni-publisher/pkg/server"
	"github.com/storacha/ipni-publisher/pkg/store"
)

var ipniNamespace = datastore.NewKey("ipni/")
var publisherNamespace = ipniNamespace.ChildString("publisher/")
var notifierNamespace = ipniNamespace.ChildString("notifier/")

func TestExample(t *testing.T) {
	priv, _, _ := crypto.GenerateEd25519Key(nil)

	// Setup publisher
	ds := datastore.NewMapDatastore()
	publisherStore := store.FromDatastore(namespace.Wrap(ds, publisherNamespace))
	publisher, _ := publisher.New(
		priv,
		publisherStore,
		publisher.WithDirectAnnounce("https://cid.contact/announce"),
		publisher.WithAnnounceAddrs("/dns4/localhost/tcp/3000/https"),
	)

	// Setup and start HTTP server (optional, but required if announce addresses configured)
	encodableStore, _ := publisherStore.(store.EncodeableStore)
	srv, _ := server.NewServer(encodableStore, server.WithHTTPListenAddrs("localhost:3000"))
	srv.Start(context.Background())
	defer srv.Shutdown(context.Background())

	// Setup remote sync notifications (optional)
	notifierStore := store.SimpleStoreFromDatastore(namespace.Wrap(ds, notifierNamespace))
	notifier, _ := notifier.NewNotifierWithStorage("https://cid.contact/", priv, notifierStore)
	notifier.Start(context.Background())
	defer notifier.Stop()

	notifier.Notify(func(ctx context.Context, head, prev ipld.Link) {
		fmt.Printf("remote sync from %s to %s\n", prev, head)
	})

	// Setup complete! Now publish an advert:
	// publisher.Publish(context.Background(), ...)

	fmt.Println(publisher)
}
