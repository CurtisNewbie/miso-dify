package dify

import (
	"bytes"
	"io"
	"os"

	"github.com/curtisnewbie/miso/miso"
	"github.com/curtisnewbie/miso/util/errs"
)

type FileInput struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	Url            string `json:"url"`
	UploadFileId   string `json:"upload_file_id"`
}

func NewFileInputById(uploadFileId string) FileInput {
	return FileInput{
		Type:           "document",
		TransferMethod: TransferMethodLocalFile,
		UploadFileId:   uploadFileId,
	}
}

type UploadFileRes struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
}

func (u UploadFileRes) ToFileInput() FileInput {
	return NewFileInputById(u.Id)
}

func UploadFile(rail miso.Rail, host string, apiKey string, user string, file *os.File, filename string) (UploadFileRes, error) {
	url := host + "/v1/files/upload"
	var res UploadFileRes
	err := miso.NewClient(rail, url).
		Require2xx().
		AddAuthBearer(apiKey).
		PostFormData(map[string]io.Reader{
			"file": miso.NewReaderFile(file, filename),
			"user": bytes.NewReader([]byte(user)),
		}).
		Json(&res)
	if err != nil {
		return res, errs.Wrapf(err, "dify UploadFile failed")
	}
	rail.Infof("File Uploaded %#v", res)
	return res, nil
}
