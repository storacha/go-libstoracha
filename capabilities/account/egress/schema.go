package egress

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	captypes "github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed egress.ipldsch
var egressSchema []byte

var egressTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := captypes.LoadSchemaBytes(egressSchema)
	if err != nil {
		panic(fmt.Errorf("loading egress schema: %w", err))
	}
	return ts
}

func GetCaveatsType() schema.Type {
	return egressTS.TypeByName("GetCaveats")
}

func GetOkType() schema.Type {
	return egressTS.TypeByName("GetOk")
}

func GetErrorType() schema.Type {
	return egressTS.TypeByName("GetError")
}
