package blob

import (
	"bytes"
	"fmt"

	"github.com/multiformats/go-multibase"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
)

func equalWith(claimed, delegated ucan.Resource) failure.Failure {
	if claimed != delegated {
		return schema.NewSchemaError(fmt.Sprintf("Resource '%s' doesn't match delegated '%s'", claimed, delegated))
	}

	return nil
}

func equalBlob(claimed, delegated types.Blob) failure.Failure {
	// Check if the blob digest matches
	if !bytes.Equal(delegated.Digest, claimed.Digest) {
		claimedDigest, _ := multibase.Encode(multibase.Base58BTC, claimed.Digest)
		delegatedDigest, _ := multibase.Encode(multibase.Base58BTC, delegated.Digest)

		return schema.NewSchemaError(fmt.Sprintf(
			"Link %s violates imposed %s constraint",
			claimedDigest, delegatedDigest,
		))
	}

	// Check size constraint
	if claimed.Size > 0 && delegated.Size > 0 {
		if claimed.Size > delegated.Size {
			return schema.NewSchemaError(fmt.Sprintf(
				"Size constraint violation: %d > %d",
				claimed.Size, delegated.Size,
			))
		}
	}

	return nil
}

func checkLink(claimed, delegated ucan.Link) failure.Failure {
	return checkConstraint(claimed.String(), delegated.String(), "link")
}

func checkSpace(claimed, delegated ucan.Resource) failure.Failure {
	return checkConstraint(claimed, delegated, "space")
}

func checkConstraint(claimed, delegated string, constraint string) failure.Failure {
	if delegated == "" || delegated == "*" {
		return nil
	}

	if claimed != delegated {
		return schema.NewSchemaError(fmt.Sprintf("%s '%s' doesn't match delegated '%s'", constraint, claimed, delegated))
	}

	return nil
}

func equalTTL(claimed, delegated int) failure.Failure {
	if claimed != delegated {
		return schema.NewSchemaError(fmt.Sprintf("TTL '%d' doesn't match delegated '%d'", claimed, delegated))
	}

	return nil
}
