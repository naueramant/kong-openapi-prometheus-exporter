package swagger

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pb33f/libopenapi"
)

func LoadURL(ctx context.Context, url string) (*Specification, error) {
	openAPISpecBytes, err := fetchFromURL(url)
	if err != nil {
		return nil, err
	}

	document, err := libopenapi.NewDocument(openAPISpecBytes)
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

func fetchFromURL(url string) ([]byte, error) {
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

	return body, nil
}
