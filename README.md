# Rate Limiter em Go

Este projeto implementa um rate limiter configurável que pode ser usado como middleware em aplicações web. Ele permite controlar o número de requisições por segundo baseado em:

- **Endereço IP**: Limita requisições vindas do mesmo IP
- **Token de Acesso**: Permite limites customizados por token (via header `API_KEY`)

O rate limiter utiliza **Redis** como backend de armazenamento e implementa o **Strategy Pattern**, permitindo fácil substituição por outros sistemas de persistência se necessário.

### Características Principais

✅ Limitação por IP e Token  
✅ Tokens customizados com limites diferentes  
✅ Token sobrepõe limitação por IP  
✅ Tokens com limite padrão (usando `RATE_LIMIT_TOKEN_DEFAULT`)  
✅ Validação de tokens registrados (rejeita tokens não cadastrados)  
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
│   ├── storage/             # Interface e implementação Redis
│   ├── limiter/             # Lógica de rate limiting
│   └── middleware/          # Middleware HTTP
├── test-scenarios.sh        # Script de teste completo
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

# 2. Crie o arquivo .env com as configurações
cat > .env << 'EOF'
# Rate Limiter Configuration

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Rate Limiter Settings
RATE_LIMIT_IP=5
RATE_LIMIT_TOKEN_DEFAULT=15
BLOCK_DURATION_SECONDS=300

# Token Configuration (example tokens with custom limits)
# Format: TOKEN_<TOKEN_VALUE>=<LIMIT>
# If limit is empty, it will use RATE_LIMIT_TOKEN_DEFAULT as default
TOKEN_abc123=10
TOKEN_xyz789=20
TOKEN_teste=
EOF

# 3. Inicie os containers (Redis + Aplicação)
docker-compose up -d

# 4. Verifique se está rodando
docker-compose ps

# 5. Veja os logs
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

## 🧪 Executar Testes

### Script de Teste Completo

O projeto inclui um script abrangente que testa todos os cenários do rate limiter:

```bash
chmod +x test-scenarios.sh
./test-scenarios.sh
```

**Cenários Testados:**

1. **Limitação por IP (sem token)**
   - Valida que requisições são limitadas por IP
   - Testa bloqueio após exceder o limite configurado

2. **Token com Limite Padrão**
   - Testa tokens configurados sem valor (`TOKEN_teste=`)
   - Valida que usam `RATE_LIMIT_TOKEN_DEFAULT`

3. **Múltiplos IPs (Isolamento)**
   - Verifica que diferentes IPs têm contadores independentes
   - Cada IP pode fazer até o limite sem afetar outros

4. **Token com Limite Customizado (abc123)**
   - Testa token com limite específico definido
   - Valida bloqueio após exceder o limite customizado

5. **Token com Outro Limite Customizado (xyz789)**
   - Testa outro token com limite diferente
   - Confirma que cada token respeita seu próprio limite

6. **Token Inválido/Não Registrado**
   - Valida rejeição de tokens não cadastrados
   - Espera HTTP 403 (Forbidden) para tokens inválidos

**Recursos do Script:**
- ✅ Limpa cache do Redis entre cada cenário
- ✅ Logs coloridos e detalhados
- ✅ Contadores de sucesso/falha por cenário
- ✅ Resumo final com todas as configurações testadas

## ⚙️ Configuração

### Como Funciona a Configuração de Tokens

1. **Tokens com Limite Customizado**: Defina `TOKEN_<nome>=<valor>` para criar um token com limite específico
   - Exemplo: `TOKEN_abc123=10` → Token "abc123" terá limite de 10 requisições/segundo

2. **Tokens com Limite Padrão**: Defina `TOKEN_<nome>=` (vazio) para usar o limite padrão
   - Exemplo: `TOKEN_teste=` → Token "teste" usará o valor de `RATE_LIMIT_TOKEN_DEFAULT`

3. **Tokens Não Registrados**: Qualquer token que não esteja definido no `.env` será **rejeitado** com HTTP 403 (Forbidden)

**Nota:** O Docker Compose carrega automaticamente as variáveis do arquivo `.env`. As configurações para o REDIS são sobrescritos quando rodando em containers.

Após alterar as configurações, é necessário recriar os containers:

```bash
# Parar e recriar os containers com as novas configurações
docker-compose down
docker-compose up -d
```

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

1. **Token customizado com valor default ou próprio de limite** (ex: `TOKEN_abc123=100`)
2. **IP** (`RATE_LIMIT_IP=10`)

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

## 👤 Autor

**Celio José dos Santos Filho**  
GitHub: [@celiojsf](https://github.com/celiojsf)

---