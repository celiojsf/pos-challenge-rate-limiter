#!/bin/bash

# Script de teste de carga para o Rate Limiter
# Este script testa o comportamento do rate limiter sob diferentes cenários

BASE_URL="http://localhost:8080"
API_ENDPOINT="$BASE_URL/api/test"

echo "=========================================="
echo "Rate Limiter - Testes de Carga"
echo "=========================================="
echo ""

# Função para fazer requisições
make_requests() {
    local count=$1
    local token=$2
    local name=$3
    
    echo "Teste: $name"
    echo "Fazendo $count requisições..."
    
    local success=0
    local blocked=0
    
    for ((i=1; i<=$count; i++)); do
        if [ -z "$token" ]; then
            response=$(curl -s -w "\n%{http_code}" "$API_ENDPOINT")
        else
            response=$(curl -s -w "\n%{http_code}" -H "API_KEY: $token" "$API_ENDPOINT")
        fi
        
        status_code=$(echo "$response" | tail -n 1)
        
        if [ "$status_code" = "200" ]; then
            ((success++))
            echo -n "✓"
        elif [ "$status_code" = "429" ]; then
            ((blocked++))
            echo -n "✗"
        else
            echo -n "?"
        fi
        
        # Pequeno delay para simular tráfego real
        sleep 0.1
    done
    
    echo ""
    echo "Resultado: $success requisições bem-sucedidas, $blocked bloqueadas"
    echo ""
}

# Teste 1: Limite por IP
echo "----------------------------------------"
echo "Teste 1: Limitação por IP"
echo "Limite configurado: 10 req/s"
echo "----------------------------------------"
make_requests 15 "" "15 requisições sem token"

sleep 2

# Teste 2: Limite por Token
echo "----------------------------------------"
echo "Teste 2: Limitação por Token"
echo "Limite configurado: 100 req/s"
echo "----------------------------------------"
make_requests 15 "abc123" "15 requisições com token abc123"

sleep 2

# Teste 3: Token customizado
echo "----------------------------------------"
echo "Teste 3: Token com limite customizado"
echo "Token xyz789 - Limite: 50 req/s"
echo "----------------------------------------"
make_requests 15 "xyz789" "15 requisições com token xyz789"

sleep 2

# Teste 4: Verificar bloqueio
echo "----------------------------------------"
echo "Teste 4: Verificação de Bloqueio"
echo "----------------------------------------"
echo "Fazendo requisições para exceder o limite..."
make_requests 12 "" "Excedendo limite por IP"

echo "Aguardando 2 segundos..."
sleep 2

echo "Tentando fazer mais requisições (devem ser bloqueadas)..."
make_requests 5 "" "Requisições após bloqueio"

echo ""
echo "=========================================="
echo "Testes concluídos!"
echo "=========================================="
