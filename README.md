# GoDrop

GoDrop será uma ferramenta de linha de comando para enviar arquivos diretamente
entre computadores conectados à mesma rede local.

## Estado atual

O projeto possui a base do CLI e já consegue receber arquivos por HTTP na rede
local. A próxima etapa adicionará o comando de envio.

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

## Testar

```bash
go test ./...
```
