package sign

import (
	"fmt"

	"github.com/storacha/go-libstoracha/capabilities/types"
	"github.com/storacha/go-ucanto/core/ipld"
	"github.com/storacha/go-ucanto/ucan"
)

type SignError struct {
	ErrorName string
	Message   string
}

func (se SignError) Name() string {
	return se.ErrorName
}

func (se SignError) Error() string {
	return se.Message
}

func (ge SignError) ToIPLD() (ipld.Node, error) {
	return ipld.WrapWithRecovery(&ge, SignErrorType(), types.Converters...)
}

func NewInvalidResourceError(expected ucan.Resource, actual ucan.Resource) error {
	return SignError{
		ErrorName: "InvalidResource",
		Message:   fmt.Sprintf("invalid resource, expected: %s got: %s", expected, actual),
	}
}
