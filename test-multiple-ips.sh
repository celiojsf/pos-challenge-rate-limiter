#!/bin/bash

# Script para testar o rate limiter com múltiplos IPs simulados

BASE_URL="http://localhost:8080"
API_ENDPOINT="$BASE_URL/api/test"

echo "=========================================="
echo "Teste de Múltiplos IPs"
echo "=========================================="
echo ""

# Simular requisições de diferentes IPs
for ip in "192.168.1.100" "192.168.1.101" "192.168.1.102"; do
    echo "Testando IP: $ip"
    echo "Fazendo 12 requisições..."
    
    success=0
    blocked=0
    
    for ((i=1; i<=12; i++)); do
        response=$(curl -s -w "\n%{http_code}" -H "X-Forwarded-For: $ip" "$API_ENDPOINT")
        status_code=$(echo "$response" | tail -n 1)
        
        if [ "$status_code" = "200" ]; then
            ((success++))
            echo -n "✓"
        elif [ "$status_code" = "429" ]; then
            ((blocked++))
            echo -n "✗"
        fi
        
        sleep 0.1
    done
    
    echo ""
    echo "IP $ip: $success sucesso, $blocked bloqueadas"
    echo ""
    
    sleep 1
done

echo "=========================================="
echo "Teste concluído!"
echo "=========================================="
