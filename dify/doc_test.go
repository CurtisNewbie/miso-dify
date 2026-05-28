package dify

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/curtisnewbie/miso/miso"
)

func TestUploadDocumentSendsDuplicateFalse(t *testing.T) {
	var capturedData string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		capturedData = r.FormValue("data")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"document":{"id":"new-uuid-123","tokens":0,"word_count":0},"batch":"20240101"}`)
	}))
	defer srv.Close()

	f, err := os.CreateTemp("", "dify-test-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(f.Name())
	f.WriteString("hello")
	f.Close()

	rail := miso.EmptyRail()
	_, err = UploadDocument(rail, srv.URL, "test-key", UploadDocumentReq{
		DatasetId: "ds-1",
		FilePath:  f.Name(),
		Filename:  "test.txt",
	})
	if err != nil {
		t.Fatalf("UploadDocument: %v", err)
	}

	var apiReq UploadDocumentApiReq
	if err := json.Unmarshal([]byte(capturedData), &apiReq); err != nil {
		t.Fatalf("unmarshal data field: %v", err)
	}
	if apiReq.Duplicate {
		t.Errorf("expected Duplicate=false, got true")
	}
}
