# PivNet Resource

Interact with [Tanzu Network](https://network.tanzu.vmware.com) from concourse.

## Installing

The recommended method to use this resource is with
[resource_types](https://concourse-ci.org/resource-types.html) in the
pipeline config as follows:

```yaml
---
resource_types:
- name: pivnet
  type: docker-image
  source:
    repository: pivotalcf/pivnet-resource
    tag: latest-final
```

Using `tag: latest-final` will automatically pull the latest final release,
which can be found on the
[releases page](https://github.com/pivotal-cf/pivnet-resource/releases).

**To avoid automatically upgrading, use a fixed tag instead e.g. `tag: v0.6.3`**

Releases are semantically versioned; these correspond to the git tags in this
repository.

## Source configuration

```yaml
resources:
- name: p-mysql
  type: pivnet
  source:
    api_token: {{api-token}}
    product_slug: p-mysql
```

* `api_token`: *Required string.*

  Token from your Tanzu Network profile. Accepts either your Legacy API Token or UAA Refresh Token.

* `product_slug`: *Required string.*

  Name of product on Tanzu Network.

* `release_type`: *Optional boolean.*

  If `true`, lock to a specific Tanzu Network [release type](https://network.tanzu.vmware.com/docs/api#releases).

* `copy_metadata`: *Optional boolean.*

  If `true`, copy specified metadata from the latest All Users release within the minor releases. Defaults to `false`.
  
  The following metadata is copied:

  * Release Notes URL
  * End of General Support
  * End of Technical Guidance
  * End of Availability
  * EULA
  * License Exception
  * ECCN
  * Controlled
  * Dependency Specifiers
  * Upgrade Path Specifiers

* `endpoint`: *Optional string.*

  Endpoint to use for communicating with Tanzu Network.

  Defaults to `https://network.tanzu.vmware.com`.

* `product_version`: *Optional string.*

  Regular expression to match against product versions, e.g. `1\.2\..*`.

  Empty values match all product versions.

* `sort_by`: *Optional string.*

  Order to use for sorting releases. One of the following:

  - `none`: the order they come back from Tanzu Network.
  - `semver`: by semantic version, in descending order from the highest-valued version.
  - `last_updated`: by last updated at time, in descending order from the most recently updated version. Please note that if an earlier release is updated then the Pivnet Resource 'check' step will return it again. 

## Example pipeline configuration

See [example pipeline configurations](https://github.com/pivotal-cf/pivnet-resource/blob/master/examples).

## Behavior

### `check`: check for new product versions on Tanzu Network

Discovers all versions of the provided product.
Returned versions are optionally filtered and ordered by the `source` configuration.

### `in`: download the product from Tanzu Network

Downloads the provided product from Tanzu Network. You will be required to accept a
EULA for any product you're downloading for the first time, as well as if the terms and
conditions associated with the product change.

The metadata for the product is written to both `metadata.json` and
`metadata.yaml` in the working directory (typically `/tmp/build/get`).
Use this to programmatically determine metadata of the release.

See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata)
for more details on the structure of the metadata file.

#### Parameters

* `globs`: *Optional array.*

  Array of globs matching files to download.

  If multiple files are matched, they are all downloaded.
  
  - The globs match on the actual *file names*, not the display names in Tanzu
  Network. This is to provide a more consistent experience between uploading and
  downloading files.
  
  - If the globs fail to match any files the release download fails
  with error.
  
  - If one or more globs fails to match any files, only the matched files will be downloaded.
  
  - If `globs` is not provided (or is nil), **all files will be downloaded**.
  
  - Setting `globs` to the empty array (i.e. `globs: []`) will not attempt to
  download any files.
  
  - Files are downloaded to the working directory (e.g. `/tmp/build/get`) and the
  file names will be the same as they are on Tanzu Network - e.g. a file with
  name `some-file.txt` will be downloaded to `/tmp/build/get/some-file.txt`.

* `unpack`: *Optional boolean.*

  If `true`, unpack the downloaded file.
  
  - This can be used to use a root filesystem that is packaged as an archive file on
  network.tanzu.vmware.com as the image to run a given Concourse task.

More generally, the `unpack` parameter can be used with `get` to pass an image to a task definition,
as in the below example.

```yaml
resource:
- name: image
  type: pivnet
  source:
    api_token: {{pivnet_token}}
    product_slug: {{image-slug}}
    product_version: 0\.0\..*

jobs:
- name: sample
  serial: true
  plan:
  - get: tasks
  - get: image
    resource: pcf-automation
    params:
      globs: ["image-*.tar"]
      unpack: true

  - task: say hello
    image: image
    file: tasks/say-hello.yml
```

### `out`: upload a product to Tanzu Network

Creates a new release on Tanzu Network with the provided version and metadata.

It can also upload one or more files to Tanzu Network bucket and calculate the
MD5 checksum locally for each file in order to add MD5 checksum to the file
metadata in Tanzu Network. Note that:

* Existing product files with the same AWS key are not deleted and recreated.

* If you want to associate an existing product file with a new release, you can do so by specifying the existing AWS key when creating the release. This will not break past release associations.

* If the AWS key matches an existing file, but the hash does not match, you will receive an error and need to rename the file.

* Existing releases with the same version will _not_ be deleted and recreated by default, and will instead result in an error.

See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata)
for more details on the structure of the metadata file.

#### Parameters

* `file_glob`: *Optional string representing a glob.*

  Glob for matching files to upload and attached to release.

  If multiple files are matched by the glob, they are all uploaded. If no files
  are matched, release creation fails with an error.

* `metadata_file`: *Required string.*

  Relative path to a file containing metadata for releases and product files.

  See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata)
  for more details on the structure of the metadata file.

* `skip_product_file_polling`: *Optional boolean.*

  If `true`, skip product file validation checks after upload. Tanzu Network still validates the files asynchronously,
  but _waiting_ for the results will not happen as part of the `put:` process. **Note:** All associated product files
  in a release must still clear validation before the release can be promoted from _Admins Only_ visibility.

* `override`: *Optional boolean.*

  If `true`, forces a re-upload of releases of releases and versions that are already present on Tanzu Network.

* `files_only`: *Optional boolean.*

If `true`, requires an existing release to be passed in and uploads and attaches product files to Tanzu Network.

See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata#updating-files-only)
for more details on the structure of the metadata file for this use case.

### Some common gotchas

#### Using glob patterns instead of regex patterns

We commonly see `product_version` patterns that look something like these:

```yaml
product_version: Go*          # Go Buildpack
#....
product_version: 1\.12\.*       # ERT
```

These superficially resemble globs, not regexes — but they are regexes. These will generally
work, but not because they are a glob. They work because the regex will also match.

For example, the first pattern, `Go*` will match "Go Buildpack 1.1.1". But it would also match
"Goooooooo" or "Go Tell It On A Mountain". The second pattern, `1\.12\.*`, will match "1.12.0".
But it will also match "1.12........." and "1.12.notanumber"

Instead, try patterns like:

```yaml
product_version: Go.*\d+\.\d+\.\d+  # Go Buildpack
#....
product_version: 1\.12\.\d+         # ERT
```

Note that [the regex syntax is Go's, which is slightly limited](https://github.com/google/re2/wiki/Syntax)
compared to PCRE and other popular syntaxes.

#### Using `check-resource` for sorted but non-sequential releases (eg. Buildpacks, Stemcells)

When doing a `check`, pivnet-resource defaults to using the server-provided order. This works
fine for simple cases where the response from the server is already in `semver` order. For example,
imagine this order from a product:

```
1.12.3
1.12.2
1.12.1
1.12.0
1.11.4
1.11.3
1.11.2
1.11.1
1.11.0
```

This list is in descending semver order. All the 1.12 patch releases are together, followed
by all the 1.11 patch releases and so on.

Some products do not group into `major` or `major.minor` groups in their responses. This is usually
because a product has multiple concurrent version releases. For example, stemcells typically
have multiple major versions available. When a CVE is announced that affects them, multiple
releases may occur at once, giving an order like:

```
9999.21
7777.19
9999.20
7777.18
```

In this example, the available versions for 9999 and 7777 are _sorted_ within the list, but not _sequential_.

To fix, use `sort_by: semver` in your resource definition.

Note that buildpack "versions" are actually a name and a version combined. You'll need to escape spaces
in your `check-resource` command for it to work properly. Eg:

```
fly -t pivnet check-resource \
  --resource pivnet-resource-bug-152616708/binary-buildpack \
  --from product_version:Binary\ 1.0.11#2017-03-23T13:57:51.214Z
```

In this example we escaped the space between "Binary" and "1.0.11".

## Integration environment

The Tanzu Network team maintains an integration environment at
`https://network-integration.tanzu.vmware.com/`.

This environment is useful for teams to develop against, as changes to products
in this account are separated from the live account.

An example configuration for the integration environment might look like:

```yaml
resources:
- name: p-mysql
  type: pivnet
  source:
    api_token: {{api-token}}
    product_slug: p-mysql
    endpoint: https://network-integration.tanzu.vmware.com
```

*The integration environment does not have registry and hence operations regarding artifact references are not enabled on it.*

## Developing

### Prerequisites

A valid install of golang is required - version 1.7.x is tested; earlier
versions may also work.

### Dependencies

We use [go modules](https://github.com/golang/go/wiki/Modules) for dependencies, so you will have to make sure to turn them on with `GO111MODULE=on`.


### Running the tests

Install the ginkgo executable with:

```
go get -u github.com/onsi/ginkgo/ginkgo
```

The tests require a valid Tanzu Network API token and valid AWS S3 configuration.

Refer to the
[official docs](https://network.tanzu.vmware.com/docs/api#how-to-authenticate)
for more details on obtaining a Tanzu Network API token.

It is advised to run the acceptance tests against the Tanzu Network integration
environment endpoint `https://network-integration.tanzu.vmware.com`.

Run the tests with the following command (dummy values must be replaced by user):

```
PRODUCT_SLUG=my-product-slug-eg-tanzu-diego-pcf \
ARTIFACT_NAME=image-in-dev-registry \
ARTIFACT_PATH=path-in-dev-registry:tag \
ARTIFACT_DIGEST=sha256:letters \
API_TOKEN=my-token \
PIVNET_ENDPOINT=some-pivnet-endpoint \
PIVNET_RESOURCE_REFRESH_TOKEN=some-pivnet-resource-token \
./bin/test
```

### Contributing

Please make all pull requests to the `master` branch, and
[ensure the tests pass locally](https://github.com/pivotal-cf/pivnet-resource#running-the-tests).
