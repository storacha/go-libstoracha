package blob

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"

	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed blob.ipldsch
var blobSchema []byte

var blobTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(blobSchema)
	if err != nil {
		panic(fmt.Errorf("loading space/blob schema: %w", err))
	}
	return ts
}

func AddCaveatsType() schema.Type {
	return blobTS.TypeByName("AddCaveats")
}

func AddOkType() schema.Type {
	return blobTS.TypeByName("AddOk")
}

func AddErrorType() schema.Type {
	return blobTS.TypeByName("AddError")
}

func RemoveCaveatsType() schema.Type {
	return blobTS.TypeByName("RemoveCaveats")
}

func RemoveOkType() schema.Type {
	return blobTS.TypeByName("RemoveOk")
}

func ListCaveatsType() schema.Type {
	return blobTS.TypeByName("ListCaveats")
}

func ListOkType() schema.Type {
	return blobTS.TypeByName("ListOk")
}

func GetCaveatsType() schema.Type {
	return blobTS.TypeByName("GetCaveats")
}

func GetOkType() schema.Type {
	return blobTS.TypeByName("GetOk")
}

func ReplicateCaveatsType() schema.Type {
	return blobTS.TypeByName("ReplicateCaveats")
}

func ReplicateOkType() schema.Type {
	return blobTS.TypeByName("ReplicateOk")
}
