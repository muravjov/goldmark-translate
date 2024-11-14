package util

import (
	"fmt"
	"io"
	"net/http"

	"moul.io/http2curl"
)

type Non2xxHTTPCodeError struct {
	StatusCode int
	Message    string
	Body       []byte
}

func (e *Non2xxHTTPCodeError) Error() string {
	return e.Message
}

func BailOut(err error) error {
	Error(err)
	return err
}

func CheckStatusCodeIs2XX(resp *http.Response) error {
	if resp.StatusCode/100 != 2 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return BailOut(err)
		}

		command, _ := http2curl.GetCurlCommand(resp.Request)

		err = &Non2xxHTTPCodeError{
			StatusCode: resp.StatusCode,
			Message: fmt.Sprintf(`bad status code: %d,
%+v
%s
request: %s`, resp.StatusCode, resp.Header, string(body), command),
			Body: body,
		}

		return BailOut(err)
	}

	return nil
}
