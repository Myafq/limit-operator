This project is aimed to help Kubernetes adminitrators define resource limit ranges dynamically for all namespaces.
# Limit Operator

A Kubernetes Operator based on the Operator SDK for cluster level resource configuration.

## Current status

This is a PoC / alpha version. Most functionality is there but it is higly likely there are bugs and improvements needed

## Supported Custom Resources

The following resources are supported:

- `ClusterLimit`

### ClusterLimit

Represents a LimitRange and a namespace Selector for the Operator to install this LimitRange to.


## Test it locally

Refer to the [Operator SDK](https://github.com/operator-framework/operator-sdk) docs in order to setup local development environment and spin up operator.

## Deploying to a Cluster

Before the Limit Operator can be deployed to a running cluster, the necessary Custom Resource Definitions (CRDs) must be installed. To do this, run `kubectl apply -f deploy/crds/`.

Then run 
- `kubectl apply -f deploy/service_account.yaml`
- `kubectl apply -f deploy/role.yaml -n <NAMESPACE>`
- `kubectl apply -f deploy/role_binding.yaml`
- `kubectl apply -f deploy/operator.yaml -n <NAMESPACE>`

## Create a ClusterLimit

- `kubectl apply -f deploy/examples/keycloak_min.json`

