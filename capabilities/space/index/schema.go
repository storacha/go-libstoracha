package index

import (
	_ "embed"
	"fmt"

	ipldprime "github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/schema"
)

//go:embed index.ipldsch
var indexSchema []byte

var indexTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := ipldprime.LoadSchemaBytes(indexSchema)
	if err != nil {
		panic(fmt.Errorf("loading index schema: %w", err))
	}
	return ts
}

func AddCaveatsType() schema.Type {
	return indexTS.TypeByName("AddCaveats")
}

func AddOkType() schema.Type {
	return indexTS.TypeByName("AddOk")
}
