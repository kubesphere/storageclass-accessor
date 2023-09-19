#!/bin/bash
# File originally from https://github.com/banzaicloud/admission-webhook-example/blob/blog/deployment/webhook-create-signed-cert.sh

set -e

usage() {
    cat <<EOF
Generate certificate suitable for use with a webhook service.
This script uses k8s' CertificateSigningRequest API to a generate a
certificate signed by k8s CA suitable for use with webhook
services. This requires permissions to create and approve CSR. See
https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
detailed explantion and additional instructions.
The server key/cert k8s CA cert are stored in a k8s secret.
usage: ${0} [OPTIONS]
The following flags are required.
       --service          Service name of webhook.
       --namespace        Namespace where webhook service and secret reside.
       --secret           Secret name for CA certificate and server certificate/key pair.
EOF
    exit 1
}

while [[ $# -gt 0 ]]; do
    case ${1} in
        --service)
            service="$2"
            shift
            ;;
        --secret)
            secret="$2"
            shift
            ;;
        --namespace)
            namespace="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done

[ -z ${service} ] && service=storage-class-accessor
[ -z ${secret} ] && secret=storage-class-accessor-certs
[ -z ${namespace} ] && namespace=default

if [ ! -x "$(command -v openssl)" ]; then
    echo "openssl not found"
    exit 1
fi

kubectl -n ${namespace} delete secret storage-class-accessor --ignore-not-found=true

basedir="$(dirname "$(readlink -f "$0")")"
keydir="$(mktemp -d)"

cat <<EOF >> ${keydir}/server.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF

# Generate the CA cert and private key
openssl req -nodes -new -x509 -keyout ${keydir}/ca.key -out ${keydir}/ca.crt -subj "/CN=${service}.${namespace}.svc"
# Generate the private key for the webhook server
openssl genrsa -out ${keydir}/webhook-server-tls.key 2048
# Generate a Certificate Signing Request (CSR) for the private key, and sign it with the private key of the CA.
openssl req -new -key ${keydir}/webhook-server-tls.key -subj "/CN=${service}.${namespace}.svc" -config ${keydir}/server.conf \
    | openssl x509 -req -CA ${keydir}/ca.crt -CAkey ${keydir}/ca.key -CAcreateserial -out ${keydir}/webhook-server-tls.crt -extensions v3_req -extfile ${keydir}/server.conf

# Create the TLS secret for the generated keys.
kubectl -n ${namespace} create secret tls storage-class-accessor \
    --cert "${keydir}/webhook-server-tls.crt" \
    --key "${keydir}/webhook-server-tls.key"

ca_pem_b64="$(openssl base64 -A <"${keydir}/ca.crt")"
cat "${basedir}/webhook-deployment-template" | sed -e 's@${CA_BUNDLE}@'"$ca_pem_b64"'@g' | sed -e 's@${NAMESPACE}@'"$namespace"'@g' | sed -e 's@${SERVICE}@'"$service"'@g' > "${basedir}/webhook-deployment.yaml"

echo "${basedir}/webhook-deployment.yaml generated"

# Delete the key directory to prevent abuse (DO NOT USE THESE KEYS ANYWHERE ELSE).
rm -rf "$keydir"
