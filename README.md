# OpenFGA Kubernetes Authorizer
An implementation of a [Kubernetes Webhook Authorizer](https://kubernetes.io/docs/reference/access-authn-authz/webhook/) that uses [OpenFGA](https://github.com/openfga/openfga) to make Authorization decisions.

The OpenFGA Authorizer implements the [Authorizer interface](https://pkg.go.dev/k8s.io/apiserver/pkg/authorization/authorizer#Authorizer).

## Setup
### Certificate Generation
```shell
mkdir -p certs && cd certs

openssl genrsa -out ca.key 2048

openssl req -x509 -new -nodes -key ca.key -subj "/CN=127.0.0.1" -days 10000 -out ca.crt

openssl genrsa -out server.key 2048
openssl req -new -key tls.key -out server.csr -config csr.conf

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key \
    -CAcreateserial -out server.crt -days 10000 \
    -extensions v3_ext -extfile csr.conf -sha256
```

### Start the Kubernetes API Server and Authorizer
```shell
docker compose up
```

### Authenticate (kubectl context)
```shell
export KUBECONFIG=./.config/kubeconfig.yaml

kubectl config use-context admin-user
```

## Kubernetes Authorization Model
### Resource Attributes
* Verb - `get`, `list`, `watch`, `create`, `update`, `delete`, `proxy` (`*` means all/any)
* Namespace
* Name (ResourceName) - `secrets`, `configmaps`, `deployments`, `pods`
* APIGroup - `rbac.authorization.k8s.io`, `api`
* APIVersion

### API Resources (top-level)
* Create a new API object (create): POST `/apis/<apiGroup>/<apiVersion>/<resource>`
* List API objects (list): GET `/apis/<apiGroup>/<apiVersion>/<resource>`
* Watch API objects (watch): GET `/apis/<apiGroup>/<apiVersion>/<resource>?watch=1`
* Get an API object with a given name (get): GET `/apis/<apiGroup>/<apiVersion>/<resource>/<name>`
* Update an API object with a given name (update): PUT `/apis/<apiGroup>/<apiVersion>/<resource>/<name>`
* Patch an API object with a given name (patch): PATCH `/apis/<apiGroup>/<apiVersion>/<resource>/<name>`
* Delete an API object with a given name (delete): DELETE `/apis/<apiGroup>/<apiVersion>/<resource>/<name>`

### Namespaced Resources
Same rules for the [API Resources](#api-resources-top-level) apply, but with a subpath scoped to the namespace. For example,

`/apis/<apiGroup>/<apiVersion>/namespaces/<namespace>/<resource>`

* Get Deployment 'foo' in Namespace 'bar' - GET /apps/v1/namespaces/bar/deployments/foo

  `Check(resource:/apps/v1/namespaces/bar/deployments/foo, get, user:jon)`

### Check /authorize endpoint with TLS
```shell
curl -v --cert certs/server.crt --key certs/server.key --cacert certs/ca.crt -H 'Content-Type: application/json' https://localhost:9443/authorize 
```