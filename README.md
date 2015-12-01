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
