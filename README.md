[![Go Reference](https://pkg.go.dev/badge/github.com/embano1/vsphere.svg)](https://pkg.go.dev/github.com/embano1/vsphere)
[![Tests](https://github.com/embano1/vsphere/actions/workflows/tests.yaml/badge.svg)](https://github.com/embano1/vsphere/actions/workflows/tests.yaml)
[![Latest Release](https://img.shields.io/github/release/embano1/vsphere.svg?logo=github&style=flat-square)](https://github.com/embano1/vsphere/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/embano1/vsphere)](https://goreportcard.com/report/github.com/embano1/vsphere)
[![codecov](https://codecov.io/gh/embano1/vsphere/branch/main/graph/badge.svg?token=TC7MW723JO)](https://codecov.io/gh/embano1/vsphere)
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/embano1/vsphere)](https://github.com/embano1/vsphere)



# tl;dr

Convenience libraries and helpers when interacting with the vSphere API. Uses
[`govmomi`](https://github.com/vmware/govmomi/).

# Usage

## Package `client`

`client` provides constructors for vSphere SOAP and REST APIs, and a generic
`Client` which combines the different APIs and useful managers in a single
component. 

The `Client` can be created with `client.New(ctx)` and is configured via
environment variables (see below) and plain text files for the `basic_auth` *username*
and *password*.

See [example](example/) and the package
[documentation](https://pkg.go.dev/github.com/embano1/vsphere) for details.


| Variable              | Description                                                                         | Required | Example                           | Default   |
|-----------------------|-------------------------------------------------------------------------------------|----------|-----------------------------------|-----------|
| `VCENTER_URL`         | vCenter Server URL                                                                  | yes      | `https://myvc-01.prod.corp.local` | `""`      |
| `VCENTER_INSECURE`    | Ignore vCenter Server certificate warnings                                          | no       | `"true"`                          | `"false"` |
| `VCENTER_SECRET_PATH` | Directory where `username` and `password` files are located to retrieve credentials | yes      | `"./"`           | `"/var/bindings/vsphere"`   |

### Use with Kubernetes

Typically this library would be used in containerized environments, e.g.
[Kubernetes](https://kubernetes.io/), so the configuration is injected via the
application manifest.

```console
# create Kubernetes secret holding the vSphere credentials
kubectl --namespace myapp create secret generic vsphere-credentials --from-literal=username=administrator@vsphere.local --from-literal=password='P@ssW0rd'
```

A basic application manifest using this library would look similar to this example:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: myapp
  name: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - image: <your_image>
          name: app
          env:
            - name: VCENTER_INSECURE
              value: "true"
            - name: VCENTER_URL
              value: "https://myvc-01.prod.corp.local"
            - name: VCENTER_SECRET_PATH
              value: "/var/bindings/vsphere" # this is the default path
          volumeMounts:
            - name: credentials
              mountPath: /var/bindings/vsphere # this is the default path
              readOnly: true
      volumes:
        - name: credentials
          secret:
            secretName: vsphere-credentials
```
