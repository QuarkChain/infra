#!/usr/bin/env bash

# ------------------------------------------------------------------------------
# Usage:
# export SIGNER_SERVER_HOST=<DNS or IP of op-signer server>
# ./tls-server.sh
# ------------------------------------------------------------------------------

set -e

VERSION="$(openssl version)"
if [[ "$VERSION" != "OpenSSL 3."* ]]; then
 echo "openssl version: ${VERSION}"
 echo "Script requires OpenSSL 3.x"
 exit 1
fi
if [ -z "$SIGNER_SERVER_HOST" ]; then 
 echo "Error: SIGNER_SERVER_HOST environment variable is not set." 
 exit 1 
fi

CA_SUBJECT="/O=SWC/CN=SWC root CA"
SERVER_SUBJECT="/O=SWC/CN=SWC op-signer"
ALT_NAME="DNS:*.beta.swc.quarkchain.io,IP:${SIGNER_SERVER_HOST}"
DAYS_VALID="3650"

SCRIPT_DIR="$(
 cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd
)"
TLS_DIR="${SCRIPT_DIR}/tls-server"
echo "TLS directory: ${TLS_DIR}"
echo "Generating mTLS credentials for server..."
mkdir -p "${TLS_DIR}"
cd "${TLS_DIR}"
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
# ------------------------------------------------------------------------------
# Generate Server Private Key and CSR
# ------------------------------------------------------------------------------
echo "Generating private key..."
openssl genpkey \
-algorithm RSA \
-out tls.key
echo "Generating TLS certificate signing request..."
openssl req \
-subj "${SERVER_SUBJECT}" \
-addext "subjectAltName=${ALT_NAME}" \
-new \
-key tls.key \
-out tls.csr
# ------------------------------------------------------------------------------
# Sign CSR with CA
# ------------------------------------------------------------------------------
echo "Signing server CSR with CA to obtain certificate..."
openssl x509 \
-req \
-extfile <(printf "subjectAltName=${ALT_NAME}") \
-days "${DAYS_VALID}" \
-in tls.csr \
-CA ca.crt \
-CAkey ca.key \
-CAcreateserial \
-out tls.crt
# Return to script directory
cd "${SCRIPT_DIR}"
echo "Done generating certificates."