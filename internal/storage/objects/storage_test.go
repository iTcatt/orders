package objects_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"iTcatt/orders/internal/storage/objects"
)

const (
	minioAccessKey = "minioadmin"
	minioSecretKey = "minioadmin"
	testBucket     = "test-bucket"
)

var (
	testStore  *objects.Store
	testClient *minio.Client
	publicURL  string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "minio/minio:latest",
			ExposedPorts: []string{"9000/tcp"},
			Env: map[string]string{
				"MINIO_ROOT_USER":     minioAccessKey,
				"MINIO_ROOT_PASSWORD": minioSecretKey,
			},
			Cmd:        []string{"server", "/data"},
			WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000"),
		},
		Started: true,
	})
	if err != nil {
		panic("failed to start minio container: " + err.Error())
	}
	defer container.Terminate(ctx) //nolint:errcheck

	host, err := container.Host(ctx)
	if err != nil {
		panic("failed to get host: " + err.Error())
	}
	port, err := container.MappedPort(ctx, "9000")
	if err != nil {
		panic("failed to get port: " + err.Error())
	}

	endpoint := fmt.Sprintf("%s:%s", host, port.Port())
	publicURL = fmt.Sprintf("http://%s", endpoint)

	testClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		panic("failed to create minio client: " + err.Error())
	}

	testStore = objects.New(testClient, testBucket, publicURL)

	if err := testStore.EnsureBucket(ctx); err != nil {
		panic("failed to ensure bucket: " + err.Error())
	}

	m.Run()
}

func cleanBucket(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	for obj := range testClient.ListObjects(ctx, testBucket, minio.ListObjectsOptions{Recursive: true}) {
		require.NoError(t, obj.Err)
		require.NoError(t, testClient.RemoveObject(ctx, testBucket, obj.Key, minio.RemoveObjectOptions{}))
	}
}

func TestStore_URL(t *testing.T) {
	store := objects.New(testClient, "my-bucket", "http://example.com")
	url := store.URL("products/123/image.jpg")
	assert.Equal(t, "http://example.com/my-bucket/products/123/image.jpg", url)
}

func TestStore_EnsureBucket(t *testing.T) {
	ctx := context.Background()

	t.Run("creates bucket when it does not exist", func(t *testing.T) {
		bucket := "temp-create-test"
		store := objects.New(testClient, bucket, publicURL)
		defer testClient.RemoveBucket(ctx, bucket) //nolint:errcheck

		err := store.EnsureBucket(ctx)
		require.NoError(t, err)

		exists, err := testClient.BucketExists(ctx, bucket)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("succeeds when bucket already exists", func(t *testing.T) {
		err := testStore.EnsureBucket(ctx)
		require.NoError(t, err)
	})
}

func TestStore_Upload(t *testing.T) {
	ctx := context.Background()

	t.Run("stores object with correct content", func(t *testing.T) {
		cleanBucket(t)
		content := []byte("hello, world")
		key := "test/hello.txt"

		err := testStore.Upload(ctx, key, bytes.NewReader(content), int64(len(content)), "text/plain")
		require.NoError(t, err)

		obj, err := testClient.GetObject(ctx, testBucket, key, minio.GetObjectOptions{})
		require.NoError(t, err)
		defer obj.Close() //nolint:errcheck

		got, err := io.ReadAll(obj)
		require.NoError(t, err)
		assert.Equal(t, content, got)
	})

	t.Run("stores object with correct content type", func(t *testing.T) {
		cleanBucket(t)
		content := []byte{0xFF, 0xD8, 0xFF, 0xE0}
		key := "test/image.jpg"

		err := testStore.Upload(ctx, key, bytes.NewReader(content), int64(len(content)), "image/jpeg")
		require.NoError(t, err)

		info, err := testClient.StatObject(ctx, testBucket, key, minio.StatObjectOptions{})
		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", info.ContentType)
	})
}

func TestStore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes existing object", func(t *testing.T) {
		cleanBucket(t)
		content := []byte("to be deleted")
		key := "test/delete-me.txt"

		_, err := testClient.PutObject(ctx, testBucket, key,
			bytes.NewReader(content), int64(len(content)),
			minio.PutObjectOptions{ContentType: "text/plain"},
		)
		require.NoError(t, err)

		err = testStore.Delete(ctx, key)
		require.NoError(t, err)

		_, err = testClient.StatObject(ctx, testBucket, key, minio.StatObjectOptions{})
		require.Error(t, err)
	})

	t.Run("succeeds for non-existent object", func(t *testing.T) {
		cleanBucket(t)
		err := testStore.Delete(ctx, "test/does-not-exist.txt")
		require.NoError(t, err)
	})
}
