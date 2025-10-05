# 🚀 Quick Start Guide

Guia rápido para começar a usar o Rate Limiter em menos de 5 minutos!

## ⚡ Início Rápido (5 minutos)

### Passo 1: Clone o Repositório

```bash
git clone https://github.com/celiojsf/pos-challenge-rate-limiter.git
cd pos-challenge-rate-limiter
```

### Passo 2: Inicie com Docker

```bash
docker-compose up -d
```

### Passo 3: Teste!

```bash
# Teste básico
curl http://localhost:8080/

# Teste o rate limiter
for i in {1..15}; do
  echo "Request $i:"
  curl -s http://localhost:8080/api/test | jq
done
```

Pronto! Você deve ver as primeiras 10 requisições sendo aceitas e as seguintes sendo bloqueadas com erro 429.

## 📋 Verificações Importantes

### ✅ Containers Rodando

```bash
docker-compose ps
```

Você deve ver:
- `rate-limiter-redis` - healthy
- `rate-limiter-app` - running

### ✅ Logs da Aplicação

```bash
docker-compose logs app
```

Deve mostrar:
```
Connected to Redis successfully
Starting server on port 8080...
Rate Limiter Configuration:
  - IP Limit: 10 requests/second
  - Token Limit: 100 requests/second
  - Block Duration: 300 seconds
```

### ✅ Health Check

```bash
curl http://localhost:8080/health
```

Deve retornar: `{"status": "healthy"}`

## 🎯 Cenários de Teste Rápidos

### Cenário 1: Rate Limit por IP (30 segundos)

```bash
# Faça 12 requisições rapidamente
for i in {1..12}; do
  curl -s http://localhost:8080/api/test | jq -r '.message // .error'
  sleep 0.1
done
```

**Resultado esperado:**
- 10 sucessos
- 2 bloqueios

### Cenário 2: Rate Limit com Token (30 segundos)

```bash
# Token permite 100 req/s (muito mais que IP)
for i in {1..15}; do
  curl -s -H "API_KEY: abc123" http://localhost:8080/api/test | jq -r '.message // .error'
  sleep 0.1
done
```

**Resultado esperado:**
- Todas as 15 requisições são aceitas

### Cenário 3: Teste Automatizado (2 minutos)

```bash
chmod +x test-rate-limiter.sh
./test-rate-limiter.sh
```

## 🔧 Configuração Rápida

### Alterar Limites

Edite o arquivo `.env`:

```env
# Aumentar limite por IP para 20 req/s
RATE_LIMIT_IP=20

# Aumentar limite por token para 200 req/s
RATE_LIMIT_TOKEN=200

# Reduzir tempo de bloqueio para 60 segundos
BLOCK_DURATION_SECONDS=60
```

Reinicie:

```bash
docker-compose restart app
```

### Adicionar Token Customizado

No `.env`:

```env
# Novo token com limite de 500 req/s
TOKEN_premium_user=500
```

Teste:

```bash
curl -H "API_KEY: premium_user" http://localhost:8080/api/test
```

## 🛑 Parar e Limpar

```bash
# Parar containers
docker-compose down

# Parar e remover volumes (limpa Redis)
docker-compose down -v
```

## 🐛 Problemas Comuns

### Problema: "Cannot connect to the Docker daemon"

**Solução:** Inicie o Docker Desktop

### Problema: "port is already allocated"

**Solução:** Altere a porta no `docker-compose.yml`:

```yaml
ports:
  - "8081:8080"  # Use 8081 em vez de 8080
```

### Problema: Todas requisições são bloqueadas

**Solução:** Limpe o Redis:

```bash
docker exec -it rate-limiter-redis redis-cli FLUSHALL
```

## 📚 Próximos Passos

1. ✅ **Leia o README completo** → `README.md`
2. 🧪 **Execute os testes** → `TESTING.md`
3. 🏗️ **Entenda a arquitetura** → `ARCHITECTURE.md`
4. 🔧 **Customize as configurações** → `.env`

## 🎓 Exemplos Práticos

### Integrar em seu projeto

```go
import (
    "github.com/celiojsf/pos-challenge-rate-limiter/internal/limiter"
    "github.com/celiojsf/pos-challenge-rate-limiter/internal/storage"
    "github.com/celiojsf/pos-challenge-rate-limiter/internal/middleware"
)

// Setup
store, _ := storage.NewRedisStorage("localhost:6379", "", 0)
rl := limiter.NewRateLimiter(limiter.Config{
    Storage:       store,
    IPLimit:       10,
    TokenLimit:    100,
    BlockDuration: 5 * time.Minute,
})

// Use como middleware
mw := middleware.NewRateLimiterMiddleware(rl)
router.Use(mw.Handle)
```

### Criar novo backend de storage

```go
type MyStorage struct {
    // seus campos
}

func (m *MyStorage) Increment(ctx context.Context, key string, exp time.Duration) (int64, error) {
    // sua implementação
}

// Implemente os outros métodos da interface Storage
```

## 💡 Dicas

1. **Use tokens para clientes conhecidos** - dá a eles limites maiores
2. **Configure limites baseados em sua capacidade** - não muito alto, não muito baixo
3. **Monitore as métricas** - veja quantas requisições são bloqueadas
4. **Teste sob carga** - use `ab`, `hey` ou `wrk`
5. **Documente seus limites** - seus usuários precisam saber

## 🆘 Precisa de Ajuda?

- 📖 Documentação completa: `README.md`
- 🧪 Guia de testes: `TESTING.md`
- 🏗️ Arquitetura: `ARCHITECTURE.md`
- 🐛 Issues: GitHub Issues
- 💬 Discussões: GitHub Discussions

## ✅ Checklist de Sucesso

- [ ] Docker está instalado e rodando
- [ ] Containers iniciaram com sucesso
- [ ] Health check retorna 200 OK
- [ ] Rate limit por IP funciona (10 req/s)
- [ ] Rate limit por token funciona (100 req/s)
- [ ] Bloqueio temporário funciona (300s)
- [ ] Mensagem 429 está correta
- [ ] Scripts de teste executam com sucesso

## 🎉 Parabéns!

Você configurou com sucesso o Rate Limiter! Agora você tem:

✅ Um rate limiter profissional em Go
✅ Proteção contra abuso de API
✅ Suporte a múltiplos clientes
✅ Sistema escalável com Redis
✅ Testes automatizados completos

---

**Próximo passo:** Leia o `README.md` completo para entender todos os recursos disponíveis!
