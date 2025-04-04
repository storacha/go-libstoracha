package types

import (
	"bytes"
	"io"
	"os"

	"github.com/ipld/go-ipld-prime/schema"
	schemadmt "github.com/ipld/go-ipld-prime/schema/dmt"
	schemadsl "github.com/ipld/go-ipld-prime/schema/dsl"
)

// LoadSchemaBytes is a shortcut for LoadSchema for the common case where
// the schema is available as a buffer or a string, such as via go:embed.
func LoadSchemaBytes(src []byte) (*schema.TypeSystem, error) {
	return LoadSchema("", bytes.NewReader(src))
}

// LoadSchemaBytes is a shortcut for LoadSchema for the common case where
// the schema is a file on disk.
func LoadSchemaFile(path string) (*schema.TypeSystem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadSchema(path, f)
}

// LoadSchema parses an IPLD Schema in its DSL form
// and compiles its types into a standalone TypeSystem.
func LoadSchema(name string, r io.Reader) (*schema.TypeSystem, error) {
	sch, err := schemadsl.Parse(name, r)
	if err != nil {
		return nil, err
	}
	ts := new(schema.TypeSystem)
	ts.Init()
	schema.SpawnDefaultBasicTypes(ts)
	baseSch, err := schemadsl.Parse("", bytes.NewReader(typesSchema))
	if err != nil {
		return nil, err
	}
	err = schemadmt.SpawnSchemaTypes(ts, baseSch)
	if err != nil {
		return nil, err
	}

	// Add the schema types
	err = schemadmt.SpawnSchemaTypes(ts, sch)
	if err != nil {
		return nil, err
	}
	// TODO: if this fails and the user forgot to check Compile's returned error,
	// we can leave the TypeSystem in an unfortunate broken state:
	// they can obtain types out of the TypeSystem and they are non-nil,
	// but trying to use them in any way may result in panics.
	// Consider making that less prone to misuse, such as making it illegal to
	// call TypeByName until ValidateGraph is happy.
	if errs := ts.ValidateGraph(); errs != nil {
		// Return the first error.
		for _, err := range errs {
			return nil, err
		}
	}

	return ts, nil
}
