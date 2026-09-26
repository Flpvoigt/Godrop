# GoDrop

GoDrop será uma ferramenta de linha de comando para enviar arquivos diretamente
entre computadores conectados à mesma rede local.

## Estado atual

O projeto já envia e recebe arquivos por HTTP na rede local, com validação de
integridade por SHA-256.

## Executar

Requer Go 1.22 ou superior.

```bash
go run ./cmd/godrop help
go run ./cmd/godrop version
```

Para iniciar o receptor na porta 8080:

```bash
go run ./cmd/godrop receive --dir ./received --port 8080
```

O receptor exige o nome no cabeçalho `X-GoDrop-Filename`, aceita um checksum
opcional em `X-GoDrop-SHA256` e nunca sobrescreve arquivos existentes.

Em outro computador, envie um arquivo informando o endereço do receptor:

```bash
go run ./cmd/godrop send --to 192.168.1.20:8080 ./foto.jpg
```

O comando mostra o progresso, calcula o checksum antes do envio e apresenta o
resultado da transferência.

## Testar

```bash
go test ./...
```
