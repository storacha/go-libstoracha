package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/namespace"
	"github.com/ipld/go-ipld-prime"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/storacha/go-libstoracha/ipnipublisher/notifier"
	"github.com/storacha/go-libstoracha/ipnipublisher/publisher"
	"github.com/storacha/go-libstoracha/ipnipublisher/server"
	"github.com/storacha/go-libstoracha/ipnipublisher/store"
)

var ipniNamespace = datastore.NewKey("ipni/")
var publisherNamespace = ipniNamespace.ChildString("publisher/")
var notifierNamespace = ipniNamespace.ChildString("notifier/")

func main() {
	ctx := context.Background()

	// generate a key pair
	priv, _, err := crypto.GenerateEd25519Key(nil)
	if err != nil {
		log.Fatalf("generating key pair: %s", err)
	}

	// setup datastore(s)
	ds := datastore.NewMapDatastore()
	publisherStore := store.FromDatastore(namespace.Wrap(ds, publisherNamespace))

	// Setup publisher
	p, err := publisher.New(
		priv,
		publisherStore,
		publisher.WithDirectAnnounce("https://cid.contact/announce"),
		publisher.WithAnnounceAddrs("/dns4/localhost/tcp/3000/https"),
	)
	if err != nil {
		log.Fatalf("creating publisher: %s", err)
	}

	// Setup and start HTTP server (optional, but required if announce addresses configured).
	encodableStore, _ := publisherStore.(store.EncodeableStore)
	srv, err := server.NewServer(encodableStore, server.WithHTTPListenAddrs("localhost:3000"))
	if err != nil {
		log.Fatalf("creating server: %s", err)
	}
	if err := srv.Start(ctx); err != nil {
		log.Fatalf("starting server: %s", err)
	}
	defer func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("shutting down server: %s", err)
		}
	}()

	// Setup remote sync notifications (optional).
	notifierStore := store.SimpleStoreFromDatastore(namespace.Wrap(ds, notifierNamespace))
	notif, err := notifier.NewNotifierWithStorage("https://cid.contact/", priv, notifierStore)
	if err != nil {
		log.Fatalf("creating notifier: %s", err)
	}
	notif.Start(ctx)
	defer notif.Stop()

	notif.Notify(func(ctx context.Context, head, prev ipld.Link) {
		fmt.Printf("remote sync from %s to %s\n", prev, head)
	})

	// Setup complete! Now publish an advert via the `Publish` method.
	// p.Publish(ctx, ...)

	fmt.Printf("Publisher: %+v\n", p)
}
