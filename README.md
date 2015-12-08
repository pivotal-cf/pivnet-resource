# Pivnet Resource

## Source Configuration

* `api_token`: *Required.*  Token from your pivnet profile.

* `resource_name`: *Required.*  Name of pivnet resource.

### Example

Resource configuration:

``` yaml
resources:
- name: p-gitlab-pivnet-resource
  type: pivnet
  source:
    api_token: my-api-token
    resource_name: p-gitlab
```

## Behavior

### `check`: Check for new resource versions.

Discovers all versions of the provided product.

### `in`: Clone the repository, at the given ref.

TBD

#### Parameters

TBD

### `out`: Push to a repository.

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

Refer to the [official
docs](https://network.pivotal.io/docs/api#how-to-authenticate) for more details.

Run the tests with the following command:

```
API_TOKEN=my-token ./scripts/test
```
