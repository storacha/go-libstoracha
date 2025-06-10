package s3_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/storacha/go-libstoracha/datastore/s3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testAccessKey = "minioadmin"
	testSecretKey = "minioadmin"
	testBucket    = "test-bucket"
	testRootDir   = "test-data"
)

// setupMinIOContainer starts a MinIO container for testing
func setupMinIOContainer(ctx context.Context, t *testing.T) (string, func()) {
	req := testcontainers.ContainerRequest{
		Image:        "minio/minio:latest",
		ExposedPorts: []string{"9000/tcp"},
		Env: map[string]string{
			"MINIO_ACCESS_KEY": testAccessKey,
			"MINIO_SECRET_KEY": testSecretKey,
		},
		Cmd:        []string{"server", "/data"},
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "9000")
	require.NoError(t, err)

	hostIP, err := container.Host(ctx)
	require.NoError(t, err)

	endpoint := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	cleanup := func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return endpoint, cleanup
}

// createTestBucket creates a test bucket in the MinIO instance
func createTestBucket(ctx context.Context, t *testing.T, endpoint string) {
	cfg := aws.NewConfig().
		WithCredentials(credentials.NewStaticCredentials(testAccessKey, testSecretKey, "")).
		WithEndpoint(endpoint).
		WithRegion("us-east-1").
		WithS3ForcePathStyle(true)

	sess, err := session.NewSession(cfg)
	require.NoError(t, err)

	s3Client := awsS3.New(sess)
	_, err = s3Client.CreateBucketWithContext(ctx, &awsS3.CreateBucketInput{
		Bucket: aws.String(testBucket),
	})
	require.NoError(t, err)
}

// newTestS3Datastore creates a new S3 datastore configured for testing
func newTestS3Datastore(t *testing.T, endpoint string) *s3.S3Bucket {
	config := s3.Config{
		AccessKey:      testAccessKey,
		SecretKey:      testSecretKey,
		Bucket:         testBucket,
		Region:         "us-east-1",
		RegionEndpoint: endpoint,
		ForcePathStyle: true,
		RootDirectory:  testRootDir,
		Workers:        10,
	}

	store, err := s3.NewS3Datastore(config)
	require.NoError(t, err)

	return store
}

func TestS3Datastore_Put(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	key := ds.NewKey("/test/key")
	value := []byte("test value")

	err := store.Put(ctx, key, value)
	require.NoError(t, err)

	// Verify the data was stored
	retrieved, err := store.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrieved)
}

func TestS3Datastore_Get(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	key := ds.NewKey("/test/get-key")
	value := []byte("get test value")

	// Put data first
	err := store.Put(ctx, key, value)
	require.NoError(t, err)

	// Test successful get
	retrieved, err := store.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrieved)

	// Test get non-existent key
	nonExistentKey := ds.NewKey("/non/existent")
	_, err = store.Get(ctx, nonExistentKey)
	require.Equal(t, ds.ErrNotFound, err)
}

func TestS3Datastore_Has(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	key := ds.NewKey("/test/has-key")
	value := []byte("has test value")

	// Test has with non-existent key
	exists, err := store.Has(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)

	// Put data
	err = store.Put(ctx, key, value)
	require.NoError(t, err)

	// Test has with existing key
	exists, err = store.Has(ctx, key)
	require.NoError(t, err)
	require.True(t, exists)
}

func TestS3Datastore_GetSize(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	key := ds.NewKey("/test/size-key")
	value := []byte("size test value")

	// Test get size with non-existent key
	_, err := store.GetSize(ctx, key)
	require.Equal(t, ds.ErrNotFound, err)

	// Put data
	err = store.Put(ctx, key, value)
	require.NoError(t, err)

	// Test get size with existing key
	size, err := store.GetSize(ctx, key)
	require.NoError(t, err)
	require.Equal(t, len(value), size)
}

func TestS3Datastore_Delete(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	key := ds.NewKey("/test/delete-key")
	value := []byte("delete test value")

	// Put data first
	err := store.Put(ctx, key, value)
	require.NoError(t, err)

	// Verify it exists
	exists, err := store.Has(ctx, key)
	require.NoError(t, err)
	require.True(t, exists)

	// Delete the key
	err = store.Delete(ctx, key)
	require.NoError(t, err)

	// Verify it no longer exists
	exists, err = store.Has(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)

	// Test delete non-existent key (should be idempotent)
	err = store.Delete(ctx, ds.NewKey("/non/existent"))
	require.NoError(t, err)
}

func TestS3Datastore_Query(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	// Put test data - using simpler key structure for debugging
	testData := map[string][]byte{
		"/users/alice": []byte("alice data"),
		"/users/bob":   []byte("bob data"),
		"/posts/jan":   []byte("january posts"),
		"/posts/feb":   []byte("february posts"),
	}

	for keyStr, value := range testData {
		err := store.Put(ctx, ds.NewKey(keyStr), value)
		require.NoError(t, err)
	}

	t.Run("query with prefix", func(t *testing.T) {
		q := dsq.Query{
			Prefix: "/users", // Use full path including leading slash
		}

		results, err := store.Query(ctx, q)
		require.NoError(t, err)

		var entries []dsq.Entry
		for {
			result, ok := results.NextSync()
			if !ok {
				break
			}
			require.NoError(t, result.Error)
			entries = append(entries, result.Entry)
		}
		results.Close()

		// Debug output
		t.Logf("Found %d entries for prefix '/users'", len(entries))
		for i, entry := range entries {
			t.Logf("Entry %d: Key=%s, Size=%d", i, entry.Key, entry.Size)
		}

		require.Len(t, entries, 2)

		// Check that we got the right keys
		foundKeys := make([]string, len(entries))
		for i, entry := range entries {
			foundKeys[i] = entry.Key
		}
		require.Contains(t, foundKeys, "/users/alice")
		require.Contains(t, foundKeys, "/users/bob")
	})

	t.Run("query keys only", func(t *testing.T) {
		q := dsq.Query{
			Prefix:   "/posts",
			KeysOnly: true,
		}

		results, err := store.Query(ctx, q)
		require.NoError(t, err)

		var entries []dsq.Entry
		for {
			result, ok := results.NextSync()
			if !ok {
				break
			}
			require.NoError(t, result.Error)
			entries = append(entries, result.Entry)
		}
		results.Close()

		// Debug output
		t.Logf("Found %d entries for prefix '/posts' (keys only)", len(entries))

		require.Len(t, entries, 2)

		// Check that values are empty when KeysOnly is true
		for _, entry := range entries {
			require.Nil(t, entry.Value)
		}
	})

	t.Run("query with unsupported filters/orders", func(t *testing.T) {
		q := dsq.Query{
			Filters: []dsq.Filter{dsq.FilterValueCompare{Op: dsq.Equal, Value: []byte("test")}},
		}

		_, err := store.Query(ctx, q)
		require.Error(t, err)
		require.Contains(t, err.Error(), "filters or orders are not supported")
	})
}

func TestS3Datastore_Batch(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	// Create a batch
	batch, err := store.Batch(ctx)
	require.NoError(t, err)

	// Add operations to batch
	testData := map[string][]byte{
		"/batch/key1": []byte("batch value 1"),
		"/batch/key2": []byte("batch value 2"),
		"/batch/key3": []byte("batch value 3"),
	}

	for keyStr, value := range testData {
		err = batch.Put(ctx, ds.NewKey(keyStr), value)
		require.NoError(t, err)
	}

	// Add delete operation
	deleteKey := ds.NewKey("/batch/delete-me")
	err = store.Put(ctx, deleteKey, []byte("to be deleted"))
	require.NoError(t, err)

	err = batch.Delete(ctx, deleteKey)
	require.NoError(t, err)

	// Commit batch
	err = batch.Commit(ctx)
	require.NoError(t, err)

	// Verify all put operations succeeded
	for keyStr, expectedValue := range testData {
		value, err := store.Get(ctx, ds.NewKey(keyStr))
		require.NoError(t, err)
		require.Equal(t, expectedValue, value)
	}

	// Verify delete operation succeeded
	exists, err := store.Has(ctx, deleteKey)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestS3Datastore_BatchLarge(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	// Create a batch with many operations
	batch, err := store.Batch(ctx)
	require.NoError(t, err)

	numOps := 50
	for i := 0; i < numOps; i++ {
		key := ds.NewKey(fmt.Sprintf("/batch/large/key-%d", i))
		value := []byte(fmt.Sprintf("large batch value %d", i))
		err = batch.Put(ctx, key, value)
		require.NoError(t, err)
	}

	// Commit batch
	err = batch.Commit(ctx)
	require.NoError(t, err)

	// Verify all operations succeeded
	for i := 0; i < numOps; i++ {
		key := ds.NewKey(fmt.Sprintf("/batch/large/key-%d", i))
		expectedValue := []byte(fmt.Sprintf("large batch value %d", i))
		value, err := store.Get(ctx, key)
		require.NoError(t, err)
		require.Equal(t, expectedValue, value)
	}
}

func TestS3Datastore_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	numRoutines := 10
	numOpsPerRoutine := 10

	// Run concurrent put operations
	done := make(chan bool, numRoutines)
	for r := 0; r < numRoutines; r++ {
		go func(routineID int) {
			defer func() { done <- true }()
			for i := 0; i < numOpsPerRoutine; i++ {
				key := ds.NewKey(fmt.Sprintf("/concurrent/routine-%d/key-%d", routineID, i))
				value := []byte(fmt.Sprintf("routine %d value %d", routineID, i))
				err := store.Put(ctx, key, value)
				require.NoError(t, err)
			}
		}(r)
	}

	// Wait for all routines to complete
	for i := 0; i < numRoutines; i++ {
		<-done
	}

	// Verify all data was stored correctly
	for r := 0; r < numRoutines; r++ {
		for i := 0; i < numOpsPerRoutine; i++ {
			key := ds.NewKey(fmt.Sprintf("/concurrent/routine-%d/key-%d", r, i))
			expectedValue := []byte(fmt.Sprintf("routine %d value %d", r, i))
			value, err := store.Get(ctx, key)
			require.NoError(t, err)
			require.Equal(t, expectedValue, value)
		}
	}
}

func TestS3Datastore_PathTransformation(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	tests := []struct {
		name        string
		key         string
		expectError bool
	}{
		{"simple key", "/simple", false},
		{"nested key", "/path/to/nested/key", false},
		{"key with special chars", "/path/with-special_chars.txt", false},
		{"deep nesting", "/very/deep/nested/path/with/many/levels", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := ds.NewKey(tt.key)
			value := []byte(fmt.Sprintf("test value for %s", tt.key))

			err := store.Put(ctx, key, value)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify retrieval
			retrieved, err := store.Get(ctx, key)
			require.NoError(t, err)
			require.Equal(t, value, retrieved)
		})
	}
}

func TestS3Datastore_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	t.Run("context cancellation", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately

		key := ds.NewKey("/test/cancelled")
		value := []byte("should not be stored")

		err := store.Put(cancelCtx, key, value)
		require.Error(t, err)
	})

	t.Run("context timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// Give context time to expire
		time.Sleep(10 * time.Millisecond)

		key := ds.NewKey("/test/timeout")
		value := []byte("should timeout")

		err := store.Put(timeoutCtx, key, value)
		require.Error(t, err)
	})
}

func TestS3Datastore_Sync(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	// Sync is a no-op for S3, but should not error
	err := store.Sync(ctx, ds.NewKey("/any/prefix"))
	require.NoError(t, err)
}

func TestS3Datastore_Close(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	// Close should not error
	err := store.Close()
	require.NoError(t, err)
}

// Test datastore interfaces implementation
func TestS3Datastore_Interfaces(t *testing.T) {
	ctx := context.Background()
	endpoint, cleanup := setupMinIOContainer(ctx, t)
	defer cleanup()

	createTestBucket(ctx, t, endpoint)
	store := newTestS3Datastore(t, endpoint)

	// Test that S3Bucket implements ds.Datastore interface
	var _ ds.Datastore = store

	// Test that S3Bucket implements ds.Batching interface
	var _ ds.Batching = store

	// Test basic datastore functionality
	key := ds.NewKey("/interface/test")
	value := []byte("interface test value")

	// Test Put
	err := store.Put(ctx, key, value)
	require.NoError(t, err)

	// Test Get
	retrieved, err := store.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, value, retrieved)

	// Test Has
	exists, err := store.Has(ctx, key)
	require.NoError(t, err)
	require.True(t, exists)

	// Test GetSize
	size, err := store.GetSize(ctx, key)
	require.NoError(t, err)
	require.Equal(t, len(value), size)

	// Test Delete
	err = store.Delete(ctx, key)
	require.NoError(t, err)

	// Verify deletion
	exists, err = store.Has(ctx, key)
	require.NoError(t, err)
	require.False(t, exists)
}
