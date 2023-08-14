package anvil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

const TEMPLATE_VERSION_LATEST = -1
const TEMPLATE_VERSION_LATEST_PUBLISHED = -2

var APP_VERSION = ""

type Anvil struct {
	APIKey         string
	RESTAPIVersion string
	BaseURL        string
	UserAgent      string
	Debug          bool

	client *http.Client
}

func New(apiKey string) (anvil *Anvil) {
	anvil = &Anvil{
		APIKey:         apiKey,
		RESTAPIVersion: "v1",
		BaseURL:        "https://app.useanvil.com",
		UserAgent:      "golang-anvil",
		client:         http.DefaultClient,
	}
	if APP_VERSION != "" {
		anvil.UserAgent += "/" + APP_VERSION
	}
	return
}

func (s *Anvil) FillPDF(templateID, templateVersion string, payload interface{}) (pdf []byte, err error) {
	var requestBody []byte
	switch v := payload.(type) {
	case FillPDFPayload, *FillPDFPayload, map[string]interface{}:
		if requestBody, err = json.Marshal(v); err != nil {
			return nil, errors.Wrap(err, "issue encoding payload")
		}
	case string:
		requestBody = []byte(v)
	case []byte:
		requestBody = v
	case io.Reader:
		if requestBody, err = io.ReadAll(v); err != nil {
			return nil, errors.Wrap(err, "issue reading payload")
		}
	default:
		return nil, errors.Errorf("payload type (%T) unsupported", v)
	}

	var queryParams url.Values
	if templateVersion != "" {
		queryParams = url.Values{"versionNumber": {templateVersion}}
	}
	response, err := s.request(http.MethodPost, fmt.Sprintf("fill/%s.pdf", templateID), requestBody, queryParams, 5)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	pdf, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "issue reading response body")
	}
	return
}

func (s *Anvil) GeneratePDF(payload interface{}) (pdf []byte, err error) {
	var requestBody []byte
	switch v := payload.(type) {
	case GeneratePDFPayload, *GeneratePDFPayload, map[string]interface{}:
		if requestBody, err = json.Marshal(v); err != nil {
			return nil, errors.Wrap(err, "issue encoding payload")
		}
	case string:
		requestBody = []byte(v)
	case []byte:
		requestBody = v
	case io.Reader:
		if requestBody, err = io.ReadAll(v); err != nil {
			return nil, errors.Wrap(err, "issue reading payload")
		}
	default:
		return nil, errors.Errorf("payload type (%T) unsupported", v)
	}

	response, err := s.request(http.MethodPost, "generate-pdf", requestBody, nil, 5)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	pdf, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "issue reading response body")
	}
	return
}
