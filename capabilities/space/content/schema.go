package content

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed content.ipldsch
var contentSchema []byte

var contentTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(contentSchema)
	if err != nil {
		panic(fmt.Errorf("loading content schema: %w", err))
	}
	return ts
}

func RetrieveCaveatsType() schema.Type {
	return contentTS.TypeByName("RetrieveCaveats")
}

func RetrieveOkType() schema.Type {
	return contentTS.TypeByName("RetrieveOk")
}

func RangeNotSatisfiableErrorType() schema.Type {
	return contentTS.TypeByName("RangeNotSatisfiableError")
}
