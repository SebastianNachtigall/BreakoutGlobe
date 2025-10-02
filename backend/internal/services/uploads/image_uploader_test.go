package uploads

import (
	"bytes"
	"context"
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestImageUploader_UploadPOIImage_Success(t *testing.T) {
	// Create mock storage
	mockStorage := &MockFileStorage{}
	uploader := NewImageUploader(mockStorage)

	// Create a mock multipart file
	fileHeader := createMockImageFile(t, "test-image.jpg", "image/jpeg", []byte("fake-image-data"))

	// Setup mock expectations
	expectedURL := "http://localhost:8080/uploads/poi-test.jpg"
	mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), "image/jpeg").Return(expectedURL, nil)

	// Upload the image
	imageURL, err := uploader.UploadPOIImage(context.Background(), fileHeader)

	// Verify success
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, imageURL)
	mockStorage.AssertExpectations(t)

}

func TestImageUploader_UploadPOIImage_InvalidFileType(t *testing.T) {
	mockStorage := &MockFileStorage{}
	uploader := NewImageUploader(mockStorage)

	// Create a mock file with invalid type
	fileHeader := createMockImageFile(t, "document.pdf", "application/pdf", []byte("fake-pdf-data"))

	// Upload should fail
	_, err := uploader.UploadPOIImage(context.Background(), fileHeader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid image type")
}

func TestImageUploader_UploadPOIImage_FileTooLarge(t *testing.T) {
	mockStorage := &MockFileStorage{}
	uploader := NewImageUploader(mockStorage)

	// Create a mock file that's too large (6MB)
	largeData := make([]byte, 6*1024*1024)
	fileHeader := createMockImageFile(t, "large-image.jpg", "image/jpeg", largeData)

	// Upload should fail
	_, err := uploader.UploadPOIImage(context.Background(), fileHeader)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image file too large")
}

func TestImageUploader_UploadPOIImage_NilFile(t *testing.T) {
	mockStorage := &MockFileStorage{}
	uploader := NewImageUploader(mockStorage)

	// Upload should fail with nil file
	_, err := uploader.UploadPOIImage(context.Background(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image file provided")
}

// Helper function to create a mock multipart file header
func createMockImageFile(t *testing.T, filename, contentType string, data []byte) *multipart.FileHeader {
	// Create a buffer to write our multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Create a form file field
	part, err := writer.CreateFormFile("image", filename)
	assert.NoError(t, err)

	// Write the data
	_, err = part.Write(data)
	assert.NoError(t, err)

	// Close the writer
	err = writer.Close()
	assert.NoError(t, err)

	// Parse the multipart form
	reader := multipart.NewReader(&buf, writer.Boundary())
	form, err := reader.ReadForm(10 << 20) // 10MB max
	assert.NoError(t, err)

	// Get the file header
	fileHeaders := form.File["image"]
	assert.Len(t, fileHeaders, 1)

	fileHeader := fileHeaders[0]
	
	// Set the content type header manually since multipart.NewReader doesn't set it
	if fileHeader.Header == nil {
		fileHeader.Header = make(map[string][]string)
	}
	fileHeader.Header.Set("Content-Type", contentType)

	return fileHeader
}

// MockFileStorage provides a mock implementation of FileStorage for testing
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) UploadFile(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	args := m.Called(ctx, key, data, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) DeleteFile(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockFileStorage) GetFileURL(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockFileStorage) FileExists(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockFileStorage) GenerateUniqueKey(prefix, userID, originalFilename string) string {
	args := m.Called(prefix, userID, originalFilename)
	return args.String(0)
}