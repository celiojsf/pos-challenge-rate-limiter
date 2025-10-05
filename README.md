# Rate Limiter em Go

Um rate limiter robusto e configurável desenvolvido em Go, capaz de limitar requisições por IP e por Token de acesso, com suporte a Redis para persistência distribuída.

## 📋 Índice

- [Características](#características)
- [Arquitetura](#arquitetura)
- [Pré-requisitos](#pré-requisitos)
- [Instalação](#instalação)
- [Configuração](#configuração)
- [Uso](#uso)
- [Testes](#testes)
- [API](#api)
- [Exemplos](#exemplos)

## ✨ Características

- **Limitação por IP**: Controla o número de requisições por endereço IP
- **Limitação por Token**: Controla requisições usando tokens de acesso (API_KEY)
- **Tokens Customizados**: Permite configurar limites diferentes para tokens específicos
- **Prioridade de Token**: Configurações de token sobrepõem as de IP
- **Bloqueio Temporário**: Bloqueia IPs/tokens que excedem o limite por tempo configurável
- **Storage Strategy Pattern**: Implementação com interface para fácil troca de backend
- **Redis Storage**: Suporte completo a Redis para ambientes distribuídos
- **Memory Storage**: Implementação em memória para testes e desenvolvimento
- **Middleware Independente**: Lógica de rate limiting separada do middleware HTTP
- **Docker Ready**: Totalmente containerizado com Docker Compose
- **Testes Automatizados**: Suite completa de testes unitários e de integração
- **Graceful Shutdown**: Desligamento seguro do servidor

## 🏗️ Arquitetura

O projeto segue uma arquitetura em camadas com separação clara de responsabilidades:

```
pos-challenge-rate-limiter/
├── cmd/
│   └── server/              # Ponto de entrada da aplicação
│       └── main.go
├── internal/
│   ├── config/              # Configurações da aplicação
│   │   └── config.go
│   ├── storage/             # Strategy Pattern para storage
│   │   ├── storage.go       # Interface
│   │   ├── redis.go         # Implementação Redis
│   │   ├── memory.go        # Implementação em memória
│   │   └── *_test.go        # Testes
│   ├── limiter/             # Lógica de rate limiting
│   │   ├── limiter.go
│   │   └── limiter_test.go
│   └── middleware/          # Middleware HTTP
│       ├── ratelimiter.go
│       └── ratelimiter_test.go
├── docker-compose.yml       # Orquestração de containers
├── Dockerfile               # Imagem da aplicação
├── Makefile                 # Comandos úteis
├── .env                     # Configurações de ambiente
└── README.md
```

### Componentes Principais

1. **Storage Layer**: Interface abstrata que permite trocar facilmente entre Redis, memória ou outros backends
2. **Limiter**: Contém toda a lógica de rate limiting, independente do framework HTTP
3. **Middleware**: Camada HTTP que integra o limiter com o servidor web
4. **Config**: Gerenciamento centralizado de configurações via variáveis de ambiente

## 🔧 Pré-requisitos

- Docker e Docker Compose (recomendado)
- **OU** Go 1.21+ e Redis (para desenvolvimento local)

## 📦 Instalação

### Usando Docker (Recomendado)

1. Clone o repositório:
```bash
git clone https://github.com/celiojsf/pos-challenge-rate-limiter.git
cd pos-challenge-rate-limiter
```

2. Inicie os containers:
```bash
docker-compose up -d
```

A aplicação estará disponível em `http://localhost:8080`

### Desenvolvimento Local

1. Clone o repositório:
```bash
git clone https://github.com/celiojsf/pos-challenge-rate-limiter.git
cd pos-challenge-rate-limiter
```

2. Instale as dependências:
```bash
go mod download
```

3. Inicie o Redis:
```bash
docker run -d -p 6379:6379 redis:7-alpine
```

4. Execute a aplicação:
```bash
go run cmd/server/main.go
```

## ⚙️ Configuração

As configurações são feitas através de variáveis de ambiente ou arquivo `.env`:

### Variáveis de Ambiente

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `REDIS_HOST` | Host do Redis | `localhost` |
| `REDIS_PORT` | Porta do Redis | `6379` |
| `REDIS_PASSWORD` | Senha do Redis | `` |
| `REDIS_DB` | Database do Redis | `0` |
| `RATE_LIMIT_IP` | Limite de requisições por IP (req/s) | `10` |
| `RATE_LIMIT_TOKEN` | Limite padrão para tokens (req/s) | `100` |
| `BLOCK_DURATION_SECONDS` | Tempo de bloqueio em segundos | `300` |
| `TOKEN_<nome>` | Limite customizado para token específico | - |

### Exemplo de Configuração

```env
# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiter
RATE_LIMIT_IP=10
RATE_LIMIT_TOKEN=100
BLOCK_DURATION_SECONDS=300

# Tokens Customizados
TOKEN_abc123=100
TOKEN_xyz789=50
TOKEN_premium_user=1000
```

## 🚀 Uso

### Comandos Make

```bash
make help              # Mostra todos os comandos disponíveis
make build             # Compila a aplicação
make run               # Executa localmente
make test              # Executa os testes
make test-coverage     # Executa testes com cobertura
make docker-up         # Inicia containers Docker
make docker-down       # Para containers Docker
make docker-logs       # Visualiza logs dos containers
make docker-rebuild    # Reconstrói containers do zero
```

### Requisições HTTP

#### Requisição Normal (limitada por IP)
```bash
curl http://localhost:8080/api/test
```

#### Requisição com Token
```bash
curl -H "API_KEY: abc123" http://localhost:8080/api/test
```

#### Health Check
```bash
curl http://localhost:8080/health
```

## 🧪 Testes

### Testes Automatizados

Execute a suite completa de testes:

```bash
make test
```

Ou com cobertura:

```bash
make test-coverage
```

### Testes de Carga

O projeto inclui scripts para testar o comportamento do rate limiter:

#### Teste Básico
```bash
chmod +x test-rate-limiter.sh
./test-rate-limiter.sh
```

#### Teste com Múltiplos IPs
```bash
chmod +x test-multiple-ips.sh
./test-multiple-ips.sh
```

#### Teste de Stress
```bash
chmod +x test-stress.sh
./test-stress.sh
```

## 📡 API

### Endpoints

#### `GET /`
Endpoint raiz com informações da API.

**Resposta:**
```json
{
  "message": "Rate Limiter API",
  "status": "ok"
}
```

#### `GET /health`
Health check da aplicação.

**Resposta:**
```json
{
  "status": "healthy"
}
```

#### `GET /api/test`
Endpoint de teste para verificar o rate limiter.

**Resposta de Sucesso (200):**
```json
{
  "message": "Test endpoint",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Resposta de Rate Limit Excedido (429):**
```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

### Headers

- `API_KEY`: Token de acesso para rate limiting baseado em token

## 📚 Exemplos

### Exemplo 1: Limitação por IP

Configuração: `RATE_LIMIT_IP=5`

```bash
# Primeiras 5 requisições são aceitas
for i in {1..5}; do
  curl http://localhost:8080/api/test
  echo ""
done

# 6ª requisição é bloqueada
curl http://localhost:8080/api/test
# Resposta: 429 Too Many Requests
```

### Exemplo 2: Limitação por Token

Configuração: `TOKEN_abc123=10`

```bash
# Primeiras 10 requisições são aceitas
for i in {1..10}; do
  curl -H "API_KEY: abc123" http://localhost:8080/api/test
  echo ""
done

# 11ª requisição é bloqueada
curl -H "API_KEY: abc123" http://localhost:8080/api/test
# Resposta: 429 Too Many Requests
```

### Exemplo 3: Token Sobrepõe IP

Configuração:
- `RATE_LIMIT_IP=5`
- `TOKEN_abc123=20`

```bash
# Mesmo que o IP tenha limite de 5, o token permite 20
for i in {1..15}; do
  curl -H "API_KEY: abc123" http://localhost:8080/api/test
  echo ""
done
# Todas as 15 requisições são aceitas
```

### Exemplo 4: Bloqueio e Recuperação

```bash
# Exceder o limite
for i in {1..12}; do
  curl http://localhost:8080/api/test
done

# IP está bloqueado
curl http://localhost:8080/api/test
# Resposta: 429

# Aguardar BLOCK_DURATION_SECONDS (300s por padrão)
sleep 300

# IP está desbloqueado
curl http://localhost:8080/api/test
# Resposta: 200 OK
```

## 🔄 Strategy Pattern - Trocar Storage

Para trocar o backend de Redis para outro sistema, basta implementar a interface `Storage`:

```go
type Storage interface {
    Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
    Get(ctx context.Context, key string) (int64, error)
    SetBlock(ctx context.Context, key string, expiration time.Duration) error
    IsBlocked(ctx context.Context, key string) (bool, error)
    Close() error
}
```

Exemplo com MongoDB:

```go
type MongoStorage struct {
    client *mongo.Client
    // ... campos necessários
}

func (m *MongoStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
    // Implementação com MongoDB
}

// Implementar outros métodos...
```

Depois, basta substituir no `main.go`:

```go
// Antes
store, err := storage.NewRedisStorage(redisAddr, cfg.Redis.Password, cfg.Redis.DB)

// Depois
store, err := storage.NewMongoStorage(mongoURI)
```

## 🧩 Como Funciona

### Fluxo de Requisição

1. **Requisição HTTP chega** ao servidor
2. **Middleware extrai** o IP e token (se presente)
3. **Limiter verifica** se IP/token está bloqueado
4. Se não bloqueado, **incrementa contador** no storage
5. Se contador **excede limite**, **bloqueia** por `BLOCK_DURATION_SECONDS`
6. **Retorna resposta** (200 OK ou 429 Too Many Requests)

### Algoritmo de Rate Limiting

O rate limiter usa um algoritmo de **janela deslizante por segundo** (sliding window):

- Cada requisição incrementa um contador com TTL de 1 segundo
- Se o contador exceder o limite, o IP/token é bloqueado
- O bloqueio dura `BLOCK_DURATION_SECONDS`
- Após o bloqueio, o contador é resetado

### Prioridade

1. **Token** sempre tem prioridade sobre IP
2. **Token customizado** tem prioridade sobre limite padrão
3. Ordem de verificação: Token Customizado → Token Padrão → IP

## 🐛 Troubleshooting

### Problema: "Failed to connect to Redis"

**Solução**: Verifique se o Redis está rodando:
```bash
docker ps | grep redis
```

Se não estiver, inicie com:
```bash
docker-compose up -d redis
```

### Problema: Testes falhando

**Solução**: Limpe o cache e recompile:
```bash
go clean -testcache
go test ./...
```

### Problema: Porta 8080 já em uso

**Solução**: Altere a porta no `docker-compose.yml` ou pare o serviço que está usando:
```bash
lsof -ti:8080 | xargs kill
```

## 📝 Licença

Este projeto foi desenvolvido como parte de um desafio técnico.

## 👤 Autor

Celio José dos Santos Filho
- GitHub: [@celiojsf](https://github.com/celiojsf)

## 🙏 Agradecimentos

Desenvolvido como parte do desafio de Pós-Graduação em Arquitetura de Software.