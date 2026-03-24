package encryption

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed encryption.ipldsch
var encryptionSchema []byte

var encryptionTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(encryptionSchema)
	if err != nil {
		panic(fmt.Errorf("loading space/encryption schema: %w", err))
	}
	return ts
}

func SetupCaveatsType() schema.Type {
	return encryptionTS.TypeByName("SetupCaveats")
}

func SetupOkType() schema.Type {
	return encryptionTS.TypeByName("SetupOk")
}

func KeyDecryptCaveatsType() schema.Type {
	return encryptionTS.TypeByName("KeyDecryptCaveats")
}

func KeyDecryptOkType() schema.Type {
	return encryptionTS.TypeByName("KeyDecryptOk")
}
