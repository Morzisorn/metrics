package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"

	"resty.dev/v3"
)

func gzipMiddleware(r *resty.Request) error {
	body := r.Body
	if body == nil {
		return nil
	}

	var jsonBody []byte
	var err error

	switch b := body.(type) {
	case []byte:
		jsonBody = b
	case string:
		jsonBody = []byte(b)
	default:
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write(jsonBody)
	if err != nil {
		return err
	}
	gz.Close()

	r.SetBody(buf.Bytes())
	r.SetHeader("Content-Encoding", "gzip")
	return nil
}
