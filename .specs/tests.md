## QA Test Report: INV-19 — 04 — CLI cmd/quotation + re-consolidação de positions + target make

**Verdict:** APROVADO PARA MERGE

### Environment

- **PR:** #28 (`feature/inv-19-cli-quotation-sync` → `main`)
- **Workspace:** `/home/pineda/projects/micro-investing/.workspaces/inv-19-cli-quotation-sync`
- **Task:** INV-19 — 04 — CLI cmd/quotation + re-consolidação de positions + target make
- **Go version:** 1.23
- **Database:** SQLite (`DB_DRIVER=sqlite`, `DB_NAME=data/micro_investing.db`)

### Decisão sobre criação de testes E2E

**Nenhum teste E2E novo foi criado.**

Justificativa:
- A task **não adiciona superfície HTTP** (handler/routes de quotation ficam fora de escopo, conforme `spec.md`).
- O `spec.md` explicitamente estabelece "CLI sem teste: orquestrador fino, mesmo tratamento do binário de seed".
- Os comportamentos de negócio do service (`SyncCurrentPrices`), repository (`UpsertCurrentPrices`) e adapter Brapi já possuem cobertura unitária/integrada (INV-16/17/18) e continuam passando.
- A regressão E2E foi feita executando a suite existente e validando o comportamento do CLI manualmente contra servidor fake local.

### Test Cases

#### TC1: Build e vet de todos os pacotes
- **Command:** `go build ./... && go vet ./...`
- **Expected:** Compilação e análise estática bem-sucedidas, sem erros.
- **Actual:** `go build ./...` e `go vet ./...` concluíram sem erros.
- **Status:** PASS

#### TC2: Testes unitários Go
- **Command:** `go test ./...`
- **Expected:** Todos os pacotes passam sem falhas.
- **Actual:** Todos os pacotes passaram (`ok` para todos os pacotes com testes; `[no test files]` para os demais).
- **Status:** PASS

#### TC3: Regressão E2E da suite existente
- **Command:** `go test -tags=integration -v ./test/e2e/...`
- **Expected:** Suite `TestE2ESuite` completa com todos os cenários PASS.
- **Actual:** Todos os cenários da suite passaram.
- **Status:** PASS

#### TC4: Build do binário CLI
- **Command:** `go build -o /tmp/quotation-test ./cmd/quotation`
- **Expected:** Binário gerado com sucesso.
- **Actual:** Binário `/tmp/quotation-test` gerado (~27 MB).
- **Status:** PASS

#### TC5: Execução sem `BRAPI_API_KEY`
- **Command:** `BRAPI_API_KEY="" go run ./cmd/quotation`
- **Expected:** Exit code 1, mensagem clara no stderr, sem stack trace/panic exposto.
- **Actual:** `failed to sync current prices: failed to create brapi client: brapi client requires BRAPI_API_KEY to be set` + `exit status 1`.
- **Status:** PASS

#### TC6: Execução via `make sync-prices` sem credencial
- **Command:** `make sync-prices BRAPI_API_KEY=""`
- **Expected:** Binário falha com exit 1 (o make pode refletir o erro como exit 2, mas o binário retorna 1).
- **Actual:** CLI emitiu mensagem de erro e `exit status 1`; make retornou `Error 1` (exit 2 do processo make).
- **Status:** PASS

#### TC7: Target `make sync-prices` expande corretamente
- **Command:** `make -n sync-prices` e `make -n sync-prices ARGS="-tickers=PETR4,VALE3"`
- **Expected:** Target executa `go run ./cmd/quotation $(ARGS)`.
- **Actual:** `go run ./cmd/quotation` e `go run ./cmd/quotation -tickers=PETR4,VALE3`.
- **Status:** PASS

#### TC8: CLI com flag `-tickers=PETR4,VALE3` e servidor fake
- **Command:** Executado via wrapper local com `httptest` fake retornando PETR4=41.18 e VALE3=67.50.
- **Expected:** Exit 0, sumário com Updated>0, Skipped=0, Failed=0, wallets consolidadas.
- **Actual:** `current prices sync summary: updated=4 skipped=0 failed=0 wallets=1 consolidation_failures=0` + `exit code: 0`.
- **Status:** PASS (com observação sobre contagem de `updated`, ver seção de observações)

#### TC9: Re-consolidação de positions após sync
- **Command:** Criada wallet + posição VALE3 (quantity=50, average_price=6000) sem preço no cache; executado CLI com servidor fake full; consultada posição via API.
- **Expected:** `current_price` atualizado para 6750 centavos, `balance` recalculado para 337500 centavos, `variation_value` refletindo o ganho.
- **Actual:** `current_price: 6750`, `balance: 337500`, `variation_value: 37500`.
- **Status:** PASS

#### TC10: Cenário "todos skipped" (Brapi responde 200 com results vazio)
- **Command:** Executado CLI via wrapper local com servidor fake em modo `empty`.
- **Expected:** Exit 0, Updated=0, Skipped=N, Failed=0, re-consolidação normal (conforme spec: falha total exige ao menos um lote falhado).
- **Actual:** `current prices sync summary: updated=0 skipped=29 failed=0 wallets=1 consolidation_failures=0` + `exit code: 0`.
- **Status:** PASS

### Summary
- Total: 10 cenários de teste/regressão
- Passed: 10
- Failed: 0

### Observações

1. **Contagem de `updated` quando o provider retorna quotes extras:** No teste com `-tickers=PETR4,VALE3` e servidor fake que sempre retornava ambos os tickers independentemente do batch, o service reportou `updated=4` (2 batches × 2 quotes). Isso ocorre porque o service concatena todos os quotes recebidos em todos os batches e não deduplica/filtra por batch solicitado. Na integração real com a Brapi v2 (e batch size 1 do plano gratuito), cada chamada retorna apenas o ticker solicitado, então a contagem permanece correta. Não considerado bloqueante, mas pode ser endurecido futuramente para ignorar tickers fora do batch atual.
2. **Exit code via `make`:** Quando a credencial está ausente, o binário retorna exit 1 conforme AC. O wrapping `make` reflete isso como exit 2 (erro do make). Em um crontab que chame o binário diretamente, o comportamento é exatamente exit 1.
3. **Ressalvas do @eris:**
   - **(1) `.specs/tests.md` sobrescrito:** este relatório substitui o conteúdo conforme workflow do agente @temis; se a política do projeto evoluir para manter histórico por task, recomenda-se renomear ou arquivar os relatórios anteriores em vez de sobrescrever.
   - **(2) Inconsistência de caminho no Makefile:** `sync-prices` usa `./cmd/quotation` enquanto `seed-stock` usa `cmd/seed/main.go`. Trata-se de estilo/cosmético; ambos funcionam. Não bloqueia merge.
   - **(3) Caso "todos skipped":** comportamento validado e alinhado com o spec — exit 0, Updated=0, Skipped=N, Failed=0, re-consolidação executada. Não recomendo mudança antes do merge; se desejado, pode-se discutir se "nenhum preço obtido" deveria ser considerado falha total independentemente de `failed`, mas isso contraria a redação atual do spec.
4. **Artefato não versionado na raiz:** existe um binário `quotation` (~27 MB) na raiz do workspace marcado como untracked. Não faz parte desta QA; recomenda-se removê-lo ou adicionar regra ao `.gitignore` para evitar commit acidental.

### Cleanup
- Artifacts removed: YES
  - Banco SQLite de QA (`data/micro_investing.db`) removido.
  - Servidor API temporário encerrado.
  - Scripts/utilitários temporários em `/tmp` descartados.
  - Nenhum arquivo de código ou teste do projeto foi alterado.
