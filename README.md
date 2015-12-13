# Pivnet Resource

## Installing

The recommended way to add this resource to a Concourse instance is via the
[BOSH release](https://github.com/pivotal-cf-experimental/pivnet-resource-boshrelease)

The rootfs of the docker image is available with each release on the
[releases page](https://github.com/pivotal-cf-experimental/pivnet-resource/releases).

The docker image is `pivotalcf/pivnet-resource`; the images are available on
[dockerhub](https://hub.docker.com/r/pivotalcf/pivnet-resource).

Both the docker images and the BOSH releases are semantically versioned, but
in general their versions will be different.

## Source Configuration

* `api_token`: *Required.*  Token from your pivnet profile.

* `product_name`: *Required.*  Name of product on Pivotal Network.

### Example

Resource configuration:

``` yaml
resources:
- name: p-gitlab-pivnet-resource
  type: pivnet
  source:
    api_token: my-api-token
    product_name: p-gitlab
```

## Behavior

### `check`: Check for new product versions on Pivotal Network.

Discovers all versions of the provided product.

### `in`: Download the product from Pivotal Network.

TBD

#### Parameters

TBD

### `out`: Upload a product to Pivotal Network.

TBD

#### Parameters

TBD

## Developing

### Prerequisites

A valid install of golang >= 1.4 is required.

### Dependencies

There are no external dependencies for the resource. The tests require ginkgo
and gomega. Obtain them with:

```
go get -u github.com/onsi/ginkgo/ginkgo
go get -u github.com/onsi/gomega
go get -u github.com/golang/protobuf/proto # transitive dependency of gomega
```

### Running the tests

The tests require a valid Pivotal Network API token.

Refer to the
[official docs](https://network.pivotal.io/docs/api#how-to-authenticate)
for more details.

Run the tests with the following command:

```
API_TOKEN=my-token ./scripts/test
```
