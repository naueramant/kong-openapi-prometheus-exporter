package swagger

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pb33f/libopenapi"
)

func LoadFile(ctx context.Context, path string) (*Specification, error) {
	specBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return newSpecification(ctx, specBytes)
}

func LoadURL(ctx context.Context, url string) (*Specification, error) {
	httpClient := &http.Client{}

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return newSpecification(ctx, body)
}

func newSpecification(ctx context.Context, specBytes []byte) (*Specification, error) {
	document, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, err
	}

	docModel, errors := document.BuildV3Model()
	if len(errors) > 0 {
		var multiErr error
		for _, err := range errors {
			multiErr = fmt.Errorf("%w\n%s", multiErr, err)
		}

		return nil, multiErr
	}

	spec, err := NewSpecification(ctx, docModel)
	if err != nil {
		return nil, err
	}

	return spec, nil
}
