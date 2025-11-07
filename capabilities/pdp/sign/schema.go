package sign

import (
	// for go:embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed sign.ipldsch
var signSchema []byte

var ts = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(signSchema)
	if err != nil {
		panic(fmt.Errorf("loading blob schema: %w", err))
	}
	return ts
}

func AuthSignatureType() schema.Type {
	return ts.TypeByName("AuthSignature")
}

func DataSetCreateCaveatsType() schema.Type {
	return ts.TypeByName("DataSetCreateCaveats")
}

func PiecesAddCaveatsType() schema.Type {
	return ts.TypeByName("PiecesAddCaveats")
}

func PiecesRemoveScheduleCaveatsType() schema.Type {
	return ts.TypeByName("PiecesRemoveScheduleCaveats")
}

func DataSetDeleteCaveatsType() schema.Type {
	return ts.TypeByName("DataSetDeleteCaveats")
}

func SignErrorType() schema.Type {
	return ts.TypeByName("InvalidResource")
}
