#!/bin/bash
set -e

# cd certs
mkdir -p server

echo "Generating server private key..."
openssl genrsa -out server/server.key 2048

echo "Creating server CSR..."
openssl req -new -key server/server.key -out server/server.csr -subj "/C=RU/ST=Chelyabinsk/L=City/O=MyOrg/OU=Server/CN=localhost"

# Создаём конфиг для SAN
cat > server/server_ext.cnf <<EOL
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOL

echo "Signing server certificate with CA and SAN..."
openssl x509 -req -in server/server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out server/server.crt -days 365 -sha256 -extfile server/server_ext.cnf

echo "Server certificate generated with SAN."
