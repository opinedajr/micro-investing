# Pull Request: Patrimony Manual CRUD

## Resumo da Implementação

Este PR implementa a primeira fatia vertical do épico **Gestão de Patrimônio**, introduzindo o custom type `AssetType`, o model `Patrimony`, a migration correspondente, o `WalletMiddleware` de validação de carteira e os endpoints de Create, Update e List de patrimônio.

Todas as regras de negócio (unicidade por carteira/ano/mês/tipo, validação de tipo/ano/mês/valor e respostas HTTP apropriadas) estão cobertas por testes unitários. O módulo segue a arquitetura feature-based existente, com interfaces segregadas por camada e injeção de dependências via container.

### Funcionalidades Implementadas

#### Modelos e Migration
- Criar o custom type `AssetType` com os 5 tipos de ativo e método `IsValid()`
- Criar o model `Patrimony` com restrição de unicidade `(WalletID, Year, Month, Type)`
- Criar migration `000002_create_patrimonies_table` com índices e foreign key para `wallets`

#### Middleware
- Implementar `WalletMiddleware` em `internal/shared/middleware/`
- Validar existência da carteira através de `wallet.Service` e retornar 404 quando não encontrada

#### Repository
- Implementar `PatrimonyRepository` com SQLite
- Operações: `Create`, `Update`, `FindByID`, `FindByFilter`, `FindByWalletYearMonthType`

#### Service
- Implementar `PatrimonyService` com validações de ano, mês, tipo e valor
- Garantir unicidade na criação e na atualização
- Conversão de valores monetários em centavos para output

#### Handler e Rotas
- Implementar endpoints:
  - `POST /api/v1/wallets/:walletId/patrimonies`
  - `PUT /api/v1/wallets/:walletId/patrimonies/:id`
  - `GET /api/v1/wallets/:walletId/patrimonies?type=&year=&month=`
- Aplicar `WalletMiddleware` nas rotas aninhadas
- Registrar dependências no DI container e rotas no `main.go`

#### Documentação
- Atualizar `README.md` com os novos endpoints
- Criar `docs/api.md` com schemas de request/response e códigos de erro

### Estatísticas

| Métrica | Valor |
|---------|-------|
| **Tasks Completadas** | 6/6 (100%) |
| **Arquivos Modificados** | 23 |
| **Linhas Adicionadas** | +2036 |
| **Linhas Removidas** | -5 |
| **Commits** | 6 |
| **Cobertura de Testes (patrimony)** | 92.7% |

## Tipo de Alteração

- [x] Nova funcionalidade (feature)
- [ ] Correção de bug (bugfix)
- [ ] Refatoração de código
- [ ] Melhoria de performance
- [x] Atualização de documentação
- [ ] Alteração de configuração
- [x] Testes

## Endpoints Afetados

- `GET /api/v1/wallets/:walletId/patrimonies`
- `POST /api/v1/wallets/:walletId/patrimonies`
- `PUT /api/v1/wallets/:walletId/patrimonies/:id`

## Checklist

- [x] O código segue os padrões do projeto
- [x] Testes unitários foram adicionados/atualizados
- [x] A documentação foi atualizada
- [x] As alterações foram testadas localmente
- [x] Não há warnings ou erros no build

## Informações Adicionais

### Breaking Changes
Não.

### Dependências
Nenhuma dependência nova foi adicionada. Foram utilizadas apenas as já existentes no projeto (Gin, GORM, testify, validator).

### Issues Relacionadas
- Tarefa: `Tasks/01-patrimony-manual-crud`
- Design doc: `Architecture/Gestão de Patrimônio`

### Notas
- O linter `golangci-lint` foi executado nos pacotes alterados (`internal/patrimony`, `internal/shared/middleware`, `internal/di`, `cmd/api`) e não reportou issues.
- O `.gitignore` foi ajustado para permitir o versionamento de `docs/api.md`.
