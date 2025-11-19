package infra

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestRecieveFile(t *testing.T) {
	bodyString := `--xxx
Content-Disposition: form-data; name="logo"; filename="text.txt"
Content-Type: application/octet-stream

Some test string
--xxx--
`
	type expResult struct {
		fileBody string
		err      error
	}

	tests := []struct {
		name              string
		reqHeader         http.Header
		formFileKey       string
		validContentTypes []string
		expResult         expResult
	}{
		{
			name:        "empty_form_file_key",
			formFileKey: "",
			expResult: expResult{
				err: newRecieveFileError(
					http.StatusInternalServerError,
					"empty FormFileKey",
				),
			},
		},
		{
			name:        "failed_parsing_multipart",
			formFileKey: "some key",
			expResult: expResult{
				err: newRecieveFileError(
					http.StatusInternalServerError,
					"failed parsing multipart form: ",
				),
			},
		},
		{
			name:        "failed_parsing_file",
			reqHeader:   http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
			formFileKey: "invalid key",
			expResult: expResult{
				err: newRecieveFileError(
					http.StatusInternalServerError,
					"failed parsing file: ",
				),
			},
		},
		{
			name:              "invalid_content_type",
			reqHeader:         http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
			formFileKey:       "logo",
			validContentTypes: []string{},
			expResult: expResult{
				err: newRecieveFileError(
					http.StatusBadRequest,
					"invalid Content-Type: ",
				),
			},
		},
		{
			name:              "success",
			reqHeader:         http.Header{"Content-Type": {`multipart/form-data; boundary=xxx`}},
			formFileKey:       "logo",
			validContentTypes: []string{"text/plain; charset=utf-8"},
			expResult: expResult{
				fileBody: "Some test string",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := &http.Request{
				Method: "POST",
				Header: test.reqHeader,
				Body:   io.NopCloser(strings.NewReader(bodyString)),
			}

			actFile, actErr := RecieveFile(req, RecieveFileOptions{
				FormFileKey:       test.formFileKey,
				ValidContentTypes: test.validContentTypes,
			})

			if actFile != nil {
				actFileBody, _ := io.ReadAll(actFile)
				if got, want := string(actFileBody), test.expResult.fileBody; got != want {
					t.Errorf("Wrong! Unexpected file body!\n\tExpected: %s\n\tActual: %s", want, got)
				}
			}

			if got, want := actErr, test.expResult.err; !errors.Is(got, want) {
				t.Errorf("Wrong! Unexpected error!\n\tExpected: %v\n\tActualt: %v", want, got)
			}
		})
	}
}
