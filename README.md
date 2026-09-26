# GoDrop

GoDrop será uma ferramenta de linha de comando para enviar arquivos diretamente
entre computadores conectados à mesma rede local.

## Estado atual

O projeto possui a base do CLI, com comandos de ajuda e versão. As próximas
etapas adicionarão o recebimento e o envio de arquivos.

## Executar

Requer Go 1.22 ou superior.

```bash
go run ./cmd/godrop help
go run ./cmd/godrop version
```

## Testar

```bash
go test ./...
```
