package ucan

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed ucan.ipldsch
var ucanSchema []byte

var ucanTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(ucanSchema)
	if err != nil {
		panic(fmt.Errorf("loading ucan schema: %w", err))
	}
	return ts
}

func ConcludeCaveatsType() schema.Type {
	return ucanTS.TypeByName("ConcludeCaveats")
}

func ConcludeOkType() schema.Type {
	return ucanTS.TypeByName("ConcludeOk")
}
