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
# Conteúdo direto
./cctools write --file <caminho> --content <conteúdo> [--encoding <encoding>] [--verbose]

# De arquivo
./cctools write --file <caminho> --content-file <arquivo_origem> [--encoding <encoding>] [--verbose]

# De stdin
./cctools write --file <caminho> --stdin [--encoding <encoding>] [--verbose]
```

### Flags
- `--file, -f`: Caminho do arquivo (obrigatório)
- `--content, -c`: Conteúdo a escrever
- `--content-file`: Ler conteúdo de arquivo especificado
- `--stdin`: Ler conteúdo do stdin
- `--encoding, -e`: Encoding (padrão: UTF-8)
- `--verbose, -v`: Saída detalhada

**Nota:** Exatamente uma das opções de conteúdo deve ser especificada (`--content`, `--content-file`, ou `--stdin`).

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
# Criar arquivo UTF-8 com conteúdo direto
./cctools write --file novo.txt --content "Hello World"

# Copiar conteúdo de outro arquivo
./cctools write --file backup.txt --content-file original.txt

# Criar arquivo via pipe
echo "Dados importantes" | ./cctools write --file saida.txt --stdin

# Criar arquivo via input interativo
./cctools write --file dados.txt --stdin

# Criar com encoding específico
./cctools write -f arquivo.pas --content "unit teste;" -e ISO-8859-1

# Sobrescrever arquivo com conteúdo de stdin
cat dados_grandes.txt | ./cctools write --file processado.txt --stdin --verbose
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
./cctools edit --file <caminho> --old <texto_antigo> --new <texto_novo> [--replace-all] [--preview] [--regex] [--fuzzy] [--ignore-whitespace] [--case-insensitive] [--verbose]
```

### Flags
- `--file, -f`: Caminho do arquivo (obrigatório)
- `--old, -o`: Texto a ser substituído (obrigatório)
- `--new, -n`: Texto de substituição (obrigatório)
- `--replace-all`: Substitui todas as ocorrências
- `--preview`: Mostra prévia das mudanças sem aplicá-las
- `--regex`: Trata old string como expressão regular
- `--fuzzy`: Habilita matching fuzzy tolerante a diferenças
- `--ignore-whitespace`: Ignora diferenças de espaçamento
- `--case-insensitive`: Busca case-insensitive
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

# Preview antes de aplicar mudanças
./cctools edit --file main.pas --old "sucesso := false;" --new "sucesso := true;" --preview

# Usar regex para substituições avançadas
./cctools edit -f codigo.js -o "function\s+\w+" -n "async function" --regex --replace-all

# Fuzzy matching para strings com pequenas diferenças
./cctools edit --file arquivo.txt --old "texto aproximado" --new "texto novo" --fuzzy

# Ignorar diferenças de espaçamento
./cctools edit -f code.py -o "if   condition:" -n "if condition:" --ignore-whitespace

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
./cctools multiedit --edits-file <arquivo_json> [--preview] [--continue-on-error] [--dry-run] [--verbose]
```

### Flags
- `--edits-file, -e`: Arquivo JSON com operações (obrigatório)
- `--preview`: Mostra prévia de todas as mudanças sem aplicá-las
- `--continue-on-error`: Continua processando mesmo se edições individuais falharem
- `--dry-run`: Executa simulação mostrando o que seria alterado
- `--verbose, -v`: Saída detalhada

### Formato do Arquivo JSON
```json
{
  "file_path": "/caminho/para/arquivo.txt",
  "continue_on_error": false,
  "dry_run": false,
  "edits": [
    {
      "old_string": "texto antigo 1",
      "new_string": "texto novo 1",
      "replace_all": false,
      "use_regex": false,
      "fuzzy_match": false
    },
    {
      "old_string": "\\w+\\.log",
      "new_string": "debug.log",
      "replace_all": true,
      "use_regex": true
    },
    {
      "old_string": "texto aproximado",
      "new_string": "texto correto",
      "fuzzy_match": true
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
# Múltiplas edições normais
./cctools multiedit --edits-file config_changes.json

# Preview antes de aplicar
./cctools multiedit --edits-file updates.json --preview

# Dry run para testar
./cctools multiedit -e refactor.json --dry-run

# Continuar mesmo com erros
./cctools multiedit --edits-file big_refactor.json --continue-on-error

# Com saída verbosa detalhada
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

---

## 5. COMANDO COPY - Cópia de Arquivos

### Propósito
Copia arquivos preservando automaticamente o encoding original. Útil para criar backups ou duplicar arquivos sem riscos de corrupção de caracteres.

### Sintaxe
```bash
./cctools copy --source <origem> --dest <destino> [--preserve-mode] [--overwrite] [--verbose]
```

### Flags
- `--source, -s`: Caminho do arquivo origem (obrigatório)
- `--dest, -d`: Caminho do arquivo destino (obrigatório)
- `--preserve-mode, -p`: Preserva permissões do arquivo original
- `--overwrite, -o`: Sobrescreve destino se existir
- `--verbose, -v`: Saída detalhada

### Exemplos de Uso
```bash
# Cópia simples
./cctools copy --source arquivo.txt --dest backup.txt

# Cópia preservando permissões
./cctools copy -s sistema.pas -d /backup/sistema.pas --preserve-mode

# Cópia com sobrescrita
./cctools copy --source config.ini --dest /new/config.ini --overwrite -v
```

### Quando Usar
- Para criar backups de arquivos importantes
- Para duplicar arquivos mantendo encoding
- Quando você precisa copiar arquivos legados sem corrupção

---

## 6. COMANDO MOVE - Movimentação de Arquivos

### Propósito
Move arquivos preservando automaticamente o encoding original. Operação atômica com rollback automático em caso de falha.

### Sintaxe
```bash
./cctools move --source <origem> --dest <destino> [--overwrite] [--verbose]
```

### Flags
- `--source, -s`: Caminho do arquivo origem (obrigatório)
- `--dest, -d`: Caminho do arquivo destino (obrigatório)
- `--overwrite, -o`: Sobrescreve destino se existir
- `--verbose, -v`: Saída detalhada

### Características de Segurança
- **Backup automático**: Cria backup antes da operação
- **Rollback**: Restaura arquivo original se operação falhar
- **Operação atômica**: Ou move completamente ou mantém original

### Exemplos de Uso
```bash
# Movimentação simples
./cctools move --source arquivo.txt --dest /nova/pasta/arquivo.txt

# Move com sobrescrita
./cctools move -s old_config.ini -d new_config.ini --overwrite

# Move verboso
./cctools move --source sistema.pas --dest /projeto/sistema.pas -v
```

### Quando Usar
- Para reorganizar arquivos mantendo encoding
- Para renomear arquivos com segurança
- Quando você precisa mover arquivos legados

---

## 7. COMANDO DELETE - Exclusão de Arquivos

### Propósito
Deleta arquivos com opção de backup para recuperação. Ideal para exclusão segura de arquivos importantes.

### Sintaxe
```bash
./cctools delete --file <arquivo> [--backup] [--backup-path <caminho>] [--verbose]
```

### Flags
- `--file, -f`: Caminho do arquivo a deletar (obrigatório)
- `--backup, -b`: Cria backup antes da exclusão
- `--backup-path`: Caminho personalizado para backup
- `--verbose, -v`: Saída detalhada

### Comportamento de Segurança
- **Backup opcional**: Cria cópia antes de deletar
- **Caminho personalizado**: Especifica onde salvar backup
- **Verificação prévia**: Confirma existência antes de deletar

### Exemplos de Uso
```bash
# Exclusão simples
./cctools delete --file arquivo_temporario.txt

# Exclusão com backup
./cctools delete --file config.ini --backup

# Exclusão com backup personalizado
./cctools delete -f sistema.pas --backup --backup-path /safe/sistema.pas.bak
```

### Quando Usar
- Para deletar arquivos com segurança
- Quando você quer manter backup antes da exclusão
- Para limpar arquivos preservando possibilidade de recuperação

---

## FUNCIONALIDADES AVANÇADAS

### Enhanced String Matching

O CCTools agora suporta múltiplas estratégias de matching para resolver os problemas comuns de detecção de strings:

#### 1. Matching Fuzzy (--fuzzy)
- **Problema resolvido**: Strings com pequenas diferenças de espaçamento ou formatação
- **Como usar**: Adicione `--fuzzy` aos comandos edit
- **Exemplo**:
```bash
# Funciona mesmo se houver espaços extras ou diferentes
./cctools edit -f arquivo.pas --old "if condition then" --new "if nova_condition then" --fuzzy
```

#### 2. Regex Support (--regex)
- **Problema resolvido**: Necessidade de patterns complexos para matching
- **Como usar**: Adicione `--regex` aos comandos edit
- **Exemplo**:
```bash
# Substitui qualquer função que termine com "Old"
./cctools edit -f code.js --old "function\\s+\\w+Old" --new "function newFunction" --regex --replace-all
```

#### 3. Ignore Whitespace (--ignore-whitespace)
- **Problema resolvido**: Diferenças de indentação e espaçamento
- **Como usar**: Adicione `--ignore-whitespace` aos comandos edit
- **Exemplo**:
```bash
# Ignora espaços extras entre palavras
./cctools edit -f code.py --old "if    condition:" --new "if condition:" --ignore-whitespace
```

#### 4. Case Insensitive (--case-insensitive)
- **Problema resolvido**: Diferenças de maiúsculas/minúsculas
- **Como usar**: Adicione `--case-insensitive` aos comandos edit

### Preview Mode

Visualize mudanças antes de aplicá-las para evitar erros:

```bash
# Preview para edit simples
./cctools edit -f arquivo.txt --old "antigo" --new "novo" --preview

# Preview para multiedit
./cctools multiedit -e changes.json --preview

# Dry run completo
./cctools multiedit -e refactor.json --dry-run
```

### Continue on Error

Para operações grandes onde alguns erros são aceitáveis:

```bash
# Continua mesmo se algumas edições falharem
./cctools multiedit -e big_changes.json --continue-on-error
```

### Input Flexível para Write

Resolução dos problemas de "unexpected EOF" com conteúdo extenso:

```bash
# Para arquivos grandes - leia de arquivo
./cctools write --file output.txt --content-file input.txt

# Para pipes - use stdin
cat large_file.txt | ./cctools write --file processed.txt --stdin

# Para input interativo
./cctools write --file notes.txt --stdin
```

### Mensagens de Erro Aprimoradas

O sistema agora fornece:
- **Contexto visual** das linhas adjacentes
- **Sugestões de matches similares** quando string não é encontrada
- **Preview diff** das mudanças propostas
- **Informações detalhadas** sobre cada match encontrado

---

## INSTRUÇÕES PARA IA - ATUALIZADAS

### Regras Obrigatórias

1. **NUNCA use ferramentas padrão de manipulação de arquivos**
2. **SEMPRE use CCTools** para qualquer operação em arquivos
3. **SEMPRE execute `read --detect-encoding`** antes de operações
4. **PREFIRA `copy`** ao invés de recriar arquivos
5. **USE `move`** para reorganização segura
6. **USE `delete --backup`** para exclusões seguras
7. **USE `multiedit`** para mudanças relacionadas múltiplas

### Fluxo Recomendado para Manipulação

```bash
# 1. Detectar encoding
./cctools read --file arquivo.txt --detect-encoding

# 2. Para cópia segura
./cctools copy --source arquivo.txt --dest backup.txt --preserve-mode

# 3. Para movimentação segura
./cctools move --source old_location.txt --dest new_location.txt

# 4. Para exclusão segura
./cctools delete --file unwanted.txt --backup
```

### Casos de Uso Atualizados

| Cenário | Comando Recomendado | Razão |
|---------|-------------------|-------|
| Arquivo novo | `write` | Controle total do encoding |
| Edição simples | `edit` | Preserva encoding original |
| Múltiplas mudanças | `multiedit` | Operação atômica |
| Backup de arquivo | `copy --preserve-mode` | Mantém características originais |
| Reorganização | `move` | Operação segura com rollback |
| Limpeza segura | `delete --backup` | Permite recuperação |
| Arquivo legado | `edit` ou `copy` | Mantém compatibilidade |

### Exemplos Completos de Fluxos

#### Backup e Edição Segura
```bash
# 1. Criar backup
./cctools copy --source sistema.pas --dest sistema.pas.backup --preserve-mode

# 2. Detectar encoding
./cctools read --file sistema.pas --detect-encoding

# 3. Fazer edição
./cctools edit --file sistema.pas --old "versao := '1.0'" --new "versao := '2.0'"
```

#### Reorganização de Projeto
```bash
# Mover arquivos mantendo encoding
./cctools move --source old/config.ini --dest new/structure/config.ini
./cctools move --source old/sistema.pas --dest new/structure/sistema.pas

# Limpar pasta antiga
./cctools delete --file old/temp.log --backup
```

---

## 8. COMANDO MKDIR - Criação de Diretórios

### Propósito
Cria diretórios únicos ou estruturas completas com permissões customizáveis.

### Sintaxe
```bash
./cctools mkdir --path <diretório> [--parents] [--mode <permissões>] [--verbose]
```

### Flags
- `--path, -p`: Caminho do diretório a criar (obrigatório)
- `--parents`: Cria diretórios pai conforme necessário
- `--mode`: Permissões em formato octal (padrão: 755)
- `--verbose, -v`: Saída detalhada

### Exemplos de Uso
```bash
# Criar diretório simples
./cctools mkdir --path novo_projeto

# Criar estrutura completa
./cctools mkdir --path projetos/web/src --parents

# Criar com permissões específicas
./cctools mkdir -p config/ssl --parents --mode 700
```

---

## 9. COMANDO COPYDIR - Cópia de Diretórios

### Propósito
Copia diretórios recursivamente preservando encoding de todos os arquivos. Ideal para backups e duplicação de projetos.

### Sintaxe
```bash
./cctools copydir --source <origem> --dest <destino> [--preserve-all] [--overwrite] [--skip-existing] [--verbose]
```

### Flags
- `--source, -s`: Diretório origem (obrigatório)
- `--dest, -d`: Diretório destino (obrigatório)
- `--preserve-all`: Preserva permissões, timestamps e atributos
- `--overwrite, -o`: Sobrescreve destino se existir
- `--skip-existing`: Pula arquivos que já existem no destino
- `--verbose, -v`: Saída detalhada com estatísticas

### Exemplos de Uso
```bash
# Cópia simples de projeto
./cctools copydir --source meu_projeto/ --dest backup_projeto/

# Cópia preservando tudo
./cctools copydir -s sistema/ -d /backup/sistema/ --preserve-all

# Cópia incremental
./cctools copydir --source src/ --dest mirror/ --skip-existing --overwrite
```

---

## 10. COMANDO MOVEDIR - Movimentação de Diretórios

### Propósito
Move diretórios com operação atômica e rollback completo. Tenta rename eficiente primeiro, depois copy+delete com segurança.

### Sintaxe
```bash
./cctools movedir --source <origem> --dest <destino> [--overwrite] [--verbose]
```

### Flags
- `--source, -s`: Diretório origem (obrigatório)
- `--dest, -d`: Diretório destino (obrigatório)
- `--overwrite, -o`: Sobrescreve destino se existir
- `--verbose, -v`: Saída detalhada

### Características Atômicas
- **Rollback completo**: Se falhar, restaura estado original
- **Operação eficiente**: Usa rename quando possível
- **Preservação total**: Mantém todos os encodings e atributos

### Exemplos de Uso
```bash
# Mover projeto para nova localização
./cctools movedir --source projeto_v1/ --dest projeto_v2/

# Reorganizar com sobrescrita
./cctools movedir -s temp/dados/ -d archive/dados/ --overwrite
```

---

## 11. COMANDO RMDIR - Remoção de Diretórios

### Propósito
Remove diretórios vazios ou recursivamente com backup opcional para recuperação completa.

### Sintaxe
```bash
./cctools rmdir --path <diretório> [--recursive] [--backup] [--backup-path <caminho>] [--verbose]
```

### Flags
- `--path, -p`: Diretório a remover (obrigatório)
- `--recursive, -r`: Remove recursivamente (conteúdo completo)
- `--backup, -b`: Cria backup antes da remoção
- `--backup-path`: Caminho personalizado para backup
- `--verbose, -v`: Saída detalhada

### Exemplos de Uso
```bash
# Remover diretório vazio
./cctools rmdir --path temp_empty/

# Remover recursivamente com backup
./cctools rmdir --path old_project/ --recursive --backup

# Remover com backup personalizado
./cctools rmdir -p dados/ -r --backup --backup-path /safe/dados_backup/
```

---

## 12. COMANDO LISTDIR - Listagem Inteligente

### Propósito
Lista conteúdo de diretórios com detecção de encoding, filtros e análise estatística.

### Sintaxe
```bash
./cctools listdir [--path <diretório>] [--recursive] [--show-encoding] [--filter <padrão>] [--show-hidden] [--verbose]
```

### Flags
- `--path, -p`: Diretório a listar (padrão: atual)
- `--recursive, -r`: Lista recursivamente
- `--show-encoding`: Detecta e mostra encoding dos arquivos
- `--filter`: Filtra por padrão (ex: "*.pas", "*.go")
- `--show-hidden`: Mostra arquivos ocultos
- `--verbose, -v`: Estatísticas detalhadas

### Exemplos de Uso
```bash
# Listagem simples
./cctools listdir --path projeto/

# Análise completa com encodings
./cctools listdir -p . --recursive --show-encoding --verbose

# Filtrar arquivos específicos
./cctools listdir --filter "*.pas" --show-encoding

# Análise completa de projeto
./cctools listdir --recursive --show-encoding --show-hidden --verbose
```

---

## INSTRUÇÕES PARA IA - ATUALIZADA COMPLETA

### Regras Obrigatórias

1. **NUNCA use ferramentas padrão de manipulação de arquivos e diretórios**
2. **SEMPRE use CCTools** para qualquer operação em arquivos e diretórios
3. **SEMPRE execute análise prévia** com `read --detect-encoding` ou `listdir --show-encoding`
4. **SEMPRE use --preview primeiro** em operações complexas para verificar matches
5. **PREFIRA operações específicas**:
   - `copy/copydir` ao invés de recriar
   - `move/movedir` para reorganização
   - `delete/rmdir --backup` para exclusões
   - `mkdir --parents` para estruturas
6. **USE `multiedit`** para mudanças relacionadas múltiplas em arquivos
7. **USE matching avançado** quando strings exatas falharem:
   - `--fuzzy` para diferenças de formatação
   - `--regex` para patterns complexos
   - `--ignore-whitespace` para problemas de espaçamento
8. **USE input flexível para write**:
   - `--content-file` para conteúdo de outros arquivos
   - `--stdin` para pipes e entrada grande

### Fluxo Recomendado Completo

```bash
# 1. Análise inicial do projeto
./cctools listdir --path . --recursive --show-encoding --verbose

# 2. Criar estrutura de backup
./cctools mkdir --path backups/$(date +%Y%m%d) --parents

# 3. Backup completo do projeto
./cctools copydir --source . --dest backups/$(date +%Y%m%d)/ --preserve-all

# 4. Preview antes de edições complexas
./cctools edit --file arquivo.pas --old "texto_complexo" --new "novo_texto" --preview

# 5. Operações nos arquivos (preservando encoding)
./cctools edit --file arquivo.pas --old "antigo" --new "novo"

# 6. Para strings problemáticas, use matching avançado
./cctools edit --file arquivo.pas --old "string aproximada" --new "nova string" --fuzzy
./cctools edit --file arquivo.js --old "function\\s+\\w+" --new "async function" --regex --replace-all

# 7. Reorganização de estrutura
./cctools movedir --source old_structure/ --dest new_structure/

# 8. Limpeza segura
./cctools rmdir --path temp/ --recursive --backup
```

### Casos de Uso Completos

| Cenário | Comando Recomendado | Razão |
|---------|-------------------|-------|
| Arquivo novo | `write` | Controle total do encoding |
| Edição simples | `edit` | Preserva encoding original |
| Múltiplas mudanças | `multiedit` | Operação atômica |
| Backup de arquivo | `copy --preserve-mode` | Mantém características |
| Backup de projeto | `copydir --preserve-all` | Backup completo |
| Reorganização arquivo | `move` | Operação segura |
| Reorganização projeto | `movedir` | Operação atômica |
| Limpeza arquivo | `delete --backup` | Permite recuperação |
| Limpeza projeto | `rmdir --recursive --backup` | Backup completo |
| Criar estrutura | `mkdir --parents` | Estruturas completas |
| Análise projeto | `listdir --recursive --show-encoding` | Visão completa |

### Fluxos Especializados

#### Backup e Refatoração Completa
```bash
# 1. Análise inicial
./cctools listdir --recursive --show-encoding --verbose

# 2. Backup completo
./cctools copydir --source . --dest ../backup_$(date +%Y%m%d)/ --preserve-all

# 3. Criar nova estrutura
./cctools mkdir --path nova_estrutura/src/main --parents
./cctools mkdir --path nova_estrutura/docs --parents

# 4. Mover componentes
./cctools movedir --source src/ --dest nova_estrutura/src/
./cctools copydir --source docs/ --dest nova_estrutura/docs/

# 5. Editar arquivos de configuração
./cctools edit --file config.ini --old "old_path" --new "nova_estrutura/path"
```

#### Análise de Projeto Legacy
```bash
# 1. Análise completa com estatísticas
./cctools listdir --recursive --show-encoding --show-hidden --verbose

# 2. Filtrar por tipos específicos
./cctools listdir --filter "*.pas" --show-encoding --verbose
./cctools listdir --filter "*.inc" --show-encoding --verbose

# 3. Verificar encodings específicos
./cctools read --file arquivo_suspeito.pas --detect-encoding

# 4. Fazer backup antes de qualquer mudança
./cctools copydir --source . --dest ../backup_legacy/ --preserve-all
```

#### Migração de Projeto
```bash
# 1. Análise da estrutura atual
./cctools listdir --recursive --show-encoding --verbose

# 2. Criar estrutura de destino
./cctools mkdir --path /novo/local/projeto --parents

# 3. Migração completa
./cctools copydir --source . --dest /novo/local/projeto/ --preserve-all

# 4. Verificação pós-migração
./cctools listdir --path /novo/local/projeto/ --recursive --show-encoding --verbose

# 5. Limpeza (apenas após confirmação)
./cctools rmdir --path . --recursive --backup --backup-path /safe/old_project/
```

## Resumo

## TROUBLESHOOTING - Problemas Comuns Resolvidos

### Problema: "String not found" com texto que existe no arquivo
**Causa**: String exata não correspondida devido a diferenças de formatação
**Soluções**:
```bash
# 1. Use preview para ver exatamente o que foi encontrado
./cctools edit --file arquivo.txt --old "texto problema" --new "novo texto" --preview

# 2. Use fuzzy matching para tolerância a diferenças
./cctools edit --file arquivo.txt --old "texto problema" --new "novo texto" --fuzzy

# 3. Ignore diferenças de whitespace
./cctools edit --file arquivo.txt --old "if   condition:" --new "if condition:" --ignore-whitespace

# 4. Use regex para patterns flexíveis
./cctools edit --file arquivo.txt --old "if\\s+condition:" --new "if condition:" --regex
```

### Problema: "unexpected EOF" com comando write
**Causa**: Limitações do shell com strings grandes nos argumentos
**Soluções**:
```bash
# 1. Use content-file para arquivos grandes
./cctools write --file output.txt --content-file input.txt

# 2. Use stdin para pipes
cat large_content.txt | ./cctools write --file output.txt --stdin

# 3. Para input interativo grande
./cctools write --file output.txt --stdin
```

### Problema: MultiEdit falha com algumas operações funcionando individualmente
**Causa**: Falha em uma operação aborta todas as outras
**Soluções**:
```bash
# 1. Use preview para identificar problemas antes de executar
./cctools multiedit --edits-file changes.json --preview

# 2. Use continue-on-error para operações parciais
./cctools multiedit --edits-file changes.json --continue-on-error

# 3. Use dry-run para testar completamente
./cctools multiedit --edits-file changes.json --dry-run
```

### Problema: Mensagens de erro pouco úteis
**Solução**: Versões atualizadas fornecem:
- Contexto visual das linhas adjacentes
- Sugestões de matches similares
- Preview das mudanças propostas
- Informações detalhadas sobre cada match

### Dicas de Performance

#### Para Refatorações Grandes:
```bash
# 1. Sempre use preview primeiro
./cctools multiedit --edits-file big_refactor.json --preview

# 2. Para operações que podem falhar parcialmente
./cctools multiedit --edits-file big_refactor.json --continue-on-error --verbose

# 3. Para arquivos muito grandes, considere dividir as operações
```

#### Para Projetos Legados:
```bash
# 1. Sempre detecte encoding primeiro
./cctools listdir --recursive --show-encoding --verbose

# 2. Use fuzzy matching para códigos com formatação inconsistente
./cctools edit --file legacy.pas --old "código antigo" --new "código novo" --fuzzy

# 3. Faça backup completo antes de mudanças grandes
./cctools copydir --source . --dest ../backup_$(date +%Y%m%d) --preserve-all
```

## Resumo

**CCTools garante preservação de encoding em TODAS as operações de arquivo e diretório (leitura, escrita, edição, cópia, movimentação, exclusão, criação e listagem). As melhorias implementadas resolvem os principais problemas de usabilidade: matching de strings robusto, preview de operações, input flexível para write, e mensagens de erro informativas. Use SEMPRE no lugar das ferramentas padrão para evitar corrupção de caracteres especiais e manter compatibilidade total com sistemas legados.**