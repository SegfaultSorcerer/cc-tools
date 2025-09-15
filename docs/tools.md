# Ferramentas de Edição de Arquivos

Este documento descreve as ferramentas disponíveis para edição de arquivos no Claude Code, incluindo seus parâmetros, funcionamento e características técnicas.

## Visão Geral

O Claude Code possui 4 ferramentas principais para edição de arquivos:

1. **Edit** - Edição simples com substituição de texto
2. **MultiEdit** - Múltiplas edições em um único arquivo
3. **Write** - Criação/sobrescrita completa de arquivos
4. **NotebookEdit** - Edição específica para notebooks Jupyter

## 1. Ferramenta Edit

### Descrição
Realiza substituições exatas de strings em arquivos existentes.

### Parâmetros
- `file_path` (obrigatório): Caminho absoluto do arquivo a ser modificado
- `old_string` (obrigatório): Texto a ser substituído
- `new_string` (obrigatório): Texto de substituição
- `replace_all` (opcional, padrão: false): Substitui todas as ocorrências

### Como Funciona
- Busca pela string exata em `old_string` no arquivo
- Substitui por `new_string`
- A string deve ser única no arquivo (exceto se `replace_all=true`)
- Preserva exatamente a indentação e formatação original

### Requisitos
- O arquivo deve ser lido com a ferramenta `Read` antes da edição
- Caminhos devem ser absolutos (iniciar com `/`)
- A string `old_string` deve existir exatamente como especificada

### Exemplo
```json
{
  "file_path": "/home/user/projeto/app.js",
  "old_string": "const port = 3000;",
  "new_string": "const port = process.env.PORT || 3000;"
}
```

## 2. Ferramenta MultiEdit

### Descrição
Permite múltiplas edições em um único arquivo de forma atômica.

### Parâmetros
- `file_path` (obrigatório): Caminho absoluto do arquivo
- `edits` (obrigatório): Array de objetos de edição, cada um contendo:
  - `old_string`: Texto a ser substituído
  - `new_string`: Texto de substituição
  - `replace_all` (opcional): Substitui todas as ocorrências

### Como Funciona
- Aplica todas as edições sequencialmente
- Se qualquer edição falhar, nenhuma é aplicada (operação atômica)
- Cada edição opera no resultado da edição anterior
- Ideal para múltiplas mudanças no mesmo arquivo

### Requisitos
- Arquivo deve ser lido previamente
- Todas as edições devem ser válidas para a operação suceder
- Planejar cuidadosamente para evitar conflitos entre edições sequenciais

### Exemplo
```json
{
  "file_path": "/home/user/projeto/config.js",
  "edits": [
    {
      "old_string": "debug: false",
      "new_string": "debug: true"
    },
    {
      "old_string": "port: 3000",
      "new_string": "port: 8080"
    }
  ]
}
```

## 3. Ferramenta Write

### Descrição
Cria novos arquivos ou sobrescreve completamente arquivos existentes.

### Parâmetros
- `file_path` (obrigatório): Caminho absoluto do arquivo
- `content` (obrigatório): Conteúdo completo do arquivo

### Como Funciona
- Sobrescreve completamente o arquivo se existir
- Cria novo arquivo se não existir
- Substitui todo o conteúdo anterior

### Requisitos
- Para arquivos existentes, deve ser lido previamente
- Caminhos devem ser absolutos
- Preferir edição de arquivos existentes ao invés de criação

### Exemplo
```json
{
  "file_path": "/home/user/projeto/novo-arquivo.js",
  "content": "console.log('Novo arquivo criado!');\nmodule.exports = {};"
}
```

## 4. Ferramenta NotebookEdit

### Descrição
Edição específica para arquivos Jupyter Notebook (.ipynb).

### Parâmetros
- `notebook_path` (obrigatório): Caminho absoluto do notebook
- `new_source` (obrigatório): Novo conteúdo da célula
- `cell_id` (opcional): ID da célula a ser editada
- `cell_type` (opcional): Tipo da célula ("code" ou "markdown")
- `edit_mode` (opcional): Modo de edição ("replace", "insert", "delete")

### Como Funciona
- `replace`: Substitui conteúdo da célula existente
- `insert`: Adiciona nova célula
- `delete`: Remove célula existente
- Mantém estrutura JSON do notebook

## Codificação de Arquivos

### Codificação Padrão
- **UTF-8**: Todos os arquivos são salvos em UTF-8 por padrão
- Suporte completo a caracteres Unicode
- Compatível com caracteres especiais, acentos e emojis

### Comportamento de Persistência
- Arquivos são salvos imediatamente após cada operação
- Não há cache ou buffer temporário
- Mudanças são persistidas diretamente no sistema de arquivos
- Preserva permissões de arquivo existentes

### Características Técnicas
- Line endings preservados conforme sistema operacional
- Indentação (tabs/espaços) preservada exatamente
- Caracteres de controle mantidos quando presentes no original

## Boas Práticas

### Antes de Editar
1. Sempre usar `Read` para examinar o arquivo primeiro
2. Verificar convenções de código existentes
3. Entender a estrutura do projeto

### Durante a Edição
1. Usar caminhos absolutos sempre
2. Preservar indentação e formatação originais
3. Testar edições complexas com `MultiEdit` quando apropriado

### Após Edição
1. Verificar se mudanças foram aplicadas corretamente
2. Executar testes quando disponíveis
3. Verificar lint/typecheck se configurados no projeto

## Limitações Importantes

- **Strings devem ser exatas**: Espaços e indentação devem coincidir perfeitamente
- **Caminhos absolutos obrigatórios**: Caminhos relativos não são aceitos
- **Leitura prévia obrigatória**: Arquivos existentes devem ser lidos antes da edição
- **Operações atômicas**: MultiEdit falha completamente se qualquer edição individual falhar
- **Sem desfazer**: Não há funcionalidade de undo integrada