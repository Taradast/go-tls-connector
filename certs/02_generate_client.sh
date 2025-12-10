#!/bin/bash
set -e

# cd certs
mkdir -p client

echo "Generating client private key..."
openssl genrsa -out client/client.key 2048

echo "Generating client CSR..."
openssl req -new -key client/client.key -out client/client.csr \
    -subj "/C=RU/ST=Chelyabinsk/L=City/O=MyOrg/OU=Client/CN=Client-001"

echo "Signing client certificate with CA..."
openssl x509 -req -in client/client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
    -out client/client.crt -days 365 -sha256

echo "Client certificate generated."
