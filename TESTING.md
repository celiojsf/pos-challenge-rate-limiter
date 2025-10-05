# Guia de Testes - Rate Limiter

Este documento fornece instruções detalhadas para testar o Rate Limiter.

## 🚀 Iniciando a Aplicação

### Com Docker (Recomendado)

```bash
# Iniciar todos os serviços
docker-compose up -d

# Verificar se os containers estão rodando
docker-compose ps

# Ver logs em tempo real
docker-compose logs -f app

# Parar os serviços
docker-compose down
```

### Localmente

```bash
# 1. Iniciar o Redis
docker run -d -p 6379:6379 redis:7-alpine

# 2. Executar a aplicação
go run cmd/server/main.go
```

## 🧪 Testes Unitários

### Executar todos os testes

```bash
make test
```

ou

```bash
go test ./... -v
```

### Executar testes com cobertura

```bash
make test-coverage
```

ou

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Executar testes de um pacote específico

```bash
# Testar apenas o limiter
go test ./internal/limiter -v

# Testar apenas o storage
go test ./internal/storage -v

# Testar apenas o middleware
go test ./internal/middleware -v
```

## 🔬 Testes Manuais

### 1. Teste Básico - Verificar se o servidor está rodando

```bash
curl http://localhost:8080/
```

**Resposta esperada:**
```json
{"message": "Rate Limiter API", "status": "ok"}
```

### 2. Teste de Health Check

```bash
curl http://localhost:8080/health
```

**Resposta esperada:**
```json
{"status": "healthy"}
```

### 3. Teste de Rate Limit por IP

Execute várias requisições sem token:

```bash
# Faça 15 requisições (limite padrão é 10)
for i in {1..15}; do
  echo "Requisição $i:"
  curl -i http://localhost:8080/api/test
  echo ""
  sleep 0.1
done
```

**Comportamento esperado:**
- As primeiras 10 requisições retornam `200 OK`
- Da 11ª em diante retornam `429 Too Many Requests`

### 4. Teste com Token de Acesso

```bash
# Token com limite de 100 req/s
for i in {1..15}; do
  echo "Requisição $i:"
  curl -H "API_KEY: abc123" http://localhost:8080/api/test
  echo ""
  sleep 0.1
done
```

**Comportamento esperado:**
- Todas as 15 requisições são aceitas (limite do token é 100)

### 5. Teste Token vs IP (Token sobrepõe IP)

```bash
# Primeiro, bloqueie seu IP fazendo muitas requisições sem token
for i in {1..12}; do
  curl -s http://localhost:8080/api/test > /dev/null
done

# Agora tente com token - deve funcionar!
curl -H "API_KEY: abc123" http://localhost:8080/api/test
```

**Comportamento esperado:**
- Requisições sem token são bloqueadas
- Requisições com token são aceitas (token sobrepõe IP)

### 6. Teste de Bloqueio Temporário

```bash
# 1. Exceda o limite
for i in {1..12}; do
  curl -s http://localhost:8080/api/test > /dev/null
done

# 2. Tente novamente imediatamente
curl http://localhost:8080/api/test
# Deve retornar 429

# 3. Aguarde o tempo de bloqueio (300 segundos por padrão)
# Ou reinicie o Redis para limpar:
docker-compose restart redis

# 4. Tente novamente
curl http://localhost:8080/api/test
# Deve retornar 200 OK
```

### 7. Teste com Diferentes IPs (simulado)

```bash
# Simule requisições de diferentes IPs usando o header X-Forwarded-For
curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/api/test
curl -H "X-Forwarded-For: 192.168.1.101" http://localhost:8080/api/test
curl -H "X-Forwarded-For: 192.168.1.102" http://localhost:8080/api/test
```

**Comportamento esperado:**
- Cada IP tem seu próprio contador independente

## 📊 Scripts de Teste Automatizados

### Teste Completo

```bash
./test-rate-limiter.sh
```

Este script testa:
- Limitação por IP
- Limitação por token
- Tokens com limites customizados
- Verificação de bloqueio

### Teste com Múltiplos IPs

```bash
./test-multiple-ips.sh
```

Simula requisições de diferentes IPs para verificar isolamento.

### Teste de Stress

```bash
./test-stress.sh
```

Envia múltiplas requisições simultâneas para testar concorrência.

## 🔍 Verificando o Redis

### Conectar ao Redis

```bash
docker exec -it rate-limiter-redis redis-cli
```

### Comandos úteis no Redis

```redis
# Listar todas as chaves
KEYS *

# Ver valor de uma chave específica
GET ratelimit:ip:192.168.1.1

# Ver tempo de expiração (TTL)
TTL ratelimit:ip:192.168.1.1

# Ver se um IP está bloqueado
GET block:ratelimit:ip:192.168.1.1

# Limpar todas as chaves (CUIDADO!)
FLUSHALL

# Sair
EXIT
```

## 🐛 Troubleshooting

### Problema: Todas as requisições são bloqueadas

**Solução:** Limpe o Redis
```bash
docker exec -it rate-limiter-redis redis-cli FLUSHALL
```

### Problema: Rate limiter não está funcionando

**Verificações:**
1. Redis está rodando?
```bash
docker-compose ps redis
```

2. Aplicação está conectada ao Redis?
```bash
docker-compose logs app | grep "Connected to Redis"
```

3. Verifique as variáveis de ambiente
```bash
docker-compose exec app env | grep RATE
```

### Problema: Porta 8080 em uso

**Solução:** Altere a porta no `docker-compose.yml`
```yaml
ports:
  - "8081:8080"  # Mude de 8080 para 8081
```

## 📈 Testes de Carga com ferramentas

### Usando Apache Bench (ab)

```bash
# Instalar (macOS)
brew install httpd

# Fazer 100 requisições, 10 concorrentes
ab -n 100 -c 10 http://localhost:8080/api/test
```

### Usando hey

```bash
# Instalar
go install github.com/rakyll/hey@latest

# Fazer 100 requisições, 10 concorrentes
hey -n 100 -c 10 http://localhost:8080/api/test
```

### Usando wrk

```bash
# Instalar (macOS)
brew install wrk

# 10 threads, 100 conexões, 30 segundos
wrk -t10 -c100 -d30s http://localhost:8080/api/test
```

## 📊 Cenários de Teste Recomendados

### Cenário 1: Tráfego Normal
- 5 requisições por segundo durante 1 minuto
- Nenhuma deve ser bloqueada

### Cenário 2: Pico de Tráfego
- 50 requisições simultâneas
- Verificar quantas são bloqueadas
- Confirmar que o sistema se recupera

### Cenário 3: Múltiplos Clientes
- 5 clientes diferentes (IPs diferentes)
- Cada um faz 15 requisições
- Verificar isolamento entre clientes

### Cenário 4: Token Privilegiado
- Cliente com token premium (limite 1000)
- Cliente sem token (limite 10)
- Verificar que o token permite mais requisições

### Cenário 5: Recuperação após Bloqueio
- Exceder limite
- Aguardar tempo de bloqueio
- Verificar que pode fazer requisições novamente

## 🎯 Critérios de Sucesso

✅ Servidor responde na porta 8080
✅ Health check retorna 200 OK
✅ Rate limiting por IP funciona corretamente
✅ Rate limiting por token funciona corretamente
✅ Token sobrepõe limitação por IP
✅ Bloqueio temporário funciona
✅ Mensagem de erro correta (429)
✅ Redis armazena os contadores corretamente
✅ Sistema se recupera após bloqueio
✅ Múltiplos IPs são tratados independentemente

## 📝 Relatório de Teste

Ao testar, documente:
- [ ] Data e hora do teste
- [ ] Configuração utilizada (limites)
- [ ] Cenários testados
- [ ] Resultados obtidos
- [ ] Problemas encontrados
- [ ] Screenshots ou logs relevantes

## 🔗 Links Úteis

- [Documentação Redis](https://redis.io/documentation)
- [Go Testing](https://golang.org/pkg/testing/)
- [Docker Compose](https://docs.docker.com/compose/)
- [HTTP Status Codes](https://httpstatuses.com/429)
