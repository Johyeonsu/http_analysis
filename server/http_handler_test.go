package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerFunc(t *testing.T) {
	assert := assert.New(t)

	t.Run("uploadHandler test", func(t *testing.T) {
		path := "./public/video.mp4"
		file, _ := os.Open(path)

		defer file.Close()
		buf := &bytes.Buffer{}
		writer := multipart.NewWriter(buf)
		multi, err := writer.CreateFormFile("upload_file", filepath.Base(path))
		assert.NoError(err)
		io.Copy(multi, file)
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", buf)
		res := httptest.NewRecorder()

		req.Header.Set("Content-type", writer.FormDataContentType())

		uploadHandler(res, req)
		assert.Equal(http.StatusOK, res.Code)

		result := fileExist(uploadPath + defaultFile)
		assert.Equal(true, result)

		result = fileCompare(uploadPath+defaultFile, path)
		assert.Equal(true, result)
	})

	t.Run("downloadHandler test", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/"+defaultFile, nil)
		res := httptest.NewRecorder()

		downloadHandler(res, req)
		assert.Equal(http.StatusOK, res.Code)

		result := fileExist(downloadPath + defaultFile)
		assert.Equal(true, result)
	})
}

func TestFileFunc(t *testing.T) {
	assert := assert.New(t)

	t.Run("Test file exist", func(t *testing.T) {
		result := fileExist(defaultFilePath)
		assert.Equal(true, result)
	})

	t.Run("Test file not exist", func(t *testing.T) {
		result := fileExist(defaultPath + "abc.mp4")
		assert.Equal(false, result)
	})

	t.Run("Test same file compare", func(t *testing.T) {
		result := fileCompare(indexFilePath, indexFilePath)
		assert.Equal(true, result)
	})

	t.Run("Test diff file compare", func(t *testing.T) {
		result := fileCompare(indexFilePath, defaultFilePath)
		assert.Equal(false, result)
	})
}
