# LCM - Local Cloud Metadata


## Main Account

Will use ApplicationDefaultCredentials by default.

Make sure the main account has `roles/iam.serviceAccountTokenCreator` on the project, which will propagate to service accounts, or that it has the correct privileges to grant itself the token creator role on requested service account on demand.

## TLS

```
# CA
openssl genpkey -algorithm ed25519 > ca.key
openssl req -x509 -new -nodes -key ca.key -sha256 -days 10000 -out ca.pem
# TLS cert
openssl genpkey -algorithm ed25519 > tls.key
openssl req -x509 -key tls.key -CA ca.pem -CAkey ca.key -sha256 -days 10000 -nodes \
    -out key.pem -subj '/CN=lc-metadata.kube-system.svc' \
    -addext 'subjectAltName=DNS:lc-metadata.kube-system.svc,DNS:lc-metadata.kube-system.svc.cluster.local'
```