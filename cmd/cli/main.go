package main

import (
	"log"
	"os"

	anvil "github.com/rmn87/golang-anvil"
	"github.com/urfave/cli/v2"
)

var debugLogger *log.Logger
var anvilService *anvil.Anvil

func main() {

	// init debug logger
	debugLogger = log.New(os.Stdout, "", 0)

	// verify Anvil API Key
	apiKey := os.Getenv("ANVIL_API_KEY")
	if apiKey == "" {
		log.Fatal("$ANVIL_API_KEY must be defined in your environment variables")
	}

	// init Anvil API
	anvilService = anvil.New(apiKey)

	// define Anvil CLI
	app := &cli.App{
		Name:            "anvil",
		Usage:           "A CLI for the Anvil API",
		Version:         anvil.APP_VERSION,
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
				Name:      "fill-pdf",
				Usage:     "Fill PDF template with data",
				UsageText: "anvil fill-pdf [-OPTIONS] [TEMPLATE_ID]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "out",
						Aliases:  []string{"o"},
						Usage:    "Filename of output PDF",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "Filename of input CSV that provides data",
						Required: true,
					},
				},
				Action: fillPDFAction,
			},
			{
				Name:      "generate-pdf",
				Usage:     "Generate a PDF",
				UsageText: "anvil generate-pdf [-OPTIONS]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "out",
						Aliases:  []string{"o"},
						Usage:    "Filename of output PDF",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "input",
						Aliases:  []string{"i"},
						Usage:    "Filename of input payload",
						Required: true,
					},
				},
				Action: generatePDFAction,
			},
		},
	}

	// run Anvil CLI
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func fillPDFAction(ctx *cli.Context) error {
	anvilService.Debug = ctx.Bool("debug")

	return nil
}
func generatePDFAction(ctx *cli.Context) error {
	anvilService.Debug = ctx.Bool("debug")
	return nil
}
