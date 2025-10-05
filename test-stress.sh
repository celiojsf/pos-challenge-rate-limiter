#!/bin/bash

# Script para teste de stress - envia muitas requisições simultâneas

BASE_URL="http://localhost:8080"
API_ENDPOINT="$BASE_URL/api/test"
CONCURRENT_REQUESTS=50

echo "=========================================="
echo "Teste de Stress - Rate Limiter"
echo "=========================================="
echo ""
echo "Enviando $CONCURRENT_REQUESTS requisições simultâneas..."
echo ""

# Array para armazenar PIDs dos processos
pids=()

# Função para fazer uma requisição
make_request() {
    local id=$1
    response=$(curl -s -w "\n%{http_code}" "$API_ENDPOINT" 2>/dev/null)
    status_code=$(echo "$response" | tail -n 1)
    
    if [ "$status_code" = "200" ]; then
        echo "[$id] ✓ Sucesso"
    elif [ "$status_code" = "429" ]; then
        echo "[$id] ✗ Bloqueado"
    else
        echo "[$id] ? Erro ($status_code)"
    fi
}

# Lançar requisições em paralelo
for ((i=1; i<=$CONCURRENT_REQUESTS; i++)); do
    make_request $i &
    pids+=($!)
done

# Aguardar todas as requisições terminarem
for pid in ${pids[@]}; do
    wait $pid
done

echo ""
echo "=========================================="
echo "Teste de stress concluído!"
echo "=========================================="
