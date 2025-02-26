#!/usr/bin/env bash

# ------------------------------------------------------------------------------
# Usage:
# export SIGNER_CLIENT_DNS=<DNS of the client>
# ./tls-client.sh
# ------------------------------------------------------------------------------

set -e

VERSION="$(openssl version)"
if [[ "$VERSION" != "OpenSSL 3."* ]]; then
  echo "Currently installed OpenSSL version: ${VERSION}"
  echo "This script requires OpenSSL 3.x to function properly."
  exit 1
fi

CLIENT_SUBJECT="/O=SWC/CN=SWC op-signer client"
DAYS_VALID="3650"
: "${SIGNER_CLIENT_DNS:?Environment variable SIGNER_CLIENT_DNS must be set.}"

SCRIPT_DIR="$(
  cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd
)"
CA_DIR="${SCRIPT_DIR}/tls-server"

if [[ ! -f "${CA_DIR}/ca.crt" || ! -f "${CA_DIR}/ca.key" ]]; then
  echo "Error: Missing required CA files (ca.crt and ca.key) in '${CA_DIR}'."
  echo "Please ensure both files exist before proceeding."
  exit 1
fi

# ------------------------------------------------------------------------------
# Generate Client Private Key and CSR
# ------------------------------------------------------------------------------
echo "Generating client mTLS credentials..."
TLS_DIR="${SCRIPT_DIR}/tls"
echo "TLS directory: ${TLS_DIR}"
mkdir -p "${TLS_DIR}"
cd "${TLS_DIR}"

KEY_FILE="tls.key"
CSR_FILE="tls.csr"
CRT_FILE="tls.crt"

echo "Generating private key..."
openssl genpkey \
  -algorithm RSA \
  -out "${KEY_FILE}"

echo "Generating TLS certificate signing request..."
openssl req \
  -new \
  -key "${KEY_FILE}" \
  -subj "${CLIENT_SUBJECT}" \
  -addext "subjectAltName=DNS:${SIGNER_CLIENT_DNS}" \
  -out "${CSR_FILE}"

# ------------------------------------------------------------------------------
# Sign CSR with CA
# ------------------------------------------------------------------------------
echo "Signing client CSR with CA to obtain certificate..."
openssl x509 \
  -req \
  -days "${DAYS_VALID}" \
  -in "${CSR_FILE}" \
  -CA "${CA_DIR}/ca.crt" \
  -CAkey "${CA_DIR}/ca.key" \
  -CAcreateserial \
  -out "${CRT_FILE}" \
  -extfile <(printf "subjectAltName=DNS:${SIGNER_CLIENT_DNS}")

cd "${SCRIPT_DIR}"

echo "Done generating client certificates."