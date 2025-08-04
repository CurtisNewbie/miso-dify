package dify

import (
	"bytes"
	"io"
	"os"

	"github.com/curtisnewbie/miso/miso"
)

type UploadFileRes struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
}

func UploadFile(rail miso.Rail, host string, apiKey string, user string, file *os.File, filename string) (UploadFileRes, error) {
	url := host + "/v1/files/upload"
	var res UploadFileRes
	err := miso.NewTClient(rail, url).
		Require2xx().
		AddHeader("Authorization", "Bearer "+apiKey).
		PostFormData(map[string]io.Reader{
			"file": miso.NewReaderFile(file, filename),
			"user": bytes.NewReader([]byte(user)),
		}).
		Json(&res)
	if err != nil {
		return res, miso.WrapErrf(err, "dify UploadFile failed")
	}
	rail.Infof("File Uploaded %#v", res)
	return res, nil
}
