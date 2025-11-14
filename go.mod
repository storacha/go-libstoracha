module github.com/storacha/go-libstoracha

go 1.25.3

require (
	github.com/aws/aws-sdk-go v1.55.7
	github.com/aws/aws-sdk-go-v2 v1.39.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.4
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.8
	github.com/ethereum/go-ethereum v1.16.7
	github.com/filecoin-project/go-data-segment v0.0.1
	github.com/filecoin-project/go-fil-commcid v0.2.0
	github.com/filecoin-project/go-fil-commp-hashhash v0.2.0
	github.com/gobwas/glob v0.2.3
	github.com/google/uuid v1.6.0
	github.com/ipfs/go-cid v0.5.0
	github.com/ipfs/go-datastore v0.8.2
	github.com/ipfs/go-log/v2 v2.6.0
	github.com/ipld/go-ipld-prime v0.21.1-0.20240917223228-6148356a4c2e
	github.com/ipni/go-libipni v0.6.18
	github.com/libp2p/go-libp2p v0.41.1
	github.com/multiformats/go-multiaddr v0.16.0
	github.com/multiformats/go-multibase v0.2.0
	github.com/multiformats/go-multicodec v0.9.1
	github.com/multiformats/go-multihash v0.2.3
	github.com/multiformats/go-varint v0.0.7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/storacha/go-ucanto v0.6.5
	github.com/stretchr/testify v1.10.0
	github.com/whyrusleeping/cbor-gen v0.2.0
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.9 // indirect
	github.com/aws/smithy-go v1.23.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect
	github.com/ipfs/bbloom v0.0.4 // indirect
	github.com/ipfs/go-block-format v0.2.0 // indirect
	github.com/ipfs/go-blockservice v0.5.2 // indirect
	github.com/ipfs/go-ipfs-blockstore v1.3.1 // indirect
	github.com/ipfs/go-ipfs-ds-help v1.1.1 // indirect
	github.com/ipfs/go-ipfs-exchange-interface v0.2.1 // indirect
	github.com/ipfs/go-ipfs-util v0.0.3 // indirect
	github.com/ipfs/go-ipld-cbor v0.1.0 // indirect
	github.com/ipfs/go-ipld-format v0.6.0 // indirect
	github.com/ipfs/go-ipld-legacy v0.2.1 // indirect
	github.com/ipfs/go-log v1.0.5 // indirect
	github.com/ipfs/go-merkledag v0.11.0 // indirect
	github.com/ipfs/go-metrics-interface v0.0.1 // indirect
	github.com/ipfs/go-verifcid v0.0.3 // indirect
	github.com/ipld/go-car v0.6.2 // indirect
	github.com/ipld/go-codec-dagpb v1.6.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/libp2p/go-buffer-pool v0.1.0 // indirect
	github.com/libp2p/go-libp2p-pubsub v0.13.1 // indirect
	github.com/libp2p/go-msgio v0.3.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	github.com/mr-tron/base58 v1.2.0 // indirect
	github.com/multiformats/go-base32 v0.1.0 // indirect
	github.com/multiformats/go-base36 v0.2.0 // indirect
	github.com/multiformats/go-multiaddr-fmt v0.1.0 // indirect
	github.com/multiformats/go-multistream v0.6.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/polydawn/refmt v0.89.1-0.20231129105047-37766d95467a // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/ucan-wg/go-ucan v0.0.0-20240916120445-37f52863156c // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/exp v0.0.0-20250218142911-aa4b98e5adaa // indirect
	golang.org/x/sys v0.36.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/blake3 v1.4.0 // indirect
)
