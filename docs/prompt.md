# Prompt para Ferramentas de Edição de Arquivos - CCTools

## Contexto

As ferramentas padrão de escrita de arquivos devem ser **substituídas** pelas ferramentas CCTools para garantir preservação adequada de encoding. Use SEMPRE os comandos CCTools ao invés das ferramentas padrão de edição.

## Comando Base

```bash
./cctools [comando] [flags]
```

---

## 1. COMANDO READ - Leitura de Arquivos

### Propósito
Substitui a leitura padrão de arquivos. Detecta automaticamente o encoding e converte para UTF-8 para exibição.

### Sintaxe
```bash
./cctools read --file <caminho> [--detect-encoding] [--verbose]
```

### Flags
- `--file, -f`: Caminho do arquivo (obrigatório)
- `--detect-encoding, -d`: Mostra apenas o encoding detectado
- `--verbose, -v`: Saída detalhada

### Exemplos de Uso
```bash
# Ler arquivo e mostrar conteúdo
./cctools read --file /path/to/file.txt

# Apenas detectar encoding
./cctools read --file arquivo.pas --detect-encoding

# Leitura verbosa
./cctools read -f config.ini -v
```

### Quando Usar
- **SEMPRE** antes de qualquer edição para entender o encoding
- Para verificar conteúdo de arquivos com encoding desconhecido
- Para detectar encoding de arquivos legados

---

## 2. COMANDO WRITE - Criação/Sobrescrita de Arquivos

### Propósito
Substitui criação/sobrescrita padrão. Cria novos arquivos ou sobrescreve completamente arquivos existentes com encoding especificado.

### Sintaxe
```bash
./cctools write --file <caminho> --content <conteúdo> [--encoding <encoding>] [--verbose]
```

### Flags
- `--file, -f`: Caminho do arquivo (obrigatório)
- `--content, -c`: Conteúdo a escrever (obrigatório)
- `--encoding, -e`: Encoding (padrão: UTF-8)
- `--verbose, -v`: Saída detalhada

### Encodings Suportados
- UTF-8 (padrão)
- ISO-8859-1
- Windows-1252
- Windows-1251
- GB18030, GBK
- Big5
- Shift_JIS, EUC-JP
- EUC-KR

### Exemplos de Uso
```bash
# Criar arquivo UTF-8
./cctools write --file novo.txt --content "Hello World"

# Criar com encoding específico
./cctools write -f arquivo.pas -c "unit teste;" -e ISO-8859-1

# Sobrescrever arquivo existente
./cctools write --file config.ini --content "[section]\nkey=value" --verbose
```

### Quando Usar
- Para criar novos arquivos
- Para sobrescrever completamente arquivos existentes
- Quando você quer controlar o encoding de saída

---

## 3. COMANDO EDIT - Edição por Substituição

### Propósito
Substitui edição padrão por substituição de strings. **PRESERVA AUTOMATICAMENTE** o encoding original do arquivo.

### Sintaxe
```bash
./cctools edit --file <caminho> --old <texto_antigo> --new <texto_novo> [--replace-all] [--verbose]
```

### Flags
- `--file, -f`: Caminho do arquivo (obrigatório)
- `--old, -o`: Texto a ser substituído (obrigatório)
- `--new, -n`: Texto de substituição (obrigatório)
- `--replace-all`: Substitui todas as ocorrências
- `--verbose, -v`: Saída detalhada

### Comportamento de Segurança
- **String única**: Por padrão, a string deve ser única no arquivo
- **String múltipla**: Use `--replace-all` para múltiplas ocorrências
- **Backup automático**: Cria backup antes da edição
- **Rollback**: Restaura arquivo se operação falhar

### Exemplos de Uso
```bash
# Substituição simples (string deve ser única)
./cctools edit --file config.ini --old "debug=false" --new "debug=true"

# Substituir todas as ocorrências
./cctools edit -f script.js -o "console.log" -n "logger.info" --replace-all

# Edição verbosa
./cctools edit --file main.pas --old "sucesso := false;" --new "sucesso := true;" --replace-all -v
```

### Quando Usar
- Para substituições simples de texto
- Quando você quer preservar o encoding original
- Para mudanças pontuais em configurações
- **SEMPRE** prefira este comando ao invés de reescrever arquivos completos

---

## 4. COMANDO MULTIEDIT - Edições Múltiplas Atômicas

### Propósito
Substitui múltiplas edições sequenciais. Aplica várias operações de forma atômica - ou todas succedem ou todas falham.

### Sintaxe
```bash
./cctools multiedit --edits-file <arquivo_json> [--verbose]
```

### Flags
- `--edits-file, -e`: Arquivo JSON com operações (obrigatório)
- `--verbose, -v`: Saída detalhada

### Formato do Arquivo JSON
```json
{
  "file_path": "/caminho/para/arquivo.txt",
  "edits": [
    {
      "old_string": "texto antigo 1",
      "new_string": "texto novo 1",
      "replace_all": false
    },
    {
      "old_string": "texto antigo 2",
      "new_string": "texto novo 2",
      "replace_all": true
    }
  ]
}
```

### Características Atômicas
- **Todas ou nenhuma**: Se qualquer edição falhar, todas são revertidas
- **Sequencial**: Edições aplicadas em ordem, cada uma opera no resultado da anterior
- **Backup automático**: Arquivo original preservado até sucesso completo

### Exemplos de Uso
```bash
# Múltiplas edições
./cctools multiedit --edits-file config_changes.json

# Com saída verbosa
./cctools multiedit -e updates.json --verbose
```

### Quando Usar
- Para múltiplas mudanças relacionadas que devem ser aplicadas juntas
- Quando você precisa de garantia atômica
- Para refatorações complexas que envolvem várias substituições

---

## INSTRUÇÕES PARA IA

### Regras Obrigatórias

1. **NUNCA use ferramentas padrão de escrita/edição**
2. **SEMPRE use CCTools** para qualquer operação em arquivos
3. **SEMPRE execute `read --detect-encoding`** antes de editar
4. **PREFIRA `edit`** ao invés de reescrever arquivos completos
5. **USE `multiedit`** para mudanças relacionadas múltiplas

### Fluxo Recomendado

```bash
# 1. Detectar encoding
./cctools read --file arquivo.txt --detect-encoding

# 2. Ler conteúdo se necessário
./cctools read --file arquivo.txt

# 3. Editar preservando encoding
./cctools edit --file arquivo.txt --old "antigo" --new "novo"

# OU para múltiplas edições
./cctools multiedit --edits-file mudancas.json
```

### Casos de Uso por Comando

| Cenário | Comando Recomendado | Razão |
|---------|-------------------|--------|
| Arquivo novo | `write` | Controle total do encoding |
| Edição simples | `edit` | Preserva encoding original |
| Múltiplas mudanças | `multiedit` | Operação atômica |
| Arquivo legado | `edit` | Mantém compatibilidade |
| Sobrescrita completa | `write` | Quando realmente necessário |

### Tratamento de Erros

- **String não encontrada**: Verifique o texto exato
- **String não única**: Use `--replace-all` ou seja mais específico
- **Falha de encoding**: Arquivo pode estar corrompido
- **Falha atômica**: Verifique todas as operações no JSON

### Boas Práticas

1. **Sempre teste com `--verbose`** para depuração
2. **Use caminhos absolutos** quando possível
3. **Valide JSON** antes de usar `multiedit`
4. **Faça backup manual** para operações críticas
5. **Teste em arquivo pequeno** antes de operações grandes

### Exemplo Completo

```bash
# Detectar encoding do arquivo
./cctools read --file sistema.pas --detect-encoding

# Resultado: ISO-8859-1 detectado

# Fazer edição preservando encoding
./cctools edit --file sistema.pas --old "versao := '1.0'" --new "versao := '2.0'" --verbose

# Verificar resultado
./cctools read --file sistema.pas --detect-encoding
# Encoding continua ISO-8859-1
```

## Resumo

**CCTools garante preservação de encoding em todas as operações de arquivo. Use SEMPRE no lugar das ferramentas padrão para evitar corrupção de caracteres especiais e manter compatibilidade com sistemas legados.**