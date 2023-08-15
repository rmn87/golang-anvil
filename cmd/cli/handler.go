package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	anvil "github.com/rmn87/golang-anvil"
	"github.com/urfave/cli/v2"
)

////////////////////
// Handlers
////////////////////

func downloadDocumentsHandler(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		anvilAPI.Logger = log.New(os.Stdout, "", 0)
	}
	documentGroupEID := ctx.String("document-group")

	// download documents
	zipBytes, err := anvilAPI.DownloadDocuments(documentGroupEID)
	if err != nil {
		return errors.Wrap(err, "issue downloading documents")
	}

	// write .zip file
	if ctx.Bool("stdout") {
		if _, err := os.Stdout.Write(zipBytes); err != nil {
			return errors.Wrap(err, "issue writing to stdout")
		}
	} else {
		filename := fmt.Sprintf("%s.zip", documentGroupEID)
		if filenameOption := ctx.String("filename"); filenameOption != "" {
			filename = filenameOption
		}
		if err := os.WriteFile(filename, zipBytes, 0644); err != nil {
			return errors.Wrap(err, "issue writing file")
		}
	}
	return nil
}

func createEtchPacketHandler(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		anvilAPI.Logger = log.New(os.Stdout, "", 0)
	}

	// read input file
	inputBytes, err := os.ReadFile(ctx.String("payload"))
	if err != nil {
		return errors.New("issue reading input file")
	}

	// create etch packet
	etchPacketID, err := anvilAPI.CreateEtchPacket(inputBytes)
	if err != nil {
		return errors.Wrap(err, "issue creating etch packet")
	}
	logger.Println("Etch packet created with id: ", etchPacketID)
	return nil
}

func generateEtchSigningURLHandler(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		anvilAPI.Logger = log.New(os.Stdout, "", 0)
	}

	// generate etch signing URL
	etchSigningURL, err := anvilAPI.GenerateEtchSigningURL(ctx.String("signer"), ctx.String("client"))
	if err != nil {
		return errors.Wrap(err, "issue generating etch signing URL")
	}
	logger.Println("Signing URL is: ", etchSigningURL)
	return nil
}

func fillPDFActionHandler(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		anvilAPI.Logger = log.New(os.Stdout, "", 0)
	}

	// get [TEMPLATE_ID] arg
	templateID := ctx.Args().First()
	if templateID == "" {
		cli.ShowCommandHelp(ctx, ctx.Command.Name)
		return errors.New("[TEMPLATE_ID] required")
	}

	// read input file
	inputJSONBytes, err := os.ReadFile(ctx.String("input"))
	if err != nil {
		return errors.New("issue reading input file")
	}

	// parse input
	var fillPDFPayloads []*anvil.FillPDFPayload
	if isJSONArray(inputJSONBytes) {
		if err := json.Unmarshal(inputJSONBytes, &fillPDFPayloads); err != nil {
			return errors.Wrap(err, "issue parsing JSON input")
		}
	} else {
		var fillPDFPayload anvil.FillPDFPayload
		if err := json.Unmarshal(inputJSONBytes, &fillPDFPayload); err != nil {
			return errors.Wrap(err, "issue parsing JSON input")
		}
		fillPDFPayloads = append(fillPDFPayloads, &fillPDFPayload)
	}

	// fill PDFs, store results in mem
	outputPDFs := make([][]byte, 0)
	for i, payload := range fillPDFPayloads {
		pdfBytes, err := anvilAPI.FillPDF(templateID, "", payload)
		if err != nil {
			return errors.Wrapf(err, "issue filling PDF (index %d)", i)
		}
		outputPDFs = append(outputPDFs, pdfBytes)
	}

	// write all files to disk
	return writeAllPDFsToDisk(ctx.String("out"), outputPDFs)
}

func generatePDFActionHandler(ctx *cli.Context) error {
	if ctx.Bool("debug") {
		anvilAPI.Logger = log.New(os.Stdout, "", 0)
	}

	// read input file
	inputJSONBytes, err := os.ReadFile(ctx.String("input"))
	if err != nil {
		return errors.New("issue reading input file")
	}

	// parse input
	var jsonPayloads []*anvil.GeneratePDFPayload
	if isJSONArray(inputJSONBytes) {
		if err := json.Unmarshal(inputJSONBytes, &jsonPayloads); err != nil {
			return errors.Wrap(err, "issue parsing JSON input")
		}
	} else {
		var jsonPayload anvil.GeneratePDFPayload
		if err := json.Unmarshal(inputJSONBytes, &jsonPayload); err != nil {
			return errors.Wrap(err, "issue parsing JSON input")
		}
		jsonPayloads = append(jsonPayloads, &jsonPayload)
	}

	// fill PDFs, store results in mem
	outputPDFs := make([][]byte, 0)
	for i, payload := range jsonPayloads {
		pdfBytes, err := anvilAPI.GeneratePDF(payload)
		if err != nil {
			return errors.Wrapf(err, "issue generating PDF (index %d)", i)
		}
		outputPDFs = append(outputPDFs, pdfBytes)
	}

	// write all files to disk
	return writeAllPDFsToDisk(ctx.String("out"), outputPDFs)
}

////////////////////
// Helpers
////////////////////

func isJSONArray(in []byte) bool {
	return strings.HasPrefix(strings.TrimSpace(string(in)), "[")
}

func writeAllPDFsToDisk(outputFilepath string, outputPDFs [][]byte) error {
	outFileDir := filepath.Dir(outputFilepath)
	outFileExt := filepath.Ext(outputFilepath)
	outFileName := strings.TrimSuffix(outputFilepath, outFileExt)
	for i, pdf := range outputPDFs {
		indexedOutFilepath := outputFilepath
		if i > 0 {
			indexedOutFilepath = filepath.Join(outFileDir, fmt.Sprintf("%s-%d%s", outFileName, i, outFileExt))
		}
		if err := os.WriteFile(indexedOutFilepath, pdf, 0644); err != nil {
			return errors.Wrapf(err, "issue writing file to disk (index %d)", i)
		}
	}
	return nil
}
