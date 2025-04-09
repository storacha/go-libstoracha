package upload

import (
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed upload.ipldsch
var uploadSchema []byte

var uploadTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(uploadSchema)
	if err != nil {
		panic(fmt.Errorf("loading upload schema: %w", err))
	}
	return ts
}

func AddCaveatsType() schema.Type {
	return uploadTS.TypeByName("AddCaveats")
}

func AddOkType() schema.Type {
	return uploadTS.TypeByName("AddOk")
}

func GetCaveatsType() schema.Type {
	return uploadTS.TypeByName("GetCaveats")
}

func GetOkType() schema.Type {
	return uploadTS.TypeByName("GetOk")
}

func RemoveCaveatsType() schema.Type {
	return uploadTS.TypeByName("RemoveCaveats")
}

func RemoveOkType() schema.Type {
	return uploadTS.TypeByName("RemoveOk")
}

func ListCaveatsType() schema.Type {
	return uploadTS.TypeByName("ListCaveats")
}

func ListItemType() schema.Type {
	return uploadTS.TypeByName("ListItem")
}

func ListOkType() schema.Type {
	return uploadTS.TypeByName("ListOk")
}