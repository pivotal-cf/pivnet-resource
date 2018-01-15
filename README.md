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
    api_token: {{api-token}}
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
  - If the globs fail to match any files the release download fails
  with error.
  - If one or more globs fails to match any files, only the matched files will be downloaded.
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

### Some Common Gotchas

#### Using Glob Patterns Instead of Regex Patterns

We commonly see `product_version` patterns that look something like these:

```yaml
product_version: Go*          # Go Buildpack
#....
product_version: 1\.12\.*       # ERT
```

These superficially resemble Globs, not Regexes. They will generally work, but not because they
are a glob. They work because the regex will also match.

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
fine for simple cases where the response from the server is already in semver order. For example,
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

Some products do not group into major or major.minor groups in their responses. This is usually
because a product has multiple concurrent version releases. For example, Stemcells typically
have multiple major versions available. When a CVE is announced affected them, multiple releases
occur at once, giving a order like:

```
9999.21
7777.19
9999.20
7777.18
```

In this example, the Stemcell versions 9999 and 7777 are _sorted_ but not _sequential_.

To fix, use `sort_by: semver` in your resource definition.

Note: Buildpack "versions" are actually a name and a version combined. You'll need to escape spaces
in your `check-resource` command for it to work properly. Eg:

```
fly -t pivnet check-resource \
  --resource pivnet-resource-bug-152616708/binary-buildpack \
  --from product_version:Binary\ 1.0.11#2017-03-23T13:57:51.214Z
```

In this example we escaped the space between "Binary" and "1.0.11".

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
    api_token: {{api-token}}
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

The tests require a valid Pivotal Network API token and valid AWS S3 configuration.

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
AWS_ACCESS_KEY_ID=my-aws-access-key-id \
AWS_SECRET_ACCESS_KEY=my-aws-secret-access-key \
PIVNET_S3_REGION=region-of-pivnet-eg-us-east-1 \
PIVNET_BUCKET_NAME=bucket-of-pivnet-eg-pivnet-bucket \
PIVNET_ENDPOINT=some-pivnet-endpoint \
PIVNET_RESOURCE_REFRESH_TOKEN=some-pivnet-resource-token \
S3_FILEPATH_PREFIX=Case-Sensitive-Path-eg-Pivotal-Diego-PCF \
./bin/test
```

### Contributing

Please make all pull requests to the `develop` branch, and
[ensure the tests pass locally](https://github.com/pivotal-cf/pivnet-resource#running-the-tests).
