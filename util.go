package main

import (
	"bytes"
	"io"
	"time"

	"encoding/json"
)

func toTime(str string) (time.Time, error) {
	layout := "2006-01-02T15:04:05Z"
	return time.Parse(layout, str)
}

func toReader(data interface{}) (io.Reader, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
