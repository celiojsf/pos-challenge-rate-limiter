# Rate Limiter em Go

Este projeto implementa um rate limiter configurável que pode ser usado como middleware em aplicações web. Ele permite controlar o número de requisições por segundo baseado em:

- **Endereço IP**: Limita requisições vindas do mesmo IP
- **Token de Acesso**: Permite limites customizados por token (via header `API_KEY`)

O rate limiter utiliza o **Strategy Pattern** para o storage, permitindo fácil substituição do backend (Redis, Memory, etc.).

### Características Principais

✅ Limitação por IP e Token  
✅ Tokens customizados com limites diferentes  
✅ Token sobrepõe limitação por IP  
✅ Bloqueio temporário configurável  
✅ Redis para persistência distribuída  
✅ Strategy Pattern para fácil troca de backend  
✅ Middleware independente da lógica de negócio  
✅ Testes automatizados completos  
✅ Docker Compose para fácil setup  

### Arquitetura

```
├── cmd/server/              # Aplicação principal
├── internal/
│   ├── config/              # Gerenciamento de configurações
│   ├── storage/             # Interface e implementações de storage
│   ├── limiter/             # Lógica de rate limiting
│   └── middleware/          # Middleware HTTP
├── test-*.sh                # Scripts de teste de carga
├── docker-compose.yml       # Orquestração Docker
├── Dockerfile
└── .env                     # Configurações
```

## 🚀 Como Executar

### Pré-requisitos

- Docker e Docker Compose
- (Opcional) Go 1.21+ para desenvolvimento local

### Iniciar com Docker Compose

```bash
# 1. Clone o repositório
git clone https://github.com/celiojsf/pos-challenge-rate-limiter.git
cd pos-challenge-rate-limiter

# 2. Inicie os containers (Redis + Aplicação)
docker-compose up -d

# 3. Verifique se está rodando
docker-compose ps

# 4. Veja os logs
docker-compose logs -f app
```

A aplicação estará disponível em **http://localhost:8080**

### Parar a Aplicação

```bash
# Parar containers
docker-compose down

# Parar e limpar volumes (limpa dados do Redis)
docker-compose down -v
```

## ⚙️ Configuração

Edite o arquivo `.env` para ajustar os limites:

```env
# Configuração do Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_DB=0

# Limites de Rate Limiting
RATE_LIMIT_IP=10                # Requisições por segundo por IP
RATE_LIMIT_TOKEN=100            # Requisições por segundo por token (padrão)
BLOCK_DURATION_SECONDS=300      # Tempo de bloqueio (5 minutos)

# Tokens Customizados (formato: TOKEN_<nome>=<limite>)
TOKEN_abc123=100
TOKEN_xyz789=50
```

Após alterar, reinicie a aplicação:

```bash
docker-compose restart app
```

## 🧪 Testes

### Testes Unitários Go

```bash
# Executar todos os testes
go test ./... -v

# Executar testes com cobertura
go test ./... -v -cover

# Ver relatório de cobertura HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Testes de Carga (Scripts Shell)

O projeto inclui 3 scripts de teste:

#### 1. Teste Completo de Rate Limiting
```bash
chmod +x test-rate-limiter.sh
./test-rate-limiter.sh
```

**O que testa:**
- Limitação por IP (10 req/s)
- Limitação por Token (100 req/s)
- Tokens customizados
- Bloqueio após exceder limite

**Resultado esperado:**
- ✓ Primeiras requisições são aceitas
- ✗ Requisições após o limite são bloqueadas

#### 2. Teste com Múltiplos IPs
```bash
chmod +x test-multiple-ips.sh
./test-multiple-ips.sh
```

**O que testa:**
- Isolamento entre diferentes IPs
- Cada IP tem seu próprio contador

**Resultado esperado:**
- Cada IP consegue fazer até 10 requisições
- IPs diferentes não interferem entre si

#### 3. Teste de Stress (Concorrência)
```bash
chmod +x test-stress.sh
./test-stress.sh
```

**O que testa:**
- Múltiplas requisições simultâneas
- Comportamento sob carga

**Resultado esperado:**
- Sistema mantém controle correto mesmo com requisições concorrentes

## 📡 Endpoints da API

### `GET /`
Informações sobre a API

```bash
curl http://localhost:8080/
```

Resposta:
```json
{"message": "Rate Limiter API", "status": "ok"}
```

### `GET /health`
Health check

```bash
curl http://localhost:8080/health
```

Resposta:
```json
{"status": "healthy"}
```

### `GET /api/test`
Endpoint de teste (com rate limiting)

```bash
# Sem token (limitado por IP)
curl http://localhost:8080/api/test

# Com token
curl -H "API_KEY: abc123" http://localhost:8080/api/test
```

Respostas:
- **200 OK**: Requisição permitida
- **429 Too Many Requests**: Limite excedido

```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

## 💡 Exemplos de Uso

### Exemplo 1: Testar Limite por IP

```bash
# Fazer 12 requisições (limite é 10)
for i in {1..12}; do
  echo "Requisição $i:"
  curl -s http://localhost:8080/api/test | jq
  sleep 0.1
done
```

**Resultado esperado:**
- Requisições 1-10: ✓ Sucesso (200 OK)
- Requisições 11-12: ✗ Bloqueadas (429)

### Exemplo 2: Testar com Token

```bash
# Token permite 100 req/s (muito mais que IP)
for i in {1..15}; do
  echo "Requisição $i com token:"
  curl -s -H "API_KEY: abc123" http://localhost:8080/api/test | jq
  sleep 0.1
done
```

**Resultado esperado:**
- Todas as 15 requisições: ✓ Sucesso (200 OK)

### Exemplo 3: Verificar Bloqueio

```bash
# 1. Exceder o limite
for i in {1..12}; do curl -s http://localhost:8080/api/test > /dev/null; done

# 2. Tentar nova requisição (deve ser bloqueada)
curl http://localhost:8080/api/test
# Resposta: 429

# 3. Limpar Redis para desbloquear
docker exec -it rate-limiter-redis redis-cli FLUSHALL

# 4. Tentar novamente (deve funcionar)
curl http://localhost:8080/api/test
# Resposta: 200
```

## 🔍 Como Funciona

### Fluxo de uma Requisição

1. Cliente faz requisição HTTP
2. Middleware extrai IP e token (header `API_KEY`)
3. Rate Limiter verifica:
   - Se token presente → usa limite do token
   - Se não → usa limite do IP
4. Verifica se está bloqueado no Redis
5. Incrementa contador (TTL de 1 segundo)
6. Se exceder limite → bloqueia por X segundos
7. Retorna 200 (OK) ou 429 (Too Many Requests)

### Prioridades

1. **Token customizado** (ex: `TOKEN_abc123=100`)
2. **Token padrão** (`RATE_LIMIT_TOKEN=100`)
3. **IP** (`RATE_LIMIT_IP=10`)

**Importante:** Token sempre sobrepõe IP!

## 🔧 Desenvolvimento Local

### Sem Docker

```bash
# 1. Inicie o Redis
docker run -d -p 6379:6379 redis:7-alpine

# 2. Instale dependências
go mod download

# 3. Execute a aplicação
go run cmd/server/main.go
```

### Verificar Redis

```bash
# Conectar ao Redis
docker exec -it rate-limiter-redis redis-cli

# Ver todas as chaves
KEYS *

# Ver contador de um IP
GET ratelimit:ip:192.168.1.1

# Ver se está bloqueado
GET block:ratelimit:ip:192.168.1.1

# Limpar tudo
FLUSHALL
```

## 🐛 Troubleshooting

### Containers não iniciam
```bash
docker-compose down -v
docker-compose up -d --build
```

### Todas requisições bloqueadas
```bash
# Limpar Redis
docker exec -it rate-limiter-redis redis-cli FLUSHALL
```

### Porta 8080 em uso
```bash
# Matar processo na porta
lsof -ti:8080 | xargs kill

# Ou altere a porta no docker-compose.yml
```

### Ver logs de erro
```bash
docker-compose logs app
docker-compose logs redis
```

## 📝 Especificação do Desafio

Este projeto atende aos seguintes requisitos:

✅ Rate limiter funciona como middleware  
✅ Configuração via variáveis de ambiente  
✅ Limitação por IP e Token  
✅ Token sobrepõe IP  
✅ Bloqueio temporário configurável  
✅ Responde 429 com mensagem correta  
✅ Storage em Redis  
✅ Strategy Pattern para trocar storage  
✅ Lógica separada do middleware  
✅ Testes automatizados  
✅ Docker Compose  
✅ Servidor na porta 8080  

## 👤 Autor

**Celio José dos Santos Filho**  
GitHub: [@celiojsf](https://github.com/celiojsf)

---

**Projeto desenvolvido para o desafio de Pós-Graduação em Arquitetura de Software**