# Pivnet Resource

Interact with [Pivotal Network](https://network.pivotal.io) from concourse.

## Installing

The recommended method to use this resource is with
[resource_types](http://concourse.ci/configuring-resource-types.html) in the
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
[releases page](https://github.com/pivotal-cf/pivnet-resource/releases)

**To avoid automatically upgrading, use a fixed tag instead e.g. `tag: v0.6.3`**

Releases are semantically versioned; these correspond to the git tags in this
repository.

## Source Configuration

```yaml
resources:
- name: p-mysql
  type: pivnet
  source:
    username: ((my-username))
    password: ((my-password))
    product_slug: p-mysql
```

* `api_token`: *Required.*
  Token from your pivnet profile.

* `product_slug`: *Required.*
  Name of product on Pivotal Network.

* `release_type`: *Optional.*
  Lock to a specific release type.

* `copy_metadata`: *Optional.*
  Set to `true` to copy metadata from the latest All Users release within the minor. Defaults to `false`.
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

* `access_key_id`: *Optional.*
  AWS access key id.

  Required for uploading products via `out`.

* `secret_access_key`: *Optional.*
  AWS secret access key.

  Required for uploading products via `out`.

* `endpoint`: *Optional.*
  Endpoint of Pivotal Network.

  Defaults to `https://network.pivotal.io`.

* `bucket`: *Optional.*
  AWS S3 bucket name used by Pivotal Network.

  Defaults to `pivotalnetwork`.

* `region`: *Optional.*
  AWS S3 region where the bucket is located.

  Defaults to `eu-west-1`.

* `product_version`: *Optional.*
  Regex to match product version e.g. `1\.2\..*`.

  Empty values match all product versions.

* `sort_by`: *Optional.*
  Mechanism for sorting releases.

  Defaults to `none` which returns them in the order they come back from Pivotal Network.

  Other permissible values for `sort_by` include:
  - `semver` - this will order the releases by semantic version,
    returning the release with the highest-valued version.

**Values for the `endpoint`, `bucket` and `region` must be consistent
or downloads and uploads may fail.**

For example, the default values of `endpoint: https://network.pivotal.io`,
`bucket: pivotalnetwork` and `region: eu-west-1`
are consistent with the production instance of Pivotal Network.

## Credential Security

We recommend that you not check your Pivotal Network credentials into any Git repo.  Instead, please use [template variables](http://concourse.ci/fly-set-pipeline.html#parameters) or the Concourse [Vault integration](http://concourse.ci/creds.html).

## Example Pipeline Configuration

See [example pipeline configurations](https://github.com/pivotal-cf/pivnet-resource/blob/master/examples).

## Behavior

### `check`: Check for new product versions on Pivotal Network.

Discovers all versions of the provided product.
Returned versions are optionally filtered and ordered by the `source` configuration.

### `in`: Download the product from Pivotal Network.

Downloads the provided product from Pivotal Network. You will be required to accept a EULA for any product you're downloading for the first time.

The metadata for the product is written to both `metadata.json` and
`metadata.yaml` in the working directory (typically `/tmp/build/get`).
Use this to programmatically determine metadata of the release.

See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata)
for more details on the structure of the metadata file.

#### Parameters

* `globs`: *Optional.* Array of globs matching files to download.

  If multiple files are matched, they are all downloaded.
  - The globs match on the actual *file names*, not the display names in Pivotal
  Network. This is to provide a more consistent experience between uploading and
  downloading files.
  - If one or more globs fails to match any files the release download fails
  with error.
  - If `globs` is not provided (or is nil), **all files will be downloaded**.
  - If `globs` is not provided (or is nil), and there are no files to download
  - Setting `globs` to the empty array (i.e. `globs: []`) will not attempt to
  download any files.
  - Files are downloaded to the working directory (e.g. `/tmp/build/get`) and the
  file names will be the same as they are on Pivotal Network - e.g. a file with
  name `some-file.txt` will be downloaded to `/tmp/build/get/some-file.txt`.

### `out`: Upload a product to Pivotal Network.

Creates a new release on Pivotal Network with the provided version and metadata.

Also optionally uploads one or more files to Pivotal Network bucket under
the provided `s3_filepath_prefix`, adding them both to Pivotal Network as well as
to the newly-created release. The MD5 checksum of each file is taken locally,
and added to the file metadata in Pivotal Network.

**Existing product files with the same AWS key will be deleted and recreated.**

**Existing releases with the same version will be deleted and recreated.**

See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata)
for more details on the structure of the metadata file.

#### Parameters

It is valid to provide both `file_glob` and `s3_filepath_prefix` or to provide
neither. If only one is present, release creation will fail. If neither are
present, file uploading is skipped.

If both `file_glob` and `s3_filepath_prefix` are present, then the source
configuration must also have `access_key_id` and `secret_access_key` or
release creation will fail.

* `file_glob`: *Optional.* Glob matching files to upload.

  If multiple files are matched by the glob, they are all uploaded.
  If no files are matched, release creation fails with error.

* `s3_filepath_prefix`: *Optional.* Case-sensitive prefix of the
  path in the S3 bucket. If the value for `s3_filepath_prefix` starts with
  anything other than `product_files` or `product-files`, it will be prefixed
  with `product_files`.

  Generally related to `product_slug`. For example, a `product_slug` might be
  `pivotal-diego-pcf` (lower-case) but the corresponding `s3_filepath_prefix`
  could be `product-files/Pivotal-Diego-PCF` (mixed-case).

* `metadata_file`: *Optional.*
  File containing metadata for releases and product files.

  See [metadata](https://github.com/pivotal-cf/pivnet-resource/blob/master/metadata)
  for more details on the structure of the metadata file.

## Integration Environment

The Pivotal Network team maintain an integration environment at
`https://pivnet-integration.cfapps.io/`.
The credentials for this environment are the same as for production, and the
corresponding S3 bucket is `pivotal-network-staging`.

This environment is useful for teams to develop against, as changes to products
in this account are separated from the live account.

Example configuration for the integration environment:

```yaml
resources:
- name: p-mysql
  type: pivnet
  source:
    username: ((my-username))
    password: ((my-password))
    product_slug: p-mysql
    endpoint: https://pivnet-integration.cfapps.io
    bucket: pivotal-network-staging
```

## Developing

### Prerequisites

A valid install of golang is required - version 1.7.x is tested; earlier
versions may also work.

### Dependencies

Dependencies are vendored in the `vendor` directory, according to the
[golang 1.5 vendor experiment](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0ahUKEwi7puWg7ZrLAhUN1WMKHeT4A7oQFggdMAA&url=https%3A%2F%2Fgolang.org%2Fs%2Fgo15vendor&usg=AFQjCNEPCAjj1lnni5apHdA7rW0crWs7Zw).

No action is require to fetch the vendored dependencies.

### Running the tests

Install the ginkgo executable with:

```
go get -u github.com/onsi/ginkgo/ginkgo
```

The tests require valid Pivotal Network credentials and valid AWS S3 configuration.

Refer to the
[official docs](https://network.pivotal.io/docs/api#how-to-authenticate)
for more details on obtaining a Pivotal Network API token.

For the AWS S3 configuration, as the tests will actually upload a few small test
files to the specified S3 bucket, ensure the bucket is already created and
permissions are set correctly such that the user associated with the provided
credentials can upload, download and delete.

It is advised to run the acceptance tests against the Pivotal Network integration
environment endpoint `https://pivnet-integration.cfapps.io` and to use the
corresponding S3 bucket `pivotal-network-staging`.

Run the tests with the following command:

```
PRODUCT_SLUG=my-product-slug-eg-pivotal-diego-pcf \
API_TOKEN=my-token \
PIVNET_RESOURCE_USERNAME=my-email \
PIVNET_RESOURCE_PASSWORD=my-password \
AWS_ACCESS_KEY_ID=my-aws-access-key-id \
AWS_SECRET_ACCESS_KEY=my-aws-secret-access-key \
PIVNET_S3_REGION=region-of-pivnet-eg-us-east-1 \
PIVNET_BUCKET_NAME=bucket-of-pivnet-eg-pivnet-bucket \
PIVNET_ENDPOINT= some-pivnet-endpoint \
S3_FILEPATH_PREFIX=Case-Sensitive-Path-eg-Pivotal-Diego-PCF \
./bin/test
```

### Contributing

Please make all pull requests to the `master` branch, and
[ensure the tests pass locally](https://github.com/pivotal-cf/pivnet-resource#running-the-tests).
