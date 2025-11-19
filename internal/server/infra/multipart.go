package infra

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

type RecieveFileError struct {
	Code int
	Err  error
}

func (me *RecieveFileError) Error() string {
	return me.Err.Error()
}

func (me *RecieveFileError) Is(target error) bool {
	var targetRFE *RecieveFileError

	if errors.As(target, &targetRFE) {
		return me.Code == targetRFE.Code &&
			strings.HasPrefix(me.Error(), targetRFE.Error())
	}

	return false
}

func newRecieveFileError(code int, msg string) *RecieveFileError {
	return &RecieveFileError{Code: code, Err: errors.New(msg)}
}

type RecieveFileOptions struct {
	MaxMemory         int64
	FormFileKey       string
	ValidContentTypes []string
}

func RecieveFile(r *http.Request, opts RecieveFileOptions) (io.ReadCloser, error) {
	if len(opts.FormFileKey) == 0 {
		errMsg := "empty FormFileKey"
		return nil, newRecieveFileError(http.StatusInternalServerError, errMsg)
	}
	if opts.MaxMemory == 0 {
		opts.MaxMemory = 4 << 20
	}

	if err := r.ParseMultipartForm(opts.MaxMemory); err != nil {
		errMsg := fmt.Sprintf("failed parsing multipart form: %v", err)
		return nil, newRecieveFileError(http.StatusInternalServerError, errMsg)
	}

	file, _, err := r.FormFile(opts.FormFileKey)
	if err != nil {
		errMsg := fmt.Sprintf("failed parsing file: %v", err)
		return nil, newRecieveFileError(http.StatusInternalServerError, errMsg)
	}

	if opts.ValidContentTypes != nil {
		file.Close()
		mtype, recycled, err := recycleReader(file)
		if err != nil {
			errMsg := fmt.Sprintf("failed detecting mimetype: %v", err)
			return nil, newRecieveFileError(http.StatusInternalServerError, errMsg)
		}

		if !slices.Contains(opts.ValidContentTypes, mtype) {
			errMsg := fmt.Sprintf("invalid Content-Type: expect %v, actual: %v", opts.ValidContentTypes, mtype)
			return nil, newRecieveFileError(http.StatusBadRequest, errMsg)
		}

		return io.NopCloser(recycled), nil
	}

	return file, nil
}

func recycleReader(input io.Reader) (mimeType string, recycled io.Reader, err error) {
	header := bytes.NewBuffer(nil)

	mtype, err := mimetype.DetectReader(io.TeeReader(input, header))
	if err != nil {
		return mimeType, recycled, err
	}

	recycled = io.MultiReader(header, input)

	return mtype.String(), recycled, err
}
