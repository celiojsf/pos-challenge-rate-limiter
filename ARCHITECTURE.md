# Arquitetura do Rate Limiter

## 📐 Visão Geral

Este documento descreve a arquitetura e as decisões de design do Rate Limiter.

## 🏛️ Princípios de Design

### 1. Separation of Concerns (Separação de Responsabilidades)
- **Storage Layer**: Responsável apenas pela persistência
- **Limiter Logic**: Contém apenas a lógica de rate limiting
- **Middleware**: Apenas integração HTTP
- **Config**: Apenas gerenciamento de configurações

### 2. Strategy Pattern
O storage usa o Strategy Pattern, permitindo trocar facilmente entre diferentes backends (Redis, Memory, MongoDB, etc.) sem alterar a lógica de negócio.

### 3. Dependency Injection
As dependências são injetadas, facilitando testes e manutenção.

### 4. Clean Architecture
Seguimos os princípios de Clean Architecture:
- Independência de frameworks
- Testabilidade
- Independência de UI
- Independência de banco de dados

## 🗂️ Estrutura de Camadas

```
┌─────────────────────────────────────┐
│         HTTP Layer                  │
│    (Chi Router + Middleware)        │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      Business Logic Layer           │
│         (Rate Limiter)              │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│       Storage Interface             │
│     (Strategy Pattern)              │
└──────────────┬──────────────────────┘
               │
       ┌───────┴────────┐
       │                │
┌──────▼─────┐   ┌─────▼──────┐
│   Redis    │   │   Memory   │
│  Storage   │   │  Storage   │
└────────────┘   └────────────┘
```

## 📦 Componentes Principais

### 1. Storage Interface

```go
type Storage interface {
    Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
    Get(ctx context.Context, key string) (int64, error)
    SetBlock(ctx context.Context, key string, expiration time.Duration) error
    IsBlocked(ctx context.Context, key string) (bool, error)
    Close() error
}
```

**Responsabilidades:**
- Definir contrato para storage
- Permitir múltiplas implementações
- Garantir consistência de interface

**Implementações:**
- `RedisStorage`: Para produção, suporta ambientes distribuídos
- `MemoryStorage`: Para desenvolvimento e testes

### 2. Rate Limiter

```go
type RateLimiter struct {
    storage       storage.Storage
    ipLimit       int
    tokenLimit    int
    blockDuration time.Duration
    tokenLimits   map[string]int
}
```

**Responsabilidades:**
- Verificar se requisição deve ser permitida
- Gerenciar contadores de requisições
- Aplicar bloqueios temporários
- Priorizar token sobre IP

**Algoritmo:**
1. Verifica se há token na requisição
2. Se sim, usa limite do token (ou customizado)
3. Se não, usa limite do IP
4. Verifica se está bloqueado
5. Incrementa contador
6. Se exceder, bloqueia por X segundos

### 3. Middleware

```go
type RateLimiterMiddleware struct {
    limiter *limiter.RateLimiter
}
```

**Responsabilidades:**
- Extrair IP da requisição
- Extrair token do header `API_KEY`
- Chamar o limiter
- Retornar resposta HTTP apropriada

**Extração de IP:**
1. Tenta `X-Forwarded-For` (proxy/load balancer)
2. Tenta `X-Real-IP`
3. Usa `RemoteAddr`

### 4. Configuration

```go
type Config struct {
    Redis         RedisConfig
    RateLimitIP   int
    RateLimitToken int
    BlockDuration  int
    TokenLimits    map[string]int
}
```

**Responsabilidades:**
- Carregar variáveis de ambiente
- Validar configurações
- Fornecer valores padrão
- Suportar tokens customizados

## 🔄 Fluxo de Dados

### Requisição HTTP Normal

```
1. HTTP Request → Chi Router
2. Router → Rate Limiter Middleware
3. Middleware → Extract IP & Token
4. Middleware → Rate Limiter.Allow()
5. Rate Limiter → Storage.IsBlocked()
6. Rate Limiter → Storage.Increment()
7. Rate Limiter → Check Limit
8. Rate Limiter → Storage.SetBlock() (se excedeu)
9. Rate Limiter → Return allowed/blocked
10. Middleware → Return HTTP 200 or 429
```

### Fluxo Detalhado do Rate Limiter

```go
func (rl *RateLimiter) Allow(ctx, ip, token) bool {
    // 1. Determinar qual limite usar
    key, limit := rl.determineKeyAndLimit(ip, token)
    
    // 2. Verificar se está bloqueado
    if blocked := storage.IsBlocked(key) {
        return false
    }
    
    // 3. Incrementar contador
    count := storage.Increment(key, 1*time.Second)
    
    // 4. Verificar se excedeu
    if count > limit {
        storage.SetBlock(key, blockDuration)
        return false
    }
    
    return true
}
```

## 🔑 Decisões de Design

### 1. Por que Redis?

**Vantagens:**
- ✅ Suporte nativo a TTL (Time To Live)
- ✅ Operações atômicas (INCR)
- ✅ Alta performance
- ✅ Suporte a ambientes distribuídos
- ✅ Persistência opcional

**Alternativas consideradas:**
- Memory: Bom para desenvolvimento, não escalável
- PostgreSQL: Muita overhead para contadores simples
- MongoDB: Desnecessariamente complexo para este caso

### 2. Janela Deslizante (Sliding Window)

Optamos por janela deslizante com TTL de 1 segundo:

**Vantagens:**
- Simples de implementar
- Precisão de 1 segundo
- Não requer limpeza manual
- Redis gerencia expiração automaticamente

**Alternativas:**
- Fixed Window: Mais simples, mas permite burst no reset
- Token Bucket: Mais complexo, permite burst controlado
- Leaky Bucket: Mais justo, mas mais complexo

### 3. Strategy Pattern para Storage

**Benefícios:**
- Fácil trocar backend sem mudar código
- Testável (pode usar Memory em testes)
- Extensível (pode adicionar MongoDB, Memcached, etc.)
- Segue SOLID (Open/Closed Principle)

### 4. Middleware Separado da Lógica

**Benefícios:**
- Lógica de rate limiting reutilizável
- Pode ser testada independentemente
- Pode ser usada em diferentes frameworks
- Facilita manutenção

## 🔒 Segurança

### 1. IP Spoofing
- Usamos `X-Forwarded-For` com cuidado
- Em produção, confiar apenas em proxies conhecidos
- Validar headers

### 2. Token Security
- Tokens devem ser longos e aleatórios
- Considerar usar JWT em produção
- Implementar rotação de tokens

### 3. DoS Protection
- Rate limiter é a primeira linha de defesa
- Considerar adicionar WAF
- Implementar circuit breaker

## 📈 Escalabilidade

### Horizontal Scaling

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  App 1   │     │  App 2   │     │  App 3   │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     └────────────────┼────────────────┘
                      │
               ┌──────▼──────┐
               │    Redis    │
               │  (Shared)   │
               └─────────────┘
```

**Benefícios:**
- Múltiplas instâncias da aplicação
- Redis compartilhado garante contadores globais
- Load balancer distribui tráfego

### Redis Clustering

Para alta disponibilidade:
- Redis Sentinel (failover automático)
- Redis Cluster (sharding)
- Redis Enterprise

## 🧪 Testabilidade

### Unit Tests
- Cada componente é testado isoladamente
- Mocks são fáceis de criar (interfaces)
- Coverage > 80%

### Integration Tests
- Testam fluxo completo
- Usam Memory Storage (rápido)
- Podem usar Redis em CI/CD

### Load Tests
- Scripts incluídos para teste de carga
- Podem usar ferramentas externas (ab, hey, wrk)

## 🚀 Performance

### Otimizações Implementadas

1. **Pipeline Redis**: Usa pipeline para INCR + EXPIRE
2. **Context**: Suporta cancelamento e timeout
3. **Goroutines**: Requisições são concorrentes
4. **Memory Storage**: Usa sync.RWMutex para concorrência

### Benchmarks Estimados

```
Redis Storage:
- Latency: ~1-2ms por requisição
- Throughput: ~10,000 req/s por instância

Memory Storage:
- Latency: ~0.1ms por requisição  
- Throughput: ~50,000 req/s por instância
```

## 🔮 Melhorias Futuras

### Curto Prazo
- [ ] Métricas com Prometheus
- [ ] Distributed tracing
- [ ] Dashboard de monitoramento
- [ ] Rate limit por endpoint

### Médio Prazo
- [ ] Token Bucket algorithm
- [ ] Burst allowance
- [ ] Whitelist/Blacklist
- [ ] Rate limit dinâmico

### Longo Prazo
- [ ] Machine Learning para detecção de anomalias
- [ ] Auto-scaling baseado em carga
- [ ] Multi-região support
- [ ] GraphQL support

## 📊 Monitoramento

### Métricas Importantes

```
rate_limiter_requests_total{status="allowed"}
rate_limiter_requests_total{status="blocked"}
rate_limiter_block_duration_seconds
rate_limiter_storage_latency_seconds
```

### Logs Importantes

```
- Requisições bloqueadas
- Erros de conexão com Redis
- Tempos de resposta altos
- Configuração ao iniciar
```

### Alertas Sugeridos

```
- Taxa de bloqueio > 50% por 5 minutos
- Latência do Redis > 10ms
- Redis desconectado
- Memória Redis > 80%
```

## 🤝 Contribuindo

Para adicionar novos backends de storage:

1. Implemente a interface `Storage`
2. Adicione testes
3. Documente
4. Submeta PR

Exemplo:

```go
type MongoStorage struct {
    client *mongo.Client
}

func (m *MongoStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
    // Implementação
}

// ... outros métodos
```

## 📚 Referências

- [Rate Limiting Algorithms](https://en.wikipedia.org/wiki/Rate_limiting)
- [Redis Documentation](https://redis.io/documentation)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [SOLID Principles](https://en.wikipedia.org/wiki/SOLID)
- [Strategy Pattern](https://refactoring.guru/design-patterns/strategy)
