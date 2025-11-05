package sign

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/schema/options"
)

var HashConverter = options.NamedBytesConverter("Hash",
	func(b []byte) (common.Hash, error) {
		return common.BytesToHash(b), nil
	},
	func(h common.Hash) ([]byte, error) {
		return h.Bytes(), nil
	})

var AddressConverter = options.NamedBytesConverter("Address",
	func(b []byte) (common.Address, error) {
		return common.BytesToAddress(b), nil
	},
	func(a common.Address) ([]byte, error) {
		return a.Bytes(), nil
	})

var Converters = append([]bindnode.Option{
	HashConverter,
	AddressConverter,
}, types.Converters...)
