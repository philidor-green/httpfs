package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestMain(t *testing.T) {
	filePath := "file.jpg"
	fieldName := "file"
	var AppFs = afero.NewMemMapFs()

	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	afero.WriteFile(AppFs, filePath, []byte("hello world"), 0644)
	file, err := AppFs.Create(filePath)
	if err != nil {
		t.Fatal(err)
	}
	w, err := mw.CreateFormFile(fieldName, filePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(w, file); err != nil {
		t.Fatal(err)
	}
	// close the writer before making the request
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	res := httptest.NewRecorder()
	handler := processFile(AppFs)

	handler.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Errorf("Expected %d, received %d", 200, res.Code)
	}
	fileName := "downloaded"
	_, err = AppFs.Stat(fileName)
	if os.IsNotExist(err) {
		t.Errorf("file \"%s\" does not exist.\n", fileName)
	}
}
