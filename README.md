# storageclass-accessor


## Contents
- [Contents](#contents)
- [Introduction](#introduction)
- [Installation](#installation)
    - [Installation by helm-charts](#installation-by-helm-charts)
    - [Quick Start](#quick-start)
    - [Accessor CR](#accessor-cr)
- [Examples](#examples)
    - [Only fieldSelector](#only-fieldselector)
    - [Only labelSelector](#only-labelselector)
    - [Both fieldSelector and labelSelector](#both-fieldselector-and-labelselector)
- [Notice](#notice)

# Introduction

The storageclass-accessor webhook is an HTTP callback which responds to admission requests.
When creating and deleting the PVC, it will take out the accessor related to this storage class, and the request will be allowed only when all accessors pass the verification.
Users can create accessors and set `namespaceSelector` to achieve **namespace-level** management on the storage class which provisions PVC.

# Installation

## Installation by helm-charts
```shell
helm install --create-namespace --namespace storageclass-accessor storageclass-accessor main/storageclass-accessor
```

See the Chart [README.md](./charts/storageclass-accessor/README.md) for detailed documentation on the Helm Chart
## Quick start

The guide describes how to deploy a storageclass-accessor webhook to a cluster and provides an example accessor based on csi-qingcloud.
### 1. Install CRD and CR
```shell
kubectl create -f  client/config/crds
```

### 2. Create certificate and secret
```bash
# This script will create a TLS certificate signed by the [cluster]It will place the public and private key into a secret on the cluster.
./deploy/create-cert.sh --service storageclass-accessor-service --secret accessor-validation-secret --namespace default # Make sure to use a different namespace
```


### 3. Patch the `ValidatingWebhookConfiguration` file from the template and fill in the CA bundle field
```shell
cat ./deploy/pvc-accessor-configuration-template | ./deploy/patch-ca-bundle.sh > ./deploy/pvc-accessor-configuration.yaml
```

### 4. Deploy
```shell
kubectl apply -f deploy
```

### 5. Write a CR
Create your accessor according to your needs by referring to [Accessor CR](#accessor-cr) and [Examples](#Examples).

### 6. Apply CR
Use the `kubectl apply` command to make the accessor you created operational.

### 7. Test
Now you can try to create a PVC. If it is created in a namespace that is not allowed, the following error will be output:

> Error from server: error when creating "PVC.yaml": admission webhook "pvc-accessor.storage.kubesphere.io" denied the request: The storageClass: **StorageClassName** does not allowed CREATE persistentVolumeClaim **PVC-NAME** in the namespace: **TARGET-NS**

## Accessor CR

A complete accessor should have the following fields:


- `spec.storageClassName`

  The accessor knows the effective sc according to this field.


- `spec.namespaceSelector`

  This field is used to fill in the limit of nameSpace, Including **labelSelector** and **fieldSelector**.


- `spec.namespaceSelector.fieldSelector`

  It is an **array of fieldExpressions** that manages whether nameSpace is available through the label of nameSpace.


- `fieldExpressions`

  It is an **array of fieldRule**. Every rule in the array needs to be verified.

  `labelRule` has the following fields:

      1.field: String. Required. Currently supports selection through the "Name" and "Status" fields.
      2.operator: String. Required. Currently supports selection through the "In" and "NotIn" fields.
      2.values: []String. Required. 

- `spec.namespaceSelector.labelSelector`

  It is an **array of matchExpressions** that manages whether nameSpace is available through the label of nameSpace.


- `spec.namespaceSelector.labelSelector.matchExpressions`

  It is an **array of labelRule**. Every rule in the array needs to be verified.

  `labelRule` has the following fields:

      1.key: String. Required. Currently supports selection through the "Name" and "Status" fields.
      2.operator: String. Required. Currently supports selection through the "In" and "NotIn" fields.
      2.values: []String. Required. 


# Examples

The following few examples of yaml may be helpful for you to design your own accessor.
### Only fieldSelector

- Only one `fieldExpression`
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: onlyFieldSelector-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    fieldSelector:
      - fieldExpressions:
          - field: "Name"
            operator: "In"
            values: ["NS1"]
```
After applying this accessor, you can create the PVC of csi-qingcloud only in namespace.name which in this array :["NS1"].


More than one fieldExpressions are allowed in a `fieldSelector`.

And multiple rules are also allowed in `fieldExpressions`.

- Multiple `fieldExpressions`
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: multipleFieldExpressions-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    fieldSelector:
      - fieldExpressions:
          - field: "Name"
            operator: "In"
            values: ["NS1"]
      - fieldExpressions:
          - field: "Name"
            operator: "In"
            values: ["NS2", "NS3"]
```
You can create the PVC of csi-qingcloud in the following namespace: (nameSpace.Name in ["NS1"]) **or** (nameSpace.Name in ["NS2", "NS3"]).

- Multiple rules in one `fieldExpressions`
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: multipleFieldExpressions-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    fieldSelector:
      - fieldExpressions:
          - field: "Name"
            operator: "NotIn"
            values: ["NS1", "NS2"]
          - field: "Status"
            operator: "In"
            values: ["Active"]
```
You can create the PVC of csi-qingcloud only in the following namespace: (nameSpace.Name NotIn ["NS1", "NS2"]) **and** (nameSpace.Status.Status in ["Active"])

It means that the rules in `fieldExpressions` must be followed at the same time.

### Only labelSelector

- Only one `matchExpressions`
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: csi-qingcloud-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    labelSelector:
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["app1", "app2"]
```
This requires nameSpace to have the key "app" label and the value in this array: ["app1", "app2"].


- Multiple `matchExpressions`
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: multipleFieldExpressions-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    labelSelector:
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["app1", "app2"]
      - matchExpressions:
          - key: "owner"
            operator: "In"
            values: ["owner1", "owner2"]
```
You can create the PVC of csi-qingcloud in the following namespace: (have the key "app" label and the value in ["app1", "app2"]) **or** (have the key "owner" label and the value in ["owner1", "owner2"]).

- Multiple rule in one `FieldExpressions`
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: multipleFieldExpressions-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    labelSelector:
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["app1"]
          - key: "role"
            operator: "In"
            values: ["owner1", "owner2"]
```
You can create the PVC of csi-qingcloud in the following namespace: (have the key "app" label and in the value in ["app1"]) **and** (have the key "owner" label and the value in ["owner1", "owner2"]).

### Both fieldSelector and labelSelector

```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: csi-qingcloud-accessor
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    fieldSelector:
      - fieldExpressions:
          - field: "Name"
            operator: "In"
            values: ["NS1", "NS2"]
      - fieldExpressions:
          - field: "Status"
            operator: "In"
            values: ["Active"]
    labelSelector:
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["app1"]
          - key: "owner"
            operator: "In"
            values: ["owner1", "owner2"]
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["app2", "app3"]
```
It is allowed to create PVC in a namespace that meets one of the following conditions:
- (name in ["NS1", "NS2"]) **and** (have the key "app" label and in the value in ["app1"]) **and** (have the key "owner" label and the value in ["owner1", "owner2"])
- (name in ["NS1", "NS2"]) **and** (have the key "app" label and in the value in ["app2", "app3"])
- (status.Status in ["Active"]) **and** (have the key "app" label and in the value in ["app1"]) **and** (have the key "owner" label and the value in ["owner1", "owner2"])
- (status.Status in ["Active"]) **and** (have the key "app" label and in the value in ["app2", "app3"])

# Notice

:warning: **Warning**: Too many accessors may cause unexpected errors in your webhook. It is recommended that one storage class should correspond to one accessor.

