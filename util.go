package main

import (
	"bytes"
	"io"
	"time"

	"encoding/json"

	"github.com/Sirupsen/logrus"
)

func toTime(str string) (time.Time, error) {
	layout := "2006-01-02T15:04:05Z"
	return time.Parse(layout, str)
}

func toReader(data interface{}) (io.Reader, error) {
	b, err := json.Marshal(data)
	if err != nil {
		logrus.Error("error marshalling data")
		return nil, err
	}
	return bytes.NewReader(b), nil
}
