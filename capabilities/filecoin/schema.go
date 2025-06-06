package filecoin

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed filecoin.ipldsch
var filecoinSchema []byte

var filecoinTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(filecoinSchema)
	if err != nil {
		panic(fmt.Errorf("loading filecoin schema: %w", err))
	}
	return ts
}

func OfferCaveatsType() schema.Type {
	return filecoinTS.TypeByName("OfferCaveats")
}

func OfferOkType() schema.Type {
	return filecoinTS.TypeByName("OfferOk")
}

func SubmitCaveatsType() schema.Type {
	return filecoinTS.TypeByName("SubmitCaveats")
}

func SubmitOkType() schema.Type {
	return filecoinTS.TypeByName("SubmitOk")
}

func AcceptCaveatsType() schema.Type {
	return filecoinTS.TypeByName("AcceptCaveats")
}

func AcceptOkType() schema.Type {
	return filecoinTS.TypeByName("AcceptOk")
}

func InfoCaveatsType() schema.Type {
	return filecoinTS.TypeByName("InfoCaveats")
}

func InfoOkType() schema.Type {
	return filecoinTS.TypeByName("InfoOk")
}

func InfoAcceptedAggregateType() schema.Type {
	return filecoinTS.TypeByName("InfoAcceptedAggregate")
}

func InfoAcceptedDealType() schema.Type {
	return filecoinTS.TypeByName("InfoAcceptedDeal")
}

func InclusionProofType() schema.Type {
	return filecoinTS.TypeByName("InclusionProof")
}

func ProofDataType() schema.Type {
	return filecoinTS.TypeByName("ProofData")
}

func DealMetadataType() schema.Type {
	return filecoinTS.TypeByName("DealMetadata")
}

func SingletonMarketSourceType() schema.Type {
	return filecoinTS.TypeByName("SingletonMarketSource")
}
