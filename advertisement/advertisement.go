package advertisement

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipni/go-libipni/maurl"
	mh "github.com/multiformats/go-multihash"
	"github.com/storacha/go-ucanto/did"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/storacha/go-libstoracha/capabilities/assert"
	"github.com/storacha/go-libstoracha/digestutil"
)

const (
	BlobUrlPlaceholder    = "{blob}"
	BlobCIDUrlPlaceholder = "{blobCID}"
)

// Encode canonically encodes ContextID data.
func EncodeContextID(space did.DID, digest mh.Multihash) ([]byte, error) {
	return mh.Sum(bytes.Join([][]byte{space.Bytes(), digest}, nil), mh.SHA2_256, -1)
}

// ShardCID extracts an alternate shard CID from the provider & location URLs in a location claim
func ShardCID(provider peer.AddrInfo, caveats assert.LocationCaveats) (*cid.Cid, error) {

	// analyze each provider address
	for _, addr := range provider.Addrs {
		// first, attempt to convert the addr to a url scheme
		url, err := maurl.ToURL(addr)
		// if it can't be converted, skip
		if err != nil {
			continue
		}
		// must be an http url
		if !(url.Scheme == "http" || url.Scheme == "https") {
			continue
		}
		// if it does not have replaceable components in a location url, skip
		if !strings.Contains(url.Path, BlobUrlPlaceholder) && !strings.Contains(url.Path, BlobCIDUrlPlaceholder) {
			continue
		}
		// generate a regex to capture matching components of the url
		urlRegex, err := urlToRegex(url)
		if err != nil {
			return nil, fmt.Errorf("parsing url regex: %w", err)
		}
		// go through each location in the claim
		for _, location := range caveats.Location {
			// if the location does not match the url regex, skip
			if !urlRegex.MatchString(location.String()) {
				continue
			}
			// get matching components of the location
			matches := urlRegex.FindStringSubmatch(location.String())
			var blob mh.Multihash
			// check for matches with the blob multihash
			if urlRegex.SubexpIndex("blob") != -1 {
				blob, err = digestutil.Parse(matches[urlRegex.SubexpIndex("blob")])
				if err != nil {
					return nil, fmt.Errorf("location format has invalid multihash: %w", err)
				}
				// if no blobCID, just use the multihash
				if urlRegex.SubexpIndex("blobCID") == -1 {
					// if location hash matches hash in the url there's no need to save a shard
					if caveats.Content.Hash().String() == blob.String() {
						return nil, nil
					}
					shard := cid.NewCidV1(cid.Raw, blob)
					return &shard, nil
				}
			}

			// check for matches with the blob cid
			blobCID, err := cid.Decode(matches[urlRegex.SubexpIndex("blobCID")])
			if err != nil {
				return nil, fmt.Errorf("location format has invalid cid: %w", err)
			}

			// if there are multihash matches and cid matches they should be equal
			if blob != nil && blob.String() != blobCID.Hash().String() {
				return nil, fmt.Errorf("location format has both cid and multihash but they do not match")
			}

			// if location hash matches the cid in the URL, just use that
			if cid.NewCidV1(cid.Raw, caveats.Content.Hash()).Equals(blobCID) {
				return nil, nil
			}

			return &blobCID, nil
		}
	}
	return nil, nil
}

func urlToRegex(u *url.URL) (*regexp.Regexp, error) {
	escapedBlobString := url.PathEscape(BlobUrlPlaceholder)
	escapedBlobCIDString := url.PathEscape(BlobCIDUrlPlaceholder)

	regexPattern := strings.ReplaceAll(regexp.QuoteMeta(u.String()), escapedBlobString, `(?P<blob>.+?)`)
	regexPattern = strings.ReplaceAll(regexPattern, escapedBlobCIDString, `(?P<blobCID>.+?)`)
	return regexp.Compile("^" + regexPattern + "$")
}
