package cli

import (
	"fmt"
	"io"
)

const helpText = `GoDrop envia arquivos diretamente entre computadores na mesma rede.

Uso:
  godrop <comando>

Comandos:
  help       Exibe esta ajuda
  version    Exibe a versão instalada
`

// Version can be replaced at build time with -ldflags.
var Version = "dev"

// Run executes the command-line interface and returns a process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, helpText)
		return 0
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "godrop %s\n", Version)
		return 0
	default:
		fmt.Fprintf(stderr, "comando desconhecido: %s\n", args[0])
		fmt.Fprintln(stderr, "use 'godrop help' para ver os comandos disponíveis")
		return 2
	}
}
