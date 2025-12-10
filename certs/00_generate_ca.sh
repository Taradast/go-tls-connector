#!/bin/bash
set -e

echo "Generating CA private key..."
openssl genrsa -out ca.key 4096

echo "Generating CA certificate..."
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 \
    -out ca.crt -subj "/C=RU/ST=Chelyabinsk/L=City/O=MyOrg/OU=CA/CN=MyRootCA"

echo "CA generation completed."
