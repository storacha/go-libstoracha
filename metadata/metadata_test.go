package metadata_test

import (
	"testing"

	"github.com/storacha/go-libstoracha/metadata"
	"github.com/storacha/go-libstoracha/testutil"
	"github.com/stretchr/testify/require"
)

func TestRoundTripLocationCommitmentMetadata(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		claim := testutil.RandomCID(t).Cid
		shard := testutil.RandomCID(t).Cid
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
		claim := testutil.RandomCID(t).Cid
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
		claim := testutil.RandomCID(t).Cid
		shard := testutil.RandomCID(t).Cid
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
		claim := testutil.RandomCID(t).Cid
		shard := testutil.RandomCID(t).Cid
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
