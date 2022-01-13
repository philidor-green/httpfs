Suppose we have a simple file upload http handler looks like this: 

```golang
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const MB = 1 << 20

func main() {
	r := &Router{&mux.Router{}}

	r.MustResponse("POST", "/", processFile())

	r.Run(":8080", "*")
}

func processFile() http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if err := req.ParseMultipartForm(50 * MB); nil != err {
			log.Printf("while parse %s", err)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		defer func() {
			err := req.MultipartForm.RemoveAll()
			if err != nil {
				log.Printf("Cant delete multipart error %s", err)
			}
		}()

		for _, fheaders := range req.MultipartForm.File {
			for _, hdr := range fheaders {
				log.Printf("Income file len: %d", hdr.Size)

				infile, err := hdr.Open()
				if err != nil {
					log.Printf("Handle open error: %v", err)
					res.WriteHeader(http.StatusInternalServerError)
					continue
				}
				defer infile.Close()

				f, err := os.OpenFile("./downloaded", os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					log.Printf("Create Read Input error %v", err)
					res.WriteHeader(http.StatusInternalServerError)
					continue
				}
				defer f.Close()
				io.Copy(f, infile)
			}
		}
		res.Header().Set("Content-Type", "text/html")
		fmt.Fprint(res, "<h2>Success</h2>")
	})
}

type Router struct {
	*mux.Router
}

func (r *Router) MustResponse(meth, path string, h http.HandlerFunc) {
	r.HandleFunc(path, h).Methods(meth)
}

func (r *Router) Run(address, origins string) {
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origins},
		AllowedMethods:   []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "If-None-Match", "Content-Length", "Accept-Encoding", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)
	http.ListenAndServe(address, handler)
}

func vars(req *http.Request) map[string]string {
	return mux.Vars(req)
}
```

The handle above is common way to upload file onto server. The code below is for testing this handler:

```golang
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

```

The problem here is that our test is actually working with real filesystem.
We need a bucnh of things in order to have desired result.

* it needs actual file to upload
* we should check uploaded file saved without errors
* we should have a cleanup procedure in order to delete saved file

There is a way to test our handler without access to real filesystem. 
The [afero](https://github.com/spf13/afero) can help here.