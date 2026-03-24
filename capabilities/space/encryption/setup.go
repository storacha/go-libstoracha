package encryption

import (
	"fmt"

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

const SetupAbility = "space/encryption/setup"

// SetupCaveats represents the caveats for a space/encryption/setup invocation.
type SetupCaveats struct {
	// The location of the KMS key to use for encryption. If not provided, the Storacha Key Manager will use the default location.
	Location *string
	// The keyring of the KMS key to use for encryption. If not provided, the Storacha Key Manager will use the default keyring.
	Keyring *string
}

func (sc SetupCaveats) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&sc, SetupCaveatsType(), types.Converters...)
}

var SetupCaveatsReader = schema.Struct[SetupCaveats](SetupCaveatsType(), nil, types.Converters...)

// SetupOk represents the result of a successful space/encryption/setup invocation.
type SetupOk struct {
	// PublicKey is the PEM-encoded RSA public key for DEK wrapping.
	PublicKey string
	// Provider is the KMS provider identifier (e.g. "storacha").
	Provider string
	// Algorithm is the wrapping algorithm (e.g. "RSA-OAEP-256").
	Algorithm string
}

func (so SetupOk) ToIPLD() (datamodel.Node, error) {
	return ipld.WrapWithRecovery(&so, SetupOkType(), types.Converters...)
}

type SetupReceipt receipt.Receipt[SetupOk, failure.Failure]
type SetupReceiptReader receipt.ReceiptReader[SetupOk, failure.Failure]

func NewSetupReceiptReader() (SetupReceiptReader, error) {
	return receipt.NewReceiptReader[SetupOk, failure.Failure](encryptionSchema, types.Converters...)
}

var SetupOkReader = schema.Struct[SetupOk](SetupOkType(), nil, types.Converters...)

// Setup is a capability that fetches (or idempotently creates) a space's RSA key pair in the KMS.
var Setup = validator.NewCapability(
	SetupAbility,
	schema.DIDString(),
	SetupCaveatsReader,
	func(claimed, delegated ucan.Capability[SetupCaveats]) failure.Failure {
		if fail := utils.EqualWith(claimed, delegated); fail != nil {
			return fail
		}
		if !equalOptionalString(claimed.Nb().Location, delegated.Nb().Location) {
			return schema.NewSchemaError(fmt.Sprintf(
				"Can not derive %s location %v from %v",
				claimed.Can(), claimed.Nb().Location, delegated.Nb().Location,
			))
		}
		if !equalOptionalString(claimed.Nb().Keyring, delegated.Nb().Keyring) {
			return schema.NewSchemaError(fmt.Sprintf(
				"Can not derive %s keyring %v from %v",
				claimed.Can(), claimed.Nb().Keyring, delegated.Nb().Keyring,
			))
		}
		return nil
	},
)

func equalOptionalString(claimed, delegated *string) bool {
	if delegated == nil && claimed == nil {
		return true
	}
	if delegated == nil || claimed == nil {
		return false
	}
	return *claimed == *delegated
}
