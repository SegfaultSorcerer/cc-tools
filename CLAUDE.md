# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CCTools é uma ferramenta CLI em Go para edição de arquivos que **preserva automaticamente a codificação original** dos arquivos durante as operações. É especialmente útil para trabalhar com códigos legados, projetos internacionais e arquivos com diferentes codificações de caracteres.

## Build and Development Commands

### Build para plataforma atual:
```bash
go build -o cctools
```

### Build multiplataforma:
```bash
chmod +x build.sh
./build.sh
```

### Gerenciamento de dependências:
```bash
go mod tidy
```

### Execução direta:
```bash
go run main.go [command] [flags]
```

## Arquitetura do Código

### Estrutura principal:
- **`main.go`**: Entry point que delega para cmd.Execute()
- **`cmd/`**: Comandos CLI usando framework Cobra
  - `root.go`: Comando raiz e configuração global
  - `read.go`, `write.go`, `edit.go`, `multiedit.go`: Implementação dos comandos principais
- **`pkg/encoding/`**: Detecção e conversão de codificação de caracteres
- **`pkg/fileops/`**: Operações de arquivo com suporte a múltiplas codificações
- **`internal/models/`**: Estruturas de dados compartilhadas

### Fluxo de operações:
1. **Detecção**: `encoding.Detector` identifica a codificação do arquivo usando chardet
2. **Operação**: `fileops.FileOperations` executa a operação preservando encoding
3. **Segurança**: Backup automático e rollback em caso de falha
4. **Atomicidade**: MultiEdit garante que todas as operações succedem ou falham juntas

### Codificações suportadas:
- UTF-8, UTF-16 (LE/BE)
- ISO-8859-1, ISO-8859-15
- Windows-1252, Windows-1251
- GB18030, GBK, Big5
- Shift_JIS, EUC-JP, EUC-KR

## Comandos disponíveis

### cctools read
Lê arquivos com detecção automática de encoding:
```bash
./cctools read --file arquivo.txt [--detect-encoding] [--verbose]
```

### cctools write
Cria/sobrescreve arquivos com encoding especificado:
```bash
./cctools write --file arquivo.txt --content "conteúdo" [--encoding UTF-8]
```

### cctools edit
Edita arquivos preservando encoding original:
```bash
./cctools edit --file arquivo.txt --old "texto antigo" --new "texto novo" [--replace-all]
```

### cctools multiedit
Múltiplas edições atômicas através de arquivo JSON:
```bash
./cctools multiedit --edits-file operacoes.json
```

## Padrões de desenvolvimento

### Quando trabalhar com este código:
1. **Use sempre encoding-aware operations**: O projeto foi criado especificamente para preservar encodings
2. **Teste com arquivos de diferentes encodings**: Verifique compatibilidade com ISO-8859-1, Windows-1252, etc.
3. **Mantenha operações atômicas**: Especialmente importante no MultiEdit
4. **Preserve a arquitetura em camadas**:
   - CLI commands (cmd/) → File operations (pkg/fileops/) → Encoding handling (pkg/encoding/)

### Testing approach:
- Teste com arquivos de diferentes encodings na pasta `test_files/` (se existir)
- Verifique se encoding é preservado após edições
- Teste rollback em cenários de falha

### Cross-platform considerations:
- O build.sh gera binários para Windows, Linux, macOS e FreeBSD
- Line endings são preservados conforme sistema operacional
- Caminhos de arquivo devem ser tratados adequadamente em todas as plataformas

## Documentação técnica

Consulte `docs/tools.md` para detalhes técnicos das ferramentas de edição e `docs/prompt.md` para instruções específicas sobre uso das ferramentas CCTools.