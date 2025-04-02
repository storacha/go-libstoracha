package http

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed http.ipldsch
var httpSchema []byte

var httpTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(httpSchema)
	if err != nil {
		panic(fmt.Errorf("loading http schema: %w", err))
	}
	return ts
}

func PutCaveatsType() schema.Type {
	return httpTS.TypeByName("PutCaveats")
}

func PutOkType() schema.Type {
	return httpTS.TypeByName("PutOk")
}
