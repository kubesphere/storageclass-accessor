# StorageClass-accessor
***
## 
The storageclass-accessor webhook is an HTTP callback which responds to admission requests.
When creating and deleting the PVC, it will take out the accessor related to this storageclass, and the request will be allowed only when all accessors pass the verification.
Users can create accessor and set namespaceSelector to achieve **namespace-level** management on the StorageClass to create pvc

## Quick Start
***
The guide shows how to deploy StorageClass accessor webhook to the cluster. And provides an example accessor about csi-qingcloud.
### 1.install CRD and CR
```shell
kubectl create -f  client/config/crds
```

### 2.create cert and secret
```bash
# This script will create a TLS certificate signed by the [cluster]It will place the public and private key into a secret on the cluster.
./deploy/create-cert.sh --service storageclass-accessor-service --secret accessor-validation-secret --namespace default # Make sure to use a different namespace
```
Move cert.pem and key.pem to the path "/etc/storageclass-accessor-webhook/certs"


### 3.Patch the `ValidatingWebhookConfiguration` file from the template, filling in the CA bundle field.
```shell
cat ./deploy/pvc-accessor-configuration-template | ./deploy/patch-ca-bundle.sh > ./deploy/pvc-accessor-configuration.yaml
```

### 4.build docker images
```shell
docker build --network host -t kubespheredev/storageclass-accessor:v1.0 .
```

### 5.deploy 
```shell
kubectl apply -f deploy
```

### 6.apply example CR
```shell
kubectl apply -f example
```

## Next
***
If you need to customize the accessor rules, write the yaml file according to the example yaml and apply it to the cluster.The accessor rule will work when the StorageClass is requested

## Example
***
This example yaml shows how to set the accessor, you can define your own namespaceSelector according to your needs
```yaml
apiVersion: storage.kubesphere.io/v1alpha1
kind: Accessor
metadata:
  name: csi-qingcloud-accessor-example
spec:
  storageClassName: "csi-qingcloud"
  namespaceSelector:
    fieldSelector:
      - fieldExpressions:
          - field: "Name"
            operator: "In"
            values: ["default"]
    labelSelector:
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["test-app"]
          - key: "role"
            operator: "In"
            values: ["admin", "user"]
      - matchExpressions:
          - key: "app"
            operator: "In"
            values: ["test-app2"]
```

- 1. When there are multiple rules in a fieldExpressions or matchExpressions, all the rules need to pass the verification to pass.
- 2. If there are multiple fieldExpressions, only one of them needs to pass, and matchExpressions are the same.
- 3. When both the fieldSelector and labelSelector pass, the namespaceSelector is judged to pass.
- 4. If a StorageClass is mentioned by multiple accessors, it needs to pass all accessor rules.
-  :warning:  Too many accessors may cause unexpected errors in the webhook. It is recommended that one storageClass corresponds to one accessor.