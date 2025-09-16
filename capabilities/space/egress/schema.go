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
		panic(fmt.Errorf("loading index schema: %w", err))
	}
	return ts
}

func TrackCaveatsType() schema.Type {
	return egressTS.TypeByName("TrackCaveats")
}

func TrackOkType() schema.Type {
	return egressTS.TypeByName("TrackOk")
}

func TrackErrorType() schema.Type {
	return egressTS.TypeByName("TrackError")
}

func ConsolidateCaveatsType() schema.Type {
	return egressTS.TypeByName("ConsolidateCaveats")
}

func ConsolidateOkType() schema.Type {
	return egressTS.TypeByName("ConsolidateOk")
}

func ConsolidateErrorType() schema.Type {
	return egressTS.TypeByName("ConsolidateError")
}
