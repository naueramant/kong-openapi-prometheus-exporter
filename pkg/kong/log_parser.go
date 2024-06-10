package kong

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

type Log struct {
	Request   Request   `json:"request"`
	Response  Response  `json:"response"`
	Latencies Latencies `json:"latencies"`
}

type Latencies struct {
	Request int `json:"request"`
}

type Request struct {
	URI     string            `json:"uri"`
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
}

type Response struct {
	Status int `json:"status"`
}

func ParseLog(body io.Reader) (*Log, error) {
	var log Log

	err := jsoniter.NewDecoder(body).Decode(&log)
	if err != nil {
		return nil, err
	}

	return &log, nil
}
