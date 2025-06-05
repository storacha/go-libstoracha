package access

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed access.ipldsch
var accessSchema []byte

var accessTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(accessSchema)
	if err != nil {
		panic(fmt.Errorf("loading access schema: %w", err))
	}
	return ts
}

func AuthorizeCaveatsType() schema.Type {
	return accessTS.TypeByName("AuthorizeCaveats")
}

func AuthorizeOkType() schema.Type {
	return accessTS.TypeByName("AuthorizeOk")
}

func ConfirmCaveatsType() schema.Type {
	return accessTS.TypeByName("ConfirmCaveats")
}

func ConfirmOkType() schema.Type {
	return accessTS.TypeByName("ConfirmOk")
}
