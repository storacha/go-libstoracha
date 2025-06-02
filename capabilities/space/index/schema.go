package index

import (
	_ "embed"
	"fmt"

	ipldprime "github.com/ipld/go-ipld-prime"
	ipldschema "github.com/ipld/go-ipld-prime/schema"
)

//go:embed index.ipldsch
var assertSchema []byte

var assertTS = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := ipldprime.LoadSchemaBytes(assertSchema)
	if err != nil {
		panic(fmt.Errorf("loading assert schema: %w", err))
	}
	return ts
}

func AddCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("AddCaveats")
}
