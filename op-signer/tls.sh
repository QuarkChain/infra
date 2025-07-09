#!/usr/bin/env bash

# ------------------------------------------------------------------------------
# Usage:
# ./tls.sh server|client <Domain Name>
# ------------------------------------------------------------------------------

set -e

VERSION="$(openssl version)"
if [[ "${VERSION}" != "OpenSSL 3."* ]]; then
    echo "Requires OpenSSL 3.x, current: ${VERSION}"
    exit 1
fi

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 [server|client] <Domain Name>"
    exit 1
fi

MODE=$1
DNS_NAME=$2
CA_SUBJECT="/O=SWC/CN=SWC root CA"
SUBJECT="/O=SWC/CN=SWC op-signer $MODE"
ALT_NAME="DNS:${DNS_NAME}"
DAYS_VALID="3650"
SCRIPT_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)
TLS_DIR_SERVER="${SCRIPT_DIR}/tls-server"

echo "Generating mTLS credentials for $MODE..."
mkdir -p "${TLS_DIR_SERVER}"
cd "${TLS_DIR_SERVER}"

case $MODE in
    server)
        # ------------------------------------------------------------------------------
        # Generate CA (if needed)
        # ------------------------------------------------------------------------------
        if [[ ! -f "ca.crt" ]]; then
            echo "Generating CA..."
            openssl req \
            -x509 \
            -newkey rsa:2048 \
            -days "${DAYS_VALID}" \
            -nodes \
            -keyout ca.key \
            -out ca.crt \
            -subj "${CA_SUBJECT}"
        fi
        ALT_NAME="DNS:localhost,DNS:${DNS_NAME}"
        ;;
    client)
        if [[ ! -f "ca.crt" ]]; then
            echo "CA certificate missing"
            exit 1
        fi
        TLS_DIR="${SCRIPT_DIR}/tls"
        mkdir -p "$TLS_DIR"
        cd "$TLS_DIR"
        cp "$TLS_DIR_SERVER"/ca.crt .
        ;;
    *)
        echo "Invalid mode: '$mode'. Must be 'server' or 'client'."
        exit 2
        ;;
esac

# ------------------------------------------------------------------------------
# Generate Private Key and CSR
# ------------------------------------------------------------------------------
echo "Generating private key..."
openssl genpkey \
-algorithm RSA \
-out tls.key

echo "Generating TLS certificate signing request..."
openssl req \
-subj "${SUBJECT}" \
-addext "subjectAltName=${ALT_NAME}" \
-new \
-key tls.key \
-out tls.csr

# ------------------------------------------------------------------------------
# Sign CSR with CA
# ------------------------------------------------------------------------------
echo "Signing CSR with CA to obtain certificate..."
openssl x509 \
-req \
-extfile <(printf "subjectAltName=${ALT_NAME}") \
-days "${DAYS_VALID}" \
-CA ca.crt \
-CAkey "$TLS_DIR_SERVER"/ca.key \
-CAcreateserial \
-in tls.csr \
-out tls.crt

rm tls.csr
cd "${SCRIPT_DIR}"
echo "TLS generation for $MODE completed."
