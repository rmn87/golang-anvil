package anvil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type FillPDFPayload struct {
	Data map[string]interface{} `json:"data"`

	// optional
	Title                *string `json:"title,omitempty"`
	FontFamily           *string `json:"fontFamily,omitempty"`
	FontSize             *int    `json:"fontSize,omitempty"`
	TextColor            *string `json:"textColor,omitempty"`
	UseInteractiveFields *bool   `json:"useInteractiveFields,omitempty"`
}

type GeneratePDFPayload struct {
	Data string `json:"data"`

	// optional
	Title            *string `json:"title,omitempty"`
	Type             *string `json:"type,omitempty"`
	FontFamily       *string `json:"fontFamily,omitempty"`
	FontSize         *int    `json:"fontSize,omitempty"`
	TextColor        *string `json:"textColor,omitempty"`
	IncludeTimestamp *bool   `json:"includeTimestamp,omitempty"`
	Logo             *Logo   `json:"logo,omitempty"`
	Page             *Page   `json:"page,omitempty"`
}

type Logo struct {
	Src       *string `json:"src,omitempty"`
	MaxWidth  *int    `json:"maxWidth,omitempty"`
	MaxHeight *int    `json:"maxHeight,omitempty"`
}

type Page struct {
	Margin       *string `json:"margin,omitempty"`
	MarginTop    *string `json:"marginTop,omitempty"`
	MarginBottom *string `json:"marginBottom,omitempty"`
	MarginLeft   *string `json:"marginLeft,omitempty"`
	MarginRight  *string `json:"marginRight,omitempty"`
	PageCount    *string `json:"pageCount,omitempty"`
	Width        *int    `json:"width,omitempty"`
	Height       *int    `json:"height,omitempty"`
}

func (s *Anvil) request(method, path string, body []byte, queryParams url.Values, numRetries int) (response *http.Response, err error) {
	url := fmt.Sprintf("%s/api/%s/%s", s.BaseURL, s.RESTAPIVersion, strings.TrimLeft(path, "/"))
	if queryParams != nil && len(queryParams) > 0 {
		url += "?" + queryParams.Encode()
	}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "issue building request")
	}
	request.SetBasicAuth(s.APIKey, "")
	request.Header.Add("Content-Type", "application/json")

	for i := 0; i < numRetries+1; i++ {
		response, err := s.client.Do(request)
		if err != nil {
			return nil, errors.Wrap(err, "issue sending request")
		}
		if response.StatusCode >= 300 {
			if response.StatusCode == http.StatusTooManyRequests {
				retryAfterSec, err := strconv.Atoi(response.Header.Get("Retry-After"))
				if err != nil {
					retryAfterSec = 1
				}
				time.Sleep(time.Duration(retryAfterSec) * time.Second)
				continue
			}
			defer response.Body.Close()
			errorResponseBody, err := io.ReadAll(response.Body)
			if err == nil {
				return nil, errors.Errorf("status code: %d, body:\n%s", response.StatusCode, string(errorResponseBody))
			}
			return nil, errors.Errorf("status code: %d", response.StatusCode)
		}

		return response, nil
	}
	return nil, errors.New("rate limit exceeded")
}
