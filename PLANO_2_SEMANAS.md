# Plano de 2 semanas — API Go + Loja Next.js

Escopo: loja com **cadastro de usuário**, vendedor publicando produto, comprador comprando,
pedido processado por fila e worker assíncrono.

~14 dias corridos, 3h/dia ≈ **42h de orçamento**. Versão longa e comentada em `PLANO_DETALHADO.md`.

---

## Fases

### Fase 0 — Fundação (dias 1-2, ~5h)

Consertar o que está quebrado. Sem isso o resto não funciona.

- Trocar `pgx.Connect` por `pgxpool.Pool` em `cmd/main.go`. `MaxConns` vindo de env. O `DBTX` gerado pelo sqlc só exige `Exec`/`Query`/`QueryRow`, que o pool tem — nada gerado precisa mudar.
- Baixar estoque de forma atômica: `UPDATE products SET quantity = quantity - $2 WHERE id = $1 AND quantity >= $2 RETURNING *`. Zero linhas = sem estoque.
- Em `PlaceOrder`, ler o produto com `qtx` e não com `svc.repo` (hoje a leitura acontece fora da transação).
- Checar o erro de `tx.Commit(ctx)`.
- Graceful shutdown com `signal.NotifyContext` + `srv.Shutdown`. Corrigir `IdleTimeout` de 1s para 60s.
- Pacote `internal/httperr`: erro de domínio vira status + `{"error":"...","request_id":"..."}`. Parar de repassar `err.Error()` ao cliente.
- CORS (`go-chi/cors`), origem por env, com `AllowCredentials` (o cookie de sessão depende disso).

### Fase 1 — Modelo de dados (dias 2-3, ~4h)

Alinhar banco e front, que hoje falam coisas diferentes, e criar a tabela que hoje não existe.

- Migration `00003` — **usuários**:
  ```sql
  CREATE TABLE users (
      id BIGSERIAL PRIMARY KEY,
      email TEXT NOT NULL UNIQUE,
      password_hash TEXT NOT NULL,
      name TEXT NOT NULL,
      role TEXT NOT NULL DEFAULT 'comprador',  -- 'comprador' | 'vendedor'
      created_at TIMESTAMPTZ NOT NULL DEFAULT now()
  );
  ```
  Uma tabela só. Sem tabela `sellers` separada, sem tabela de permissões.
- Migration `00004` — `products` ganha `description`, `image_url`, `slug`, `category`, `active BOOLEAN DEFAULT true` e **`seller_id BIGINT NOT NULL REFERENCES users(id)`**. Índices em `slug`, `category`, `seller_id`.
- Migration `00005` — `orders` ganha `status` e `total_cents`, e `customer_id` finalmente vira `REFERENCES users(id)`. FK em `order_items.product_id`, índice em `order_items.order_id`.
- Migration `00006` — seed: dois usuários (um vendedor, um comprador) e os 7 produtos que hoje estão em `data/constants/produtos.ts`, com preço convertido para centavos e `seller_id` apontando pro vendedor.
- DTOs em `internal/products/dto.go`, `internal/orders/dto.go`, `internal/users/dto.go`. Handlers param de devolver `repo.Product` cru. **`password_hash` nunca entra em DTO.**
- `orders.NewService` passa a receber `repo.Querier` (interface) em vez de `*repo.Queries` — sem isso não dá pra testar com mock.

### Fase 2 — Autenticação (dias 3-4, ~3h)

- Pacote `internal/auth`: hash com `bcrypt` (`golang.org/x/crypto/bcrypt`), custo 12. Nunca inventar hash.
- JWT HS256 (`golang-jwt/jwt/v5`), segredo por env, expiração de 24h. Claims: `sub` (user id) e `role`.
- Token vai em cookie `httpOnly`, `SameSite=Lax`, `Secure` em produção. **Sem refresh token** — decisão de escopo, declarada no README.
- Middleware `RequireAuth` (valida token, injeta claims no `context`) e `RequireRole("vendedor")`.
- Endpoints:
  ```
  POST /auth/register     cria usuário; role no corpo
  POST /auth/login        seta o cookie
  POST /auth/logout       limpa o cookie
  GET  /me                dados do usuário logado
  ```
- Login com email inexistente e senha errada devem devolver **a mesma** resposta e levar o mesmo tempo. Vazar qual dos dois falhou entrega quais emails existem na base.

### Fase 3 — Endpoints (dias 4-6, ~6h)

De 4 para 15 endpoints.

**Público**
```
GET  /products?limit=&cursor=&q=&category=   paginação por cursor + busca ILIKE; filtra active=true
GET  /products/{id}                          (renomear de /product/{id}); 404 vs 500 corretos
```

**Vendedor** (`RequireAuth` + `RequireRole("vendedor")`)
```
POST   /products             cria; seller_id vem do token, não do corpo
PATCH  /products/{id}        só o dono
DELETE /products/{id}        soft delete (active=false), só o dono
GET    /me/products          catálogo do vendedor, incluindo inativos
```
- Autorização é uma comparação: `produto.seller_id == claims.sub`. Não construir framework de permissão.
- Imagem é campo `image_url` de texto. **Sem upload** — S3/MinIO, presigned URL e validação de MIME são um fim de semana e não mostram Go nenhum.

**Comprador** (`RequireAuth`)
```
POST /orders          customer_id vem do token; deixa de ser aceito no corpo
GET  /orders/{id}      só o dono do pedido
GET  /me/orders        listagem paginada
```
> `/me/orders` em vez de `/customers/{id}/orders`: com o id na URL você precisa checar se o solicitante é aquele cliente em toda chamada, e um esquecimento vira leitura do pedido alheio. Tirando o id da URL o problema deixa de existir.

**Operacional**
```
GET /healthz    liveness — 200 se o processo vive
GET /readyz     readiness — pinga Postgres e Redis
```
Remover o `/health` que devolve texto puro.

### Fase 4 — Fila e worker (dias 6-8, ~5h)

- Redis no `docker-compose` + `go-redis/v9`, cliente pendurado em `application.rdb`.
- `POST /orders` faz `XADD orders.created` após o commit.
- `cmd/worker/main.go`: binário separado, mesmo módulo. `XREADGROUP` com consumer group, simula pagamento (sleep 2-5s + falha em ~10%), atualiza status, `XACK`. Falha demais vai pra stream `orders.dead`.
- Status do pedido passa a ser consultado via `GET /orders/{id}` / `GET /me/orders` — sem push em tempo real nesta entrega (ver "Cortado para o cadastro caber").
- Anotar no README: publicar depois do commit não é atômico. A solução é *transactional outbox*, e ficou fora de escopo. Declarar isso vale mais que implementar mal.

### Fase 5 — Frontend (dias 8-11, ~8h)

**Correções do que existe**
- Corrigir a mutação de estado no `ContextoCarrinho` (hoje faz `i.quantidade -= 1` no objeto original). Usar spread. `useMemo` no `value` do Provider, `useCallback` nas funções.
- Persistir carrinho em `localStorage` (ler dentro de `useEffect`, não no `useState` inicial).
- Tirar `'use client'` de `app/(loja)/page.tsx` — a home vira Server Component e busca da API. Só `CartaoProduto` continua client.
- `loading.tsx`, `error.tsx`, `not-found.tsx`.
- Faxina: `lang="pt-BR"`, `w-[1200px]` → `max-w-[1200px] w-full px-4`, `images.domains` → `remotePatterns`, formatar centavos com `Intl.NumberFormat`.
- `data/services/api.ts` centralizando os `fetch`. Apagar `data/constants/produtos.ts`.

**Contas**
- `app/(auth)/cadastrar/page.tsx` e `app/(auth)/entrar/page.tsx`, com escolha de perfil no cadastro.
- `middleware.ts` protegendo `/painel/*` e `/checkout`.
- Contexto de sessão alimentado por `GET /me`.
- Cabeçalho muda conforme o estado: visitante vê "Entrar"; comprador vê carrinho e "Meus pedidos"; vendedor vê "Painel".

**Vendedor**
- `app/painel/page.tsx`: tabela dos produtos + formulário de criar/editar. **Tailwind cru, sem polimento** — é vaga de Go, os endpoints valem mais que a tela.
- Após criar produto, `revalidatePath('/')` invalida o cache da home (que é Server Component).

**Comprador**
- Checkout chamando `POST /orders` via Server Action.
- `pedido/[id]`: mostra o status atual do pedido (`pendente → pago → separando → enviado`) via `GET /orders/{id}`. Sem atualização automática — o comprador atualiza a página pra ver o novo status. É a limitação declarada no README.

### Fase 6 — Testes e entrega (dias 11-14, ~6h)

- `internal/orders/service_test.go` table-driven com mock: customer zerado, itens vazios, produto inexistente, sem estoque, caminho feliz.
- **Teste de concorrência** com `testcontainers-go`: Postgres real, estoque 10, 50 goroutines pedindo 1 cada. Exatamente 10 passam, estoque final 0, nunca negativo. É o artefato mais forte do repositório.
- Testes de autorização: vendedor A não edita produto do vendedor B (403); comprador não cria produto (403); pedido de outro usuário dá 404.
- `httptest` nos handlers: status e forma do JSON.
- `Dockerfile` multi-stage, `CGO_ENABLED=0`, usuário não-root. `docker-compose` com `postgres`, `redis`, `api`, `worker`, `migrate`.
- `Makefile`: `up`, `migrate`, `sqlc`, `test`, `lint`.
- README com diagrama do fluxo, "Decisões e trade-offs" e "Limitações conhecidas".
- GitHub Actions: `go vet`, `golangci-lint`, `go test ./...`, `npm run build`.

### Cortado para o cadastro caber

Não é esquecimento — é troca consciente. Citar como limitação no README:

- **Status em tempo real (WebSocket)** — o comprador consulta o status via `GET /orders/{id}`, sem push. Fan-out com Redis Pub/Sub multi-réplica é peça de infra adicional (auth no handshake, reconexão, hub por conexão) que não cabe no orçamento sem cortar o teste de concorrência ou os testes de autorização, que valem mais numa entrevista de backend.
- **`/metrics` (Prometheus)** — instrumentação de `http_requests_total`, `http_request_duration_seconds` e `pool.Stat()` fica como próximo passo declarado, não implementado nesta entrega.
- `POST /orders/{id}/cancel` — transação compensatória é bom papo, mas o `UPDATE` atômico já cobre o assunto estoque.
- Header `Idempotency-Key` — o que mais dói perder. Vale citar de viva voz: *"sei que clique duplo gera pedido duplicado; a solução é chave de idempotência no servidor"*.
- Página de detalhe do produto com `generateMetadata` — o painel do vendedor já demonstra rota dinâmica.

### Se sobrar tempo, nesta ordem

Rate limit em `/auth/login` e `POST /orders` (`httprate`) → deploy no Fly.io ou Railway → Idempotency-Key de volta → status em tempo real via WebSocket + Redis Pub/Sub → `/metrics` (Prometheus).

---

## O que muda

| | Hoje | Depois |
|---|---|---|
| **Usuários** | Não existem. `orders.customer_id` é um número inventado no JSON, sem FK e sem tabela | Tabela `users` com papel, senha em bcrypt, FK de verdade |
| **Produtos** | 7 registros fixos em `data/constants/produtos.ts`, ninguém cadastra | Vendedor publica, edita e desativa pelo painel; comprador vê a lista viva |
| **Autorização** | Nenhuma. Qualquer um faz qualquer coisa | JWT em cookie httpOnly; dono do produto e dono do pedido verificados |
| **Identidade no pedido** | Cliente manda `customer_id` que quiser | Vem do token; `/me/orders` não tem id na URL pra ninguém trocar |
| **Conexão com o banco** | Uma `pgx.Conn`, não concurrency-safe — a API inteira serializa | Pool com `MaxConns` configurável |
| **Estoque** | Nunca é baixado. Dá pra comprar infinito | `UPDATE` condicional atômico, com teste de 50 goroutines provando |
| **Transação** | A leitura do produto acontece fora dela | Tudo dentro de `qtx`, commit com erro verificado |
| **Erros** | `err.Error()` do Postgres vaza pro cliente | JSON estruturado com `request_id`; erro real só no log |
| **Endpoints** | 4 | 15, com paginação por cursor, papéis e 404 correto |
| **Contrato** | Handler devolve `repo.Product` — schema acoplado ao HTTP | DTOs; mudar coluna não quebra cliente |
| **Front ↔ Back** | Não conversam. Sem CORS, produtos hardcoded no front | Next como BFF nas chamadas HTTP |
| **Preço** | Float em reais no front, centavos no banco | Centavos como inteiro em toda a stack |
| **Processamento** | Tudo síncrono dentro do request | Redis Stream + worker separado, com retry e DLQ |
| **Status do pedido** | Não existe | Consultado via `GET /orders/{id}`; tempo real fora de escopo (ver "Cortado") |
| **Carrinho** | Muta estado, some no F5, é beco sem saída | Imutável, persiste, termina em checkout e pedido rastreável |
| **Next.js** | Tudo `'use client'`, App Router desperdiçado | Home em RSC, `middleware.ts` protegendo rota, `revalidatePath` após publicar |
| **Testes** | Zero | Unitários com mock + concorrência com Postgres real + autorização + handlers |
| **Deploy** | `go run` local, só o Postgres em container | `docker compose up` sobe tudo; imagem distroless, shutdown gracioso |
| **Operação** | `/health` devolvendo `all good` em texto | `/healthz`, `/readyz` separados (`/metrics` fica pro "se sobrar tempo") |

Em uma frase: hoje são dois projetos que não se falam, sem usuário nenhum, vendendo produto fixo que nunca sai do estoque. Depois é um fluxo fechado — vendedor cadastra, comprador se cadastra e compra, a transação baixa estoque de verdade, o evento vai pra fila e o worker processa de forma assíncrona e resiliente.

---

## Por que não Kafka

O Kafka existe pra reter log particionado com retenção longa, ordenação por partição e throughput de centenas de milhares de mensagens por segundo. Você tem um fluxo, um tipo de evento e volume de demonstração.

Redis Streams entrega o que a conversa exige — produtor, consumer group, ack, retry, dead letter — em um container e ~60 linhas. Kafka pede ZooKeeper ou KRaft, configuração de partição, e vira metade do prazo em infraestrutura em vez de código.

E a resposta na entrevista é melhor: *"escolhi Redis Streams porque o volume não justifica o custo operacional do Kafka, e consumer group já me dá at-least-once com distribuição entre réplicas"* soa como julgamento. Kafka num projeto de portfólio soa como currículo.

## Por que não Kubernetes

Manifesto que você nunca depurou em produção é transparente numa entrevista. Basta uma pergunta sobre `livenessProbe` derrubando pod durante GC, ou sobre `PodDisruptionBudget`, e o teatro cai — e aí o dano é maior do que se você nunca tivesse citado.

O custo também não fecha: Kubernetes de verdade consome o prazo inteiro em YAML, e sobra zero pra cadastro, fila e testes — que é onde está o valor.

O ponto central: **você não precisa rodar K8s pra conversar sobre K8s.** `/healthz` e `/readyz` separados, config por env, imagem distroless, `SIGTERM` tratado são exatamente o que um orquestrador exige da aplicação. Isso é pouco trabalho e te deixa dizer, com verdade: *"a aplicação está pronta pra orquestrador — liveness e readiness são endpoints diferentes porque se o banco cair eu quero sair do load balancer, não ser reiniciado. Não rodei em cluster, rodei em Compose."*

Resposta honesta e tecnicamente correta. Vale mais que um `deployment.yaml` copiado.

Se quiser Kubernetes, WebSocket ou `/metrics` depois, com `/readyz` já pronto e a fila já desacoplada em worker separado, isso vira incremento de um fim de semana, não reescrita.