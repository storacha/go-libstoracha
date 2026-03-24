package encryption

import (
	"bytes"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-libstoracha/capabilities/utils"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/core/receipt"
	"github.com/storacha/go-ucanto/core/result/failure"
	"github.com/storacha/go-ucanto/core/schema"
	"github.com/storacha/go-ucanto/ucan"
	"github.com/storacha/go-ucanto/validator"
)

const KeyDecryptAbility = "space/encryption/key/decrypt"

// KeyDecryptCaveats represents the caveats for a space/encryption/key/decrypt invocation.
type KeyDecryptCaveats struct {
	// Key is the wrapped DEK blob (RSA-OAEP or AES-GCM ciphertext) to be unwrapped.
	Key []byte
}

func (kdc KeyDecryptCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&kdc, KeyDecryptCaveatsType(), types.Converters...)
}

var KeyDecryptCaveatsReader = schema.Struct[KeyDecryptCaveats](KeyDecryptCaveatsType(), nil, types.Converters...)

// KeyDecryptOk represents the result of a successful space/encryption/key/decrypt invocation.
type KeyDecryptOk struct {
	// DecryptedSymmetricKey is the base64-encoded CBOR({path, dek}) payload returned by the KMS.
	DecryptedSymmetricKey string
}

func (kdo KeyDecryptOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&kdo, KeyDecryptOkType(), types.Converters...)
}

type KeyDecryptReceipt receipt.Receipt[KeyDecryptOk, failure.Failure]
type KeyDecryptReceiptReader receipt.ReceiptReader[KeyDecryptOk, failure.Failure]

func NewKeyDecryptReceiptReader() (KeyDecryptReceiptReader, error) {
	return receipt.NewReceiptReader[KeyDecryptOk, failure.Failure](encryptionSchema, types.Converters...)
}

var KeyDecryptOkReader = schema.Struct[KeyDecryptOk](KeyDecryptOkType(), nil, types.Converters...)

// KeyDecrypt is a capability that requests the KMS to unwrap a wrapped DEK blob.
var KeyDecrypt = validator.NewCapability(
	KeyDecryptAbility,
	schema.DIDString(),
	KeyDecryptCaveatsReader,
	func(claimed, delegated ucan.Capability[KeyDecryptCaveats]) failure.Failure {
		if fail := utils.EqualWith(claimed, delegated); fail != nil {
			return fail
		}
		if !bytes.Equal(claimed.Nb().Key, delegated.Nb().Key) {
			return schema.NewSchemaError("Constraint violation: 'key' mismatch")
		}
		return nil
	},
)
