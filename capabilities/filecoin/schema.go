package filecoin

import (
	// for schema embed
	_ "embed"
	"fmt"

	"github.com/ipld/go-ipld-prime/schema"
	"github.com/storacha/go-libstoracha/capabilities/types"
)

//go:embed storefront.ipldsch
var storefrontSchema []byte

var storefrontTS = mustLoadTS()

func mustLoadTS() *schema.TypeSystem {
	ts, err := types.LoadSchemaBytes(storefrontSchema)
	if err != nil {
		panic(fmt.Errorf("loading storefront schema: %w", err))
	}
	return ts
}

func OfferCaveatsType() schema.Type {
	return storefrontTS.TypeByName("OfferCaveats")
}

func OfferOkType() schema.Type {
	return storefrontTS.TypeByName("OfferOk")
}

func SubmitCaveatsType() schema.Type {
	return storefrontTS.TypeByName("SubmitCaveats")
}

func SubmitOkType() schema.Type {
	return storefrontTS.TypeByName("SubmitOk")
}

func AcceptCaveatsType() schema.Type {
	return storefrontTS.TypeByName("AcceptCaveats")
}

func AcceptOkType() schema.Type {
	return storefrontTS.TypeByName("AcceptOk")
}

func InfoCaveatsType() schema.Type {
	return storefrontTS.TypeByName("InfoCaveats")
}

func InfoOkType() schema.Type {
	return storefrontTS.TypeByName("InfoOk")
}

func InfoAcceptedAggregateType() schema.Type {
	return storefrontTS.TypeByName("InfoAcceptedAggregate")
}

func InfoAcceptedDealType() schema.Type {
	return storefrontTS.TypeByName("InfoAcceptedDeal")
}

func InclusionProofType() schema.Type {
	return storefrontTS.TypeByName("InclusionProof")
}

func ProofDataType() schema.Type {
	return storefrontTS.TypeByName("ProofData")
}

func DealMetadataType() schema.Type {
	return storefrontTS.TypeByName("DealMetadata")
}

func SingletonMarketSourceType() schema.Type {
	return storefrontTS.TypeByName("SingletonMarketSource")
}