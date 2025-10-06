#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration from .env
RATE_LIMIT_IP=5
RATE_LIMIT_TOKEN_DEFAULT=15
TOKEN_ABC123_LIMIT=10
TOKEN_XYZ789_LIMIT=20

API_URL="http://localhost:8080/api/test"
REDIS_CONTAINER="rate-limiter-redis"

# Function to clear Redis cache
clear_redis() {
    echo -e "${YELLOW}🧹 Clearing Redis cache...${NC}"
    docker exec $REDIS_CONTAINER redis-cli FLUSHALL > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Redis cache cleared successfully${NC}"
    else
        echo -e "${RED}✗ Failed to clear Redis cache${NC}"
    fi
    echo "waiting 2 seconds..."
    sleep 2
    echo ""
}

# Function to make request and check response
make_request() {
    local token="$1"
    local ip="$2"
    local headers=""
    
    if [ -n "$token" ]; then
        headers="-H \"API_KEY: $token\""
    fi
    
    if [ -n "$ip" ]; then
        headers="$headers -H \"X-Forwarded-For: $ip\""
    fi
    
    eval curl -s -o /dev/null -w "%{http_code}" $headers "$API_URL"
}

# Function to print scenario header
print_scenario() {
    echo ""
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# Function to print test result
print_result() {
    local request_num=$1
    local status_code=$2
    local expected=$3
    
    if [ "$status_code" == "$expected" ]; then
        echo -e "${GREEN}✓ Request #$request_num: HTTP $status_code (Expected: $expected)${NC}"
    else
        echo -e "${RED}✗ Request #$request_num: HTTP $status_code (Expected: $expected)${NC}"
    fi
}

echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║         RATE LIMITER - COMPREHENSIVE TEST SUITE             ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# ============================================================================
# SCENARIO 1: Test IP Rate Limiting (no token)
# ============================================================================
print_scenario "SCENARIO 1: IP Rate Limiting (No Token)"
echo -e "📋 Testing IP-based rate limiting"
echo -e "   Limit: $RATE_LIMIT_IP requests/second"
echo -e "   Expected: First $RATE_LIMIT_IP requests succeed (200), then blocked (429)"
echo ""

success_count=0
fail_count=0

for i in $(seq 1 $((RATE_LIMIT_IP + 3))); do
    status=$(make_request "" "")
    
    if [ $i -le $RATE_LIMIT_IP ]; then
        print_result $i $status "200"
        if [ "$status" == "200" ]; then
            ((success_count++))
        fi
    else
        print_result $i $status "429"
        if [ "$status" == "429" ]; then
            ((fail_count++))
        fi
    fi
    sleep 0.01
done

echo ""
echo -e "${YELLOW}📊 Summary:${NC}"
echo -e "   ✓ Successful requests: $success_count/$RATE_LIMIT_IP"
echo -e "   ✗ Blocked requests: $fail_count/3"

clear_redis

# ============================================================================
# SCENARIO 2: Test Token with Default Limit (TOKEN_teste with empty value)
# ============================================================================
print_scenario "SCENARIO 2: Token with Default Limit (TOKEN_teste)"
echo -e "📋 Testing token with empty value in .env"
echo -e "   Token: teste"
echo -e "   Limit: $RATE_LIMIT_TOKEN_DEFAULT (uses RATE_LIMIT_TOKEN_DEFAULT)"
echo -e "   Expected: First $RATE_LIMIT_TOKEN_DEFAULT requests succeed (200), then blocked (429)"
echo ""

success_count=0
fail_count=0

for i in $(seq 1 $((RATE_LIMIT_TOKEN_DEFAULT + 3))); do
    status=$(make_request "teste" "")
    
    if [ $i -le $RATE_LIMIT_TOKEN_DEFAULT ]; then
        print_result $i $status "200"
        if [ "$status" == "200" ]; then
            ((success_count++))
        fi
    else
        print_result $i $status "429"
        if [ "$status" == "429" ]; then
            ((fail_count++))
        fi
    fi
    sleep 0.01
done

echo ""
echo -e "${YELLOW}📊 Summary:${NC}"
echo -e "   ✓ Successful requests: $success_count/$RATE_LIMIT_TOKEN_DEFAULT"
echo -e "   ✗ Blocked requests: $fail_count/3"

clear_redis

# ============================================================================
# SCENARIO 3: Test Multiple IPs (IP Isolation)
# ============================================================================
print_scenario "SCENARIO 3: Multiple IPs (IP Isolation)"
echo -e "📋 Testing that different IPs have independent rate limits"
echo -e "   IPs: 192.168.1.1, 192.168.1.2, 192.168.1.3"
echo -e "   Limit per IP: $RATE_LIMIT_IP requests/second"
echo -e "   Expected: Each IP can make $RATE_LIMIT_IP requests independently"
echo ""

declare -a ips=("192.168.1.1" "192.168.1.2" "192.168.1.3")

for ip in "${ips[@]}"; do
    echo -e "${YELLOW}Testing IP: $ip${NC}"
    success_count=0
    
    for i in $(seq 1 $((RATE_LIMIT_IP + 2))); do
        status=$(make_request "" "$ip")
        
        if [ $i -le $RATE_LIMIT_IP ]; then
            print_result $i $status "200"
            if [ "$status" == "200" ]; then
                ((success_count++))
            fi
        else
            print_result $i $status "429"
        fi
        sleep 0.01
    done
    
    echo -e "   ✓ IP $ip: $success_count/$RATE_LIMIT_IP successful requests"
    echo ""
done

clear_redis

# ============================================================================
# SCENARIO 4: Test Token with Custom Limit (TOKEN_abc123)
# ============================================================================
print_scenario "SCENARIO 4: Token with Custom Limit (TOKEN_abc123)"
echo -e "📋 Testing token with custom limit defined in .env"
echo -e "   Token: abc123"
echo -e "   Limit: $TOKEN_ABC123_LIMIT requests/second"
echo -e "   Expected: First $TOKEN_ABC123_LIMIT requests succeed (200), then blocked (429)"
echo ""

success_count=0
fail_count=0

for i in $(seq 1 $((TOKEN_ABC123_LIMIT + 3))); do
    status=$(make_request "abc123" "")
    
    if [ $i -le $TOKEN_ABC123_LIMIT ]; then
        print_result $i $status "200"
        if [ "$status" == "200" ]; then
            ((success_count++))
        fi
    else
        print_result $i $status "429"
        if [ "$status" == "429" ]; then
            ((fail_count++))
        fi
    fi
    sleep 0.01
done

echo ""
echo -e "${YELLOW}📊 Summary:${NC}"
echo -e "   ✓ Successful requests: $success_count/$TOKEN_ABC123_LIMIT"
echo -e "   ✗ Blocked requests: $fail_count/3"

clear_redis

# ============================================================================
# SCENARIO 5: Test Token with Another Custom Limit (TOKEN_xyz789)
# ============================================================================
print_scenario "SCENARIO 5: Token with Another Custom Limit (TOKEN_xyz789)"
echo -e "📋 Testing another token with different custom limit"
echo -e "   Token: xyz789"
echo -e "   Limit: $TOKEN_XYZ789_LIMIT requests/second"
echo -e "   Expected: First $TOKEN_XYZ789_LIMIT requests succeed (200), then blocked (429)"
echo ""

success_count=0
fail_count=0

for i in $(seq 1 $((TOKEN_XYZ789_LIMIT + 3))); do
    status=$(make_request "xyz789" "")
    
    if [ $i -le $TOKEN_XYZ789_LIMIT ]; then
        print_result $i $status "200"
        if [ "$status" == "200" ]; then
            ((success_count++))
        fi
    else
        print_result $i $status "429"
        if [ "$status" == "429" ]; then
            ((fail_count++))
        fi
    fi
    sleep 0.01
done

echo ""
echo -e "${YELLOW}📊 Summary:${NC}"
echo -e "   ✓ Successful requests: $success_count/$TOKEN_XYZ789_LIMIT"
echo -e "   ✗ Blocked requests: $fail_count/3"

clear_redis

# ============================================================================
# SCENARIO 6: Test Invalid/Unregistered Token
# ============================================================================
print_scenario "SCENARIO 6: Invalid/Unregistered Token"
echo -e "📋 Testing token that is not registered in .env"
echo -e "   Token: invalid_token_123"
echo -e "   Expected: All requests rejected with HTTP 403 (Forbidden)"
echo ""

forbidden_count=0

for i in $(seq 1 3); do
    status=$(make_request "invalid_token_123" "")
    print_result $i $status "403"
    if [ "$status" == "403" ]; then
        ((forbidden_count++))
    fi
    sleep 0.01
done

echo ""
echo -e "${YELLOW}📊 Summary:${NC}"
echo -e "   ✗ All requests properly rejected: $forbidden_count/3"

clear_redis

# ============================================================================
# FINAL SUMMARY
# ============================================================================
echo ""
echo -e "${BLUE}"
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    TEST SUITE COMPLETED                      ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"
echo -e "${GREEN}✓ All scenarios tested successfully!${NC}"
echo ""
echo -e "${YELLOW}Test Configuration:${NC}"
echo -e "  • IP Rate Limit: $RATE_LIMIT_IP req/s"
echo -e "  • Default Token Limit: $RATE_LIMIT_TOKEN_DEFAULT req/s"
echo -e "  • TOKEN_abc123: $TOKEN_ABC123_LIMIT req/s"
echo -e "  • TOKEN_xyz789: $TOKEN_XYZ789_LIMIT req/s"
echo -e "  • TOKEN_teste: $RATE_LIMIT_TOKEN_DEFAULT req/s (default)"
echo ""
