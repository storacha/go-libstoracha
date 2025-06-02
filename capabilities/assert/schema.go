package assert

import (
	// for schema embed
	_ "embed"
	"fmt"

	ipldschema "github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed assert.ipldsch
var assertSchema []byte

var assertTS = mustLoadTS()

func mustLoadTS() *ipldschema.TypeSystem {
	ts, err := types.LoadSchemaBytes(assertSchema)
	if err != nil {
		panic(fmt.Errorf("loading assert schema: %w", err))
	}
	return ts
}

func LocationCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("LocationCaveats")
}

func InclusionCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("InclusionCaveats")
}

func IndexCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("IndexCaveats")
}

func PartitionCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("PartitionCaveats")
}

func RelationCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("RelationCaveats")
}

func EqualsCaveatsType() ipldschema.Type {
	return assertTS.TypeByName("EqualsCaveats")
}
