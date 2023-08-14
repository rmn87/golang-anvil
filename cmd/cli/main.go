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

var logger *log.Logger
var anvilAPI *anvil.Anvil

func main() {

	// init debug logger
	logger = log.New(os.Stdout, "", 0)

	// verify Anvil API Key
	apiKey := os.Getenv("ANVIL_API_KEY")
	if apiKey == "" {
		logger.Fatal("$ANVIL_API_KEY must be defined in your environment variables")
	}

	// init Anvil API
	anvilAPI = anvil.New(apiKey)

	// define Anvil CLI
	anvilCLI := &cli.App{
		Name:            "anvil",
		Usage:           "A CLI for the Anvil API",
		HideVersion:     false,
		Version:         anvil.VERSION,
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "debug",
				Usage:    "show debug logs",
				Required: false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:            "fill-pdf",
				Usage:           "Fill PDF template with data",
				UsageText:       "anvil fill-pdf [-OPTIONS] [TEMPLATE_ID]",
				HideHelpCommand: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "out",
						Aliases:  []string{"o"},
						Usage:    "Filename of output PDF",
						Required: true,
					},
					&cli.StringFlag{
						Name:      "input",
						Aliases:   []string{"i"},
						Usage:     "Filename of JSON payload input",
						Required:  true,
						TakesFile: true,
					},
				},
				Action: fillPDFAction,
			},
			{
				Name:            "generate-pdf",
				Usage:           "Generate a PDF",
				UsageText:       "anvil generate-pdf [-OPTIONS]",
				HideHelpCommand: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "out",
						Aliases:  []string{"o"},
						Usage:    "Filename of output PDF",
						Required: true,
					},
					&cli.StringFlag{
						Name:      "input",
						Aliases:   []string{"i"},
						Usage:     "Filename of JSON payload input",
						Required:  true,
						TakesFile: true,
					},
				},
				Action: generatePDFAction,
			},
		},
	}

	// run Anvil CLI
	if err := anvilCLI.Run(os.Args); err != nil {
		logger.Fatal("Error: ", err)
	}
}

func fillPDFAction(ctx *cli.Context) error {
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
	var jsonPayloads []*anvil.FillPDFPayload
	if isJSONArray(inputJSONBytes) {
		if err := json.Unmarshal(inputJSONBytes, &jsonPayloads); err != nil {
			return errors.Wrap(err, "issue parsing JSON input")
		}
	} else {
		var jsonPayload anvil.FillPDFPayload
		if err := json.Unmarshal(inputJSONBytes, &jsonPayload); err != nil {
			return errors.Wrap(err, "issue parsing JSON input")
		}
		jsonPayloads = append(jsonPayloads, &jsonPayload)
	}

	// fill PDFs, store results in mem
	outputPDFs := make([][]byte, 0)
	for i, payload := range jsonPayloads {
		pdfBytes, err := anvilAPI.FillPDF(templateID, "", payload)
		if err != nil {
			return errors.Wrapf(err, "issue filling PDF (index %d)", i)
		}
		outputPDFs = append(outputPDFs, pdfBytes)
	}

	// write all files to disk
	return writeAllPDFsToDisk(ctx.String("out"), outputPDFs)
}

func isJSONArray(in []byte) bool {
	if strings.HasPrefix(strings.TrimSpace(string(in)), "[") {
		return true
	}
	return false
}

func generatePDFAction(ctx *cli.Context) error {
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
