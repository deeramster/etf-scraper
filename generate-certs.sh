#!/bin/bash
# generate-certs.sh - Генерация сертификатов для mTLS

set -e

CERTS_DIR="./certs"
DAYS_VALID=365

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}Генерация сертификатов для mTLS${NC}"
echo -e "${GREEN}==================================================${NC}"

# Создаем директорию для сертификатов
mkdir -p "$CERTS_DIR"

# 1. Генерация CA (Certificate Authority)
echo -e "\n${YELLOW}[1/3] Генерация CA сертификата...${NC}"

if [ -f "$CERTS_DIR/ca.key" ]; then
    echo -e "${RED}CA сертификат уже существует. Удалите $CERTS_DIR для пересоздания.${NC}"
    exit 1
fi

openssl genrsa -out "$CERTS_DIR/ca.key" 4096

openssl req -new -x509 -days $DAYS_VALID \
    -key "$CERTS_DIR/ca.key" \
    -out "$CERTS_DIR/ca.crt" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=ETF Scraper/OU=Security/CN=ETF Scraper CA"

echo -e "${GREEN}✓ CA сертификат создан${NC}"

# 2. Генерация серверного сертификата
echo -e "\n${YELLOW}[2/3] Генерация серверного сертификата...${NC}"

openssl genrsa -out "$CERTS_DIR/server.key" 2048

openssl req -new \
    -key "$CERTS_DIR/server.key" \
    -out "$CERTS_DIR/server.csr" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=ETF Scraper/OU=Server/CN=localhost"

# Создаем конфигурацию для SAN (Subject Alternative Names)
cat > "$CERTS_DIR/server.ext" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = etf-scraper.local
IP.1 = 127.0.0.1
EOF

openssl x509 -req -days $DAYS_VALID \
    -in "$CERTS_DIR/server.csr" \
    -CA "$CERTS_DIR/ca.crt" \
    -CAkey "$CERTS_DIR/ca.key" \
    -CAcreateserial \
    -out "$CERTS_DIR/server.crt" \
    -extfile "$CERTS_DIR/server.ext"

echo -e "${GREEN}✓ Серверный сертификат создан${NC}"

# 3. Генерация клиентского сертификата для администратора
echo -e "\n${YELLOW}[3/3] Генерация клиентского сертификата для администратора...${NC}"

read -p "Введите CN для администратора (например, admin): " ADMIN_CN
ADMIN_CN=${ADMIN_CN:-admin}

read -p "Введите Organization для администратора (например, Admins): " ADMIN_O
ADMIN_O=${ADMIN_O:-Admins}

openssl genrsa -out "$CERTS_DIR/client-${ADMIN_CN}.key" 2048

openssl req -new \
    -key "$CERTS_DIR/client-${ADMIN_CN}.key" \
    -out "$CERTS_DIR/client-${ADMIN_CN}.csr" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=${ADMIN_O}/OU=Administrators/CN=${ADMIN_CN}"

openssl x509 -req -days $DAYS_VALID \
    -in "$CERTS_DIR/client-${ADMIN_CN}.csr" \
    -CA "$CERTS_DIR/ca.crt" \
    -CAkey "$CERTS_DIR/ca.key" \
    -CAcreateserial \
    -out "$CERTS_DIR/client-${ADMIN_CN}.crt"

# Создаем PKCS12 для удобного импорта в браузер
openssl pkcs12 -export \
    -out "$CERTS_DIR/client-${ADMIN_CN}.p12" \
    -inkey "$CERTS_DIR/client-${ADMIN_CN}.key" \
    -in "$CERTS_DIR/client-${ADMIN_CN}.crt" \
    -certfile "$CERTS_DIR/ca.crt" \
    -passout pass:changeme

echo -e "${GREEN}✓ Клиентский сертификат создан${NC}"

# Получаем DN клиента
CLIENT_DN=$(openssl x509 -in "$CERTS_DIR/client-${ADMIN_CN}.crt" -noout -subject | sed 's/subject=//')

# Очистка временных файлов
rm -f "$CERTS_DIR"/*.csr "$CERTS_DIR"/*.ext "$CERTS_DIR"/*.srl

echo -e "\n${GREEN}==================================================${NC}"
echo -e "${GREEN}Сертификаты успешно созданы!${NC}"
echo -e "${GREEN}==================================================${NC}"
echo -e "\n${YELLOW}Файлы:${NC}"
echo "CA сертификат:       $CERTS_DIR/ca.crt"
echo "Серверный ключ:      $CERTS_DIR/server.key"
echo "Серверный сертификат: $CERTS_DIR/server.crt"
echo "Клиентский ключ:     $CERTS_DIR/client-${ADMIN_CN}.key"
echo "Клиентский сертификат: $CERTS_DIR/client-${ADMIN_CN}.crt"
echo "PKCS12 для браузера: $CERTS_DIR/client-${ADMIN_CN}.p12 (пароль: changeme)"

echo -e "\n${YELLOW}Distinguished Name клиента:${NC}"
echo -e "${GREEN}$CLIENT_DN${NC}"

echo -e "\n${YELLOW}Настройте переменную окружения:${NC}"
echo -e "${GREEN}export ADMIN_ALLOWED_DNS=\"$CLIENT_DN\"${NC}"

echo -e "\n${YELLOW}Для импорта в браузер:${NC}"
echo "1. Откройте настройки браузера"
echo "2. Импортируйте $CERTS_DIR/client-${ADMIN_CN}.p12"
echo "3. Введите пароль: changeme"
echo "4. Импортируйте $CERTS_DIR/ca.crt как доверенный CA"

echo -e "\n${YELLOW}Для curl:${NC}"
echo -e "${GREEN}curl --cert $CERTS_DIR/client-${ADMIN_CN}.crt --key $CERTS_DIR/client-${ADMIN_CN}.key --cacert $CERTS_DIR/ca.crt https://localhost:8443/admin/info${NC}"

chmod 600 "$CERTS_DIR"/*.key
chmod 644 "$CERTS_DIR"/*.crt

echo -e "\n${GREEN}✓ Права доступа настроены${NC}"
echo -e "${GREEN}==================================================${NC}\n"