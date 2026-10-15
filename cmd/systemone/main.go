// Command systemone answers questions about a codebase without the answers' source
// material entering an agent's context.
//
// An agent pays for its context on every turn, not once: a file read early in a
// session is re-sent with every later request. systemone reads the files itself,
// sends them to a model that charges $0.042 per million tokens, and returns a
// few lines. The bytes are read once, by something cheap, and never come back.
package main

import (
	"fmt"
	"os"

	"github.com/thesimpleassociate/systemone/internal/run"
	"github.com/thesimpleassociate/systemone/internal/usage"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, rootUsage)
		os.Exit(2)
	}

	var err error
	switch cmd := os.Args[1]; cmd {
	case "find":
		err = run.Find(os.Args[2:])
	case "ask":
		err = run.Ask(os.Args[2:])
	case "scan":
		err = run.Scan(os.Args[2:])
	case "hook":
		err = run.Hook(os.Args[2:])
	case "lint":
		err = run.Lint(os.Args[2:])
	case "probe":
		err = run.Probe(os.Args[2:])
	case "gain":
		history := len(os.Args) > 2 && (os.Args[2] == "--history" || os.Args[2] == "-history")
		err = usage.Report(os.Stdout, history)
	case "version", "--version", "-v":
		fmt.Printf("systemone %s\n", version)
	case "help", "--help", "-h":
		fmt.Print(rootUsage)
	default:
		fmt.Fprintf(os.Stderr, "systemone: unknown command %q\n\n%s", cmd, rootUsage)
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "systemone: %v\n", err)
		os.Exit(1)
	}
}

const rootUsage = `systemone — answer questions about code without reading it into context

usage: systemone <command> [flags] [arguments]

commands:
  find    rank files against a plain-language description of what you want
  ask     put a yes/no question to one or more files
  scan    dry run: what would be sent, and what it would cost (no API call)
  probe   verify the API contract and print a raw response
  gain    show what systemone has cost and how much it read out-of-context

  systemone find "where is the register flow"
  systemone ask "does this retry on failure?" internal/app/api.go
  systemone gain

Set SYSTEMONE_ENDPOINT to your Foundry decision API URL (required). Authenticates with Entra ID via the Azure CLI: run "az login" once. For an API key set SYSTEMONE_AUTH=key and AZURE_API_KEY. SYSTEMONE_MODEL overrides the deployment name (default Decision-1). Run "systemone <command> -h" for flags.
`
