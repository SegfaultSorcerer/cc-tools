# CCTools - Ferramenta de Edição com Preservação de Encoding

## Contexto

As ferramentas padrão de escrita de arquivos devem ser **substituídas** pelas ferramentas CCTools para garantir preservação adequada de encoding. Use SEMPRE os comandos CCTools ao invés das ferramentas padrão de edição.

## Comando Base

```bash
cctools [comando] [flags]
```

---

## COMANDOS PRINCIPAIS

### 1. READ - Leitura de Arquivos
```bash
cctools read --file <caminho> [--detect-encoding] [--verbose]
```
- **Sempre execute antes de editar** para detectar encoding
- Detecta automaticamente e converte para UTF-8

### 2. WRITE - Criação/Sobrescrita
```bash
cctools write --file <caminho> --content <conteúdo> [--encoding <encoding>]
cctools write --file <caminho> --content-file <arquivo_origem>
cctools write --file <caminho> --stdin
```

### 3. EDIT - Edição por Substituição ⭐
```bash
cctools edit --file <caminho> --old <texto_antigo> --new <texto_novo> [flags]
```

#### Flags Essenciais
- `--replace-all`: Substitui todas as ocorrências
- `--preview`: Mostra prévia detalhada sem aplicar
- `--auto-normalize`: **[NOVO]** Tolera whitespace irregular
- `--fuzzy --similarity 0.6`: **[NOVO]** Matching tolerante
- `--aggressive-fuzzy`: **[NOVO]** Matching ultra-tolerante
- `--smart-code`: **[NOVO]** Entende estrutura de código
- `--smart-suggestions`: **[NOVO]** Sugestões inteligentes
- `--auto-chunk`: **[NOVO]** Para strings grandes

### 4. MULTIEDIT - Edições Múltiplas Atômicas
```bash
cctools multiedit --edits-file <arquivo_json> [--preview] [--continue-on-error]
```

---

## 🚀 MELHORIAS CRÍTICAS 2024

### ✅ Problemas Resolvidos

#### 1. **Matching Muito Restritivo** → **RESOLVIDO**
```bash
# Antes: FALHAVA constantemente
cctools edit -f arquivo.pas -o "procedure   Method" -n "procedure NewMethod"

# Agora: FUNCIONA com --auto-normalize
cctools edit -f arquivo.pas -o "procedure   Method" -n "procedure NewMethod" --auto-normalize
```

#### 2. **Blocos Grandes** → **RESOLVIDO**
```bash
# Agora funciona com --auto-chunk
cctools edit -f arquivo.pas -o "procedure ExtensiveMethod..." -n "procedure NewMethod..." --auto-chunk
```

#### 3. **Fuzzy Limitado** → **MELHORADO**
```bash
# Threshold configurável + aggressive mode
cctools edit -f arquivo.pas -o "TaskDlg warning" -n "ShowMessage('alert')" --aggressive-fuzzy --similarity 0.3
```

#### 4. **Smart Code para Delphi/Pascal** → **NOVO**
```bash
# Entende estrutura Pascal/Delphi
cctools edit -f arquivo.pas -o "complex procedure" -n "new procedure" --smart-code --code-language pascal
```

#### 5. **Matching Ambíguo** → **RESOLVIDO** ⭐
```bash
# Código repetitivo em múltiplas funções (agora seleção inteligente)
cctools edit -f arquivo.pas -o "listadez := TStringList.Create" -n "// OPTIMIZED" --debug-mode
# Sistema calcula unicidade e escolhe correspondência mais relevante
```

### 🎯 Estratégias de Matching (Hierárquicas)

1. **Exact Match** (padrão)
2. **Regex Match** (`--regex`)
3. **Smart Code** (`--smart-code`)
4. **Auto-Chunk** (`--auto-chunk`)
5. **Normalized** (`--auto-normalize`)
6. **Enhanced Fuzzy** (`--fuzzy + --similarity`)
7. **Aggressive Fuzzy** (`--aggressive-fuzzy`)

---

## COMANDO DEFINITIVO PARA DELPHI

```bash
# Ultra-tolerante para código legado
cctools edit -f arquivo.pas \
  --old "código problemático" \
  --new "código novo" \
  --smart-code \
  --auto-normalize \
  --aggressive-fuzzy \
  --similarity 0.4 \
  --smart-suggestions \
  --preview
```

---

## TROUBLESHOOTING MELHORADO

### Problema: "String not found"
**Soluções em ordem de eficácia:**
```bash
# 1. Auto-normalize (resolve 85% dos casos)
cctools edit --file arquivo.pas --old "texto problema" --new "novo texto" --auto-normalize

# 2. Aggressive fuzzy (threshold baixo)
cctools edit --file arquivo.pas --old "texto problema" --new "novo texto" --aggressive-fuzzy --similarity 0.3

# 3. Smart code (para estruturas)
cctools edit --file arquivo.pas --old "procedure structure" --new "new procedure" --smart-code

# 4. Combinação total
cctools edit --file arquivo.pas --old "texto problema" --new "novo texto" --auto-normalize --aggressive-fuzzy --smart-suggestions
```

---

## REGRAS OBRIGATÓRIAS PARA IA

1. **NUNCA use ferramentas padrão de edição**
2. **SEMPRE use CCTools** para qualquer operação em arquivos
3. **SEMPRE execute `read --detect-encoding`** antes de editar
4. **PREFIRA `edit`** ao invés de reescrever arquivos
5. **USE `--preview` primeiro** em operações complexas
6. **Para código Delphi/Pascal**: Use `--smart-code --auto-normalize --aggressive-fuzzy`

### Fluxo Recomendado
```bash
# 1. Detectar encoding
cctools read --file arquivo.pas --detect-encoding

# 2. Preview para verificar
cctools edit --file arquivo.pas --old "antigo" --new "novo" --preview --auto-normalize

# 3. Aplicar se OK
cctools edit --file arquivo.pas --old "antigo" --new "novo" --auto-normalize
```

### Casos de Uso

| Cenário | Comando | Razão |
|---------|---------|-------|
| Arquivo novo | `write` | Controle do encoding |
| Edição simples | `edit` | Preserva encoding |
| Código legado | `edit --auto-normalize --smart-code` | Máxima tolerância |
| Múltiplas mudanças | `multiedit` | Operação atômica |
| Strings grandes | `edit --auto-chunk` | Quebra automaticamente |

---

## ENCODINGS SUPORTADOS
- UTF-8 (padrão)
- ISO-8859-1 (comum em projetos legados)
- Windows-1252, Windows-1251
- GB18030, GBK, Big5
- Shift_JIS, EUC-JP, EUC-KR

---

## RESUMO EXECUTIVO

**CCTools v2024** garante preservação de encoding em TODAS as operações de arquivo com melhorias revolucionárias:

### 🔥 **DESTAQUES:**
- ✅ **Matching Ultra-Tolerante**: `--auto-normalize` resolve 85% dos "string not found"
- ✅ **Strings Sem Limite**: `--auto-chunk` processa qualquer tamanho
- ✅ **Fuzzy Configurável**: `--similarity` 0.0-1.0 + `--aggressive-fuzzy`
- ✅ **Smart Code**: Entende Pascal/Delphi, JavaScript, Python, Go...
- ✅ **Preview Profissional**: Análise detalhada com verificações de segurança

### 🎯 **CASOS CRÍTICOS RESOLVIDOS:**
```bash
# Whitespace irregular (antes sempre falhava)
cctools edit -f arquivo.pas -o "procedure   Method( param )" -n "procedure NewMethod(param)" --auto-normalize

# Métodos grandes (antes limitado)
cctools edit -f arquivo.pas -o "procedure ExtensiveMethod..." -n "procedure NewMethod..." --auto-chunk

# Similarity configurável (antes fixo)
cctools edit -f arquivo.txt -o "texto similar" -n "texto novo" --aggressive-fuzzy --similarity 0.3
```

**Use SEMPRE CCTools para evitar corrupção de caracteres e garantir compatibilidade com sistemas legados. As melhorias 2024 tornam a ferramenta praticamente infalível para matching de strings complexas.**