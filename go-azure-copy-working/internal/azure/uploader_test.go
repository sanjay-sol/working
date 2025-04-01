package azure

import (
	"context"
	"errors"

	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAzureClient struct {
	mock.Mock
}

func (m *MockAzureClient) UploadBuffer(ctx context.Context, container, blobName string, data []byte) (string, error) {
	args := m.Called(ctx, container, blobName, data)
	return args.String(0), args.Error(1)
}

func TestUploadImage_Success(t *testing.T) {
	mockClient := new(MockAzureClient)
	mockClient.On("UploadBuffer", mock.Anything, "test-container", "test-image.png", mock.Anything).Return("mock-etag", nil)

	imageData := []byte("fake image data")
	etag, err := mockClient.UploadBuffer(context.Background(), "test-container", "test-image.png", imageData)

	assert.NoError(t, err, "Expected no error on successful upload")
	assert.Equal(t, "mock-etag", etag, "Expected ETag to be returned correctly")
	mockClient.AssertExpectations(t)
}

func TestUploadImage_Failure(t *testing.T) {
	mockClient := new(MockAzureClient)
	mockClient.On("UploadBuffer", mock.Anything, "test-container", "test-image.png", mock.Anything).
		Return("", errors.New("upload failed"))

	imageData := []byte("fake image data")
	etag, err := mockClient.UploadBuffer(context.Background(), "test-container", "test-image.png", imageData)

	assert.Error(t, err, "Expected an error on failed upload")
	assert.Equal(t, "", etag, "ETag should be empty on failure")
	mockClient.AssertExpectations(t)
}

func TestUploadImage_EmptyData(t *testing.T) {
	mockClient := new(MockAzureClient)
	mockClient.On("UploadBuffer", mock.Anything, "test-container", "test-image.png", []byte{}).
		Return("", errors.New("empty data"))

	etag, err := mockClient.UploadBuffer(context.Background(), "test-container", "test-image.png", []byte{})

	assert.Error(t, err, "Expected an error on empty data")
	assert.Equal(t, "", etag, "ETag should be empty when data is empty")
	mockClient.AssertExpectations(t)
}
