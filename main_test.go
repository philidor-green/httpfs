package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	filePath := "file.jpg"
	fieldName := "file"
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	file, err := os.Open(filePath)
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
	handler := processFile()

	handler.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Errorf("Expected %d, received %d", 200, res.Code)
	}
}
