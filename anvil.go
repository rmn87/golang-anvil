// Package anvil provides an interface to access the Anvil API
package anvil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/hasura/go-graphql-client"
	"github.com/pkg/errors"
)

const TEMPLATE_VERSION_LATEST = -1
const TEMPLATE_VERSION_LATEST_PUBLISHED = -2

var VERSION = ""

type Anvil struct {
	APIKey         string
	RESTAPIVersion string
	BaseURL        string
	UserAgent      string
	Logger         *log.Logger

	restClient *http.Client
	gqlClient  *graphql.Client
}

type authInjectRoundTripper struct {
	apiKey string
}

func (rt *authInjectRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(rt.apiKey, "")
	return http.DefaultTransport.RoundTrip(req)
}

func New(apiKey string) (anvil *Anvil) {
	httpClient := http.DefaultClient
	httpClient.Transport = &authInjectRoundTripper{apiKey: apiKey}
	anvil = &Anvil{
		APIKey:         apiKey,
		RESTAPIVersion: "v1",
		BaseURL:        "https://app.useanvil.com",
		UserAgent:      "golang-anvil",
		Logger:         log.New(nil, "", 0),
		restClient:     httpClient,
		gqlClient:      graphql.NewClient("https://graphql.useanvil.com", httpClient),
	}
	if VERSION != "" {
		anvil.UserAgent += "/" + VERSION
	}
	return
}

// FillPDF fills an existing template with provided payload data.
//
// Use the casts graphql query to get a list of available templates you can use for this request.
//
// By default, the request will use the latest published version.
// You can also use the constants `anvil.TEMPLATE_VERSION_LATEST_PUBLISHED` and `anvil.TEMPLATE_VERSION_LATEST`
// instead of providing a specific version number.
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
	response, err := s.restRequest(http.MethodPost, fmt.Sprintf("fill/%s.pdf", templateID), requestBody, queryParams, 5)
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

// GeneratePDF dynamically generates a new PDF with provided payload data.
// Useful for agreements, invoices, disclosures, or any other text-heavy documents.
//
// By default, GeneratePDF will format data assuming it's in markdown.
//
// HTML is another supported input type. This can be used by providing
// `"type": "html"` in the payload and making the `data` field a dict containing
// keys `"html"` and an optional `"css"`.
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

	response, err := s.restRequest(http.MethodPost, "generate-pdf", requestBody, nil, 5)
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

// DownloadDocuments retrieves all completed documents in zip form.
func (s *Anvil) DownloadDocuments(documentGroupEID string) (zip []byte, err error) {

	response, err := s.restRequest(http.MethodGet, fmt.Sprintf("document-group/%s.zip", documentGroupEID), nil, nil, 5)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	zip, err = io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "issue reading response body")
	}
	return
}

// CreateEtchPacket creates an etch packet via a graphql mutation.
func (s *Anvil) CreateEtchPacket(payload interface{}) (etchPacketID string, err error) {
	var mutation struct {
		CreateEtchPacket struct {
			EID        string
			Name       string
			DetailsURL string
		} `graphql:" createEtchPacket (name: $name,files: $files,isDraft: $isDraft,isTest: $isTest,signatureEmailSubject: $signatureEmailSubject,signatureEmailBody: $signatureEmailBody,signatureProvider: $signatureProvider,signaturePageOptions: $signaturePageOptions,signers: $signers,data: $data)"`
	}
	var variables map[string]interface{}
	switch v := payload.(type) {
	case map[string]interface{}:
		variables = v
	case string:
		if err := json.Unmarshal([]byte(v), &variables); err != nil {
			return "", errors.Wrap(err, "issue parsing payload")
		}
	case []byte:
		if err := json.Unmarshal(v, &variables); err != nil {
			return "", errors.Wrap(err, "issue parsing payload")
		}
	case io.Reader:
		payloadBytes, err := io.ReadAll(v)
		if err != nil {
			return "", errors.Wrap(err, "issue reading payload")
		}
		if err := json.Unmarshal(payloadBytes, &variables); err != nil {
			return "", errors.Wrap(err, "issue parsing payload")
		}
	default:
		return "", errors.Errorf("payload type (%T) unsupported", v)
	}

	if err := s.gqlClient.Mutate(
		context.Background(), &mutation, variables); err != nil {
		return "", err
	}
	return mutation.CreateEtchPacket.EID, nil
}

// GenerateEtchSigningURL generates a signing URL for a given user.
func (s *Anvil) GenerateEtchSigningURL(signerEID, clientUserID string) (etchSigningURL string, err error) {
	var mutation struct {
		GenerateEtchSignURL string `graphql:"generateEtchSignURL(signerEid: $signerEid, clientUserId: $clientUserId)"`
	}
	var variables = map[string]interface{}{
		"clientUserId": clientUserID,
		"signerEid":    signerEID,
	}
	if err := s.gqlClient.Mutate(
		context.Background(), &mutation, variables); err != nil {
		return "", err
	}
	return mutation.GenerateEtchSignURL, nil
}
