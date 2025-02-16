package main

import (
	"api-usage/pkg/kong"
	"api-usage/pkg/logger"
)

func main() {
	// cmd.Execute()

	log := kong.Log{
		Request: kong.Request{
			URI: "/some/uri",
			Headers: map[string]string{
				"User-Agent":    "curl/7.64.1",
				"Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			},
			Method: "GET",
			Size:   100,
			URL:    "http://example.com/some/uri",
		},
		Response: kong.Response{
			Status: 200,
			Size:   200,
		},
		Latencies: kong.Latencies{
			Request: 100,
		},
		ClientIP: "192.10.2.3",
	}

	lgr, err := logger.New([]logger.Field{
		{
			Name:     "path",
			Property: "Request.URI",
		},
		{
			Name:     "duration",
			Property: "Latencies.Request",
		},
		{
			Name:        "user_id",
			Property:    `Request.Headers["Authorization"]`,
			Transformer: "jwt",
			With: map[string]string{
				"key":  "payload.sub",
				"type": "number",
			},
		},
		{
			Name:        "host",
			Property:    `Request.URL`,
			Transformer: "url",
			With: map[string]string{
				"key": "host",
			},
		},
		{
			Name:     "request_size",
			Property: `Request.Size`,
		},
		{
			Name:     "response_size",
			Property: `Response.Size`,
		},
		{
			Name:     "client_ip",
			Property: `ClientIP`,
		},
		{
			Name:     "user_agent",
			Property: `Request.Headers["User-Agent"]`,
		},
		{
			Name:     "status_code",
			Property: `Response.Status`,
		},
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10000; i++ {
		lgr.Log(log)
	}

}
