#!/bin/bash

echo "=== Testando validação de tokens ==="
echo ""

# Teste 1: Token registrado com valor customizado
echo "1. Testando token 'abc123' (registrado com limite de 100):"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/ -H "API_KEY: abc123"
echo ""

# Teste 2: Token registrado sem valor (deve usar RATE_LIMIT_TOKEN)
echo "2. Testando token 'teste' (registrado sem valor, deve usar limite padrão de 100):"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/ -H "API_KEY: teste"
echo ""

# Teste 3: Token não registrado (deve retornar 403 - Forbidden)
echo "3. Testando token 'invalido' (não registrado, deve retornar acesso negado):"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/ -H "API_KEY: invalido"
echo ""

# Teste 4: Sem token (deve usar limite por IP)
echo "4. Testando sem token (deve usar limite por IP de 10):"
curl -s -w "\nHTTP Status: %{http_code}\n" http://localhost:8080/
echo ""
