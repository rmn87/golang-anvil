package main

import (
	"log"
	"os"

	anvil "github.com/rmn87/golang-anvil"
	"github.com/urfave/cli/v2"
)

var logger *log.Logger
var anvilAPI *anvil.Anvil

func main() {

	// init logger
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
				Action: fillPDFActionHandler,
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
				Action: generatePDFActionHandler,
			},
			{
				Name:            "create-etch",
				Usage:           "Create an etch packet with a JSON file",
				UsageText:       "anvil create-etch [-OPTIONS]",
				HideHelpCommand: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:      "payload",
						Aliases:   []string{"p"},
						Usage:     "File that contains JSON payload",
						Required:  true,
						TakesFile: true,
					},
				},
				Action: createEtchPacketHandler,
			},
			{
				Name:            "generate-etch-url",
				Usage:           "Generate an etch url for a signer",
				UsageText:       "anvil generate-etch-url [-OPTIONS]",
				HideHelpCommand: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "client",
						Aliases:  []string{"c"},
						Usage:    "The signer's user id in your system belongs here",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "signer",
						Aliases:  []string{"s"},
						Usage:    "The eid of the next signer belongs here. The signer's eid can be found in the response of the 'createEtchPacket' mutation",
						Required: true,
					},
				},
				Action: generateEtchSigningURLHandler,
			},
			{
				Name:            "download-documents",
				Usage:           "Download etch documents",
				UsageText:       "anvil download-documents [-OPTIONS]",
				HideHelpCommand: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "document-group",
						Aliases:  []string{"d"},
						Usage:    "The documentGroupEid can be found in the response of the createEtchPacket or sendEtchPacket mutations",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "filename",
						Aliases:  []string{"f"},
						Usage:    "Optional filename for the downloaded zip file",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "stdout",
						Usage:    "Instead of writing to a file, output data to STDOUT",
						Required: false,
					},
				},
				Action: downloadDocumentsHandler,
			},
		},
	}

	// run Anvil CLI
	if err := anvilCLI.Run(os.Args); err != nil {
		logger.Fatal("Error: ", err)
	}
}
