package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Flpvoigt/Golang_project/internal/receiver"
	"github.com/Flpvoigt/Golang_project/internal/sender"
)

const helpText = `GoDrop envia arquivos diretamente entre computadores na mesma rede.

Uso:
  godrop <comando>

Comandos:
  help       Exibe esta ajuda
  version    Exibe a versão instalada
  receive    Recebe arquivos enviados pela rede local
  send       Envia um arquivo para outro computador
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
	case "receive":
		return runReceive(args[1:], stdout, stderr)
	case "send":
		return runSend(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "comando desconhecido: %s\n", args[0])
		fmt.Fprintln(stderr, "use 'godrop help' para ver os comandos disponíveis")
		return 2
	}
}

func runSend(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("send", flag.ContinueOnError)
	flags.SetOutput(stderr)
	target := flags.String("to", "", "endereço do receptor, por exemplo 192.168.1.20:8080")
	timeout := flags.Duration("timeout", 30*time.Minute, "tempo máximo da transferência")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *target == "" || flags.NArg() != 1 {
		fmt.Fprintln(stderr, "uso: godrop send --to <endereço> <arquivo>")
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	result, err := sender.Send(ctx, http.DefaultClient, *target, flags.Arg(0), stdout)
	if err != nil {
		fmt.Fprintf(stderr, "não foi possível enviar o arquivo: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Enviado: %d bytes\nSHA-256: %s\n", result.Bytes, result.Checksum)
	return 0
}

func runReceive(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("receive", flag.ContinueOnError)
	flags.SetOutput(stderr)
	port := flags.Int("port", 8080, "porta HTTP usada para receber arquivos")
	dir := flags.String("dir", "./received", "pasta onde os arquivos serão salvos")
	maxMB := flags.Int64("max-mb", 1024, "tamanho máximo de cada arquivo em MB")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *port < 1 || *port > 65535 {
		fmt.Fprintln(stderr, "a porta deve estar entre 1 e 65535")
		return 2
	}
	if *maxMB < 1 {
		fmt.Fprintln(stderr, "max-mb deve ser maior que zero")
		return 2
	}
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		fmt.Fprintf(stderr, "não foi possível criar a pasta de destino: %v\n", err)
		return 1
	}

	address := fmt.Sprintf(":%d", *port)
	fmt.Fprintf(stdout, "GoDrop aguardando arquivos em http://0.0.0.0%s/upload\n", address)
	fmt.Fprintf(stdout, "Arquivos serão salvos em %s\n", *dir)
	if err := http.ListenAndServe(address, receiver.NewHandler(*dir, *maxMB*1024*1024)); err != nil {
		fmt.Fprintf(stderr, "servidor encerrado: %v\n", err)
		return 1
	}
	return 0
}
