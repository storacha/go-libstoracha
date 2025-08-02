package metadata_test

import (
	"testing"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/storacha/go-libstoracha/metadata"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripLocationCommitmentMetadata(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		claim := testutil.RandomCID(t).(cidlink.Link).Cid
		shard := testutil.RandomCID(t).(cidlink.Link).Cid
		length := uint64(138)
		rng := metadata.Range{
			Offset: 10,
			Length: &length,
		}
		meta0 := metadata.LocationCommitmentMetadata{
			Shard:      &shard,
			Range:      &rng,
			Expiration: 1234,
			Claim:      claim,
		}

		bytes, err := meta0.MarshalBinary()
		require.NoError(t, err)

		meta1 := metadata.LocationCommitmentMetadata{}
		err = meta1.UnmarshalBinary(bytes)
		require.NoError(t, err)

		require.Equal(t, meta0, meta1)
	})

	t.Run("optional shard", func(t *testing.T) {
		claim := testutil.RandomCID(t).(cidlink.Link).Cid
		length := uint64(138)
		rng := metadata.Range{
			Offset: 10,
			Length: &length,
		}
		meta0 := metadata.LocationCommitmentMetadata{
			Range:      &rng,
			Expiration: 1234,
			Claim:      claim,
		}

		bytes, err := meta0.MarshalBinary()
		require.NoError(t, err)

		meta1 := metadata.LocationCommitmentMetadata{}
		err = meta1.UnmarshalBinary(bytes)
		require.NoError(t, err)

		require.Equal(t, meta0, meta1)
	})

	t.Run("optional range", func(t *testing.T) {
		claim := testutil.RandomCID(t).(cidlink.Link).Cid
		shard := testutil.RandomCID(t).(cidlink.Link).Cid
		meta0 := metadata.LocationCommitmentMetadata{
			Shard:      &shard,
			Expiration: 1234,
			Claim:      claim,
		}

		bytes, err := meta0.MarshalBinary()
		require.NoError(t, err)

		meta1 := metadata.LocationCommitmentMetadata{}
		err = meta1.UnmarshalBinary(bytes)
		require.NoError(t, err)

		require.Equal(t, meta0, meta1)
	})

	t.Run("optional range length", func(t *testing.T) {
		claim := testutil.RandomCID(t).(cidlink.Link).Cid
		shard := testutil.RandomCID(t).(cidlink.Link).Cid
		rng := metadata.Range{Offset: 10}
		meta0 := metadata.LocationCommitmentMetadata{
			Shard:      &shard,
			Range:      &rng,
			Expiration: 1234,
			Claim:      claim,
		}

		bytes, err := meta0.MarshalBinary()
		require.NoError(t, err)

		meta1 := metadata.LocationCommitmentMetadata{}
		err = meta1.UnmarshalBinary(bytes)
		require.NoError(t, err)

		require.Equal(t, meta0, meta1)
	})
}
