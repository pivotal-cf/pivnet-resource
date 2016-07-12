# Pivnet Resource

Interact with [Pivotal Network](https://network.pivotal.io) from concourse.

## Installing

For Concourse versions 0.74.0 and higher, the recommended method to use this
resource is with `resource_types` in the pipeline config as follows:

```yaml
---
resource_types:
- name: pivnet
  type: docker-image
  source:
    repository: pivotalcf/pivnet-resource
    tag: latest-final
```
See [concourse docs](http://concourse.ci/configuring-resource-types.html) for more details
on adding `resource_types` to a pipeline config.

**Using `tag: latest-final` will pull the latest final release, which can be
found on the
[releases page](https://github.com/pivotal-cf-experimental/pivnet-resource/releases)**

**To avoid automatically upgrading, use a fixed tag instead e.g. `tag: v0.6.3`**

The docker image is `pivotalcf/pivnet-resource`; the images are available on
[dockerhub](https://hub.docker.com/r/pivotalcf/pivnet-resource).

For Concourse versions 0.73.0 and earlier, the recommended way to add this
resource to a Concourse instance is via the
[BOSH release](https://github.com/pivotal-cf-experimental/pivnet-resource-boshrelease)

The rootfs of the docker image is available with each release on the
[releases page](https://github.com/pivotal-cf-experimental/pivnet-resource/releases).

Both the docker images and the BOSH releases are semantically versioned;
they have the same version. These versions correspond to the git tags in this
repository and in the
[BOSH release](https://github.com/pivotal-cf-experimental/pivnet-resource-boshrelease)
repository.

BOSH releases are available on
[bosh.io](http://bosh.io/releases/github.com/pivotal-cf-experimental/pivnet-resource-boshrelease).

## Source Configuration

* `api_token`: *Required.*  Token from your pivnet profile.

* `product_slug`: *Required.*  Name of product on Pivotal Network.

* `release_type`: *Optional.*  Lock to a specific release type.

* `access_key_id`: *Optional.*  AWS access key id. Required for uploading products via `out`.

* `secret_access_key`: *Optional.*  AWS secret access key. Required for uploading products via `out`.

* `endpoint`: *Optional.*  Endpoint of Pivotal Network. Defaults to `https://network.pivotal.io`.

* `bucket`: *Optional.*  AWS S3 bucket name used by Pivotal Network. Defaults to `pivotalnetwork`.

* `region`: *Optional.* AWS S3 region where the bucket is located. Defaults to `eu-west-1`.

* `product_version`: *Optional.* Regex to match product version e.g. `1\.2\..*`. Empty values match all product versions.

* `sort_by`: *Optional.* Mechanism for sorting releases. Defaults to `none` which returns them in the order they come back from Pivotal Network.

  Other permissible values for `sort_by` include:
  - `semver` - this will order the releases by semantic version, returning the release with the highest-valued version.

**Values for the `endpoint`, `bucket` and `region` must be consistent or downloads and uploads may fail.**

For example, the default values of `endpoint: https://network.pivotal.io`,
`bucket: pivotalnetwork` and `region: eu-west-1`
are consistent with the production instance of Pivotal Network.

### Example Pipeline Configuration

#### Check

``` yaml
---
resources:
- name: stemcells
  type: pivnet
  source:
    api_token: my-api-token
    product_slug: stemcells
```

#### Get

Resource configuration as above for Check, with the following job configuration.

``` yaml
---
jobs:
- name: download-aws-and-vsphere-stemcells
  plan:
  - get: stemcells
    params:
      globs:
      - "*aws*"
      - "*vsphere*"
```

#### Put

``` yaml
---
resources:
- name: p-gitlab-pivnet
  type: pivnet
  source:
    api_token: my-api-token
    product_slug: p-gitlab
    access_key_id: my-aws-access-key-id
    secret_access_key: my-aws-secret-access-key


jobs:
- name: create-p-gitlab-pivnet
  plan:
  - put: p-gitlab-pivnet
    params:
      metadata_file: some-metadata-file
      file_glob: some-source-files/*
      s3_filepath_prefix: P-Gitlab
```

## Behavior

### `check`: Check for new product versions on Pivotal Network.

Discovers all versions of the provided product.
Returned versions are filtered by the `source` configuration.

### `in`: Download the product from Pivotal Network.

Downloads the provided product from Pivotal Network. **Any EULAs that have not
already been accepted will be automatically accepted at this point.**

The metadata for the product is written to both `metadata.json` and
`metadata.yaml` in the working directory (typically `/tmp/build/get`).
Use this to programmatically determine metadata of the release.
See [Metadata file](#metadata-file) for more details.

#### Parameters

* `globs`: *Optional.* Array of globs matching files to download.
  If multiple files are matched, they are all downloaded.
  - The globs match on the actual *file names*, not the display names in Pivotal
  Network. This is to provide a more consistent experience between uploading and
  downloading files.
  - If one or more globs fails to match any files the release download fails
  with error.
  - If `globs` is not provided, **no files will be downloaded**.
  - Files are downloaded to the working directory (e.g. `/tmp/build/get`) and the
  file names will be the same as they are on Pivotal Network - e.g. a file with
  name `some-file.txt` will be downloaded to `/tmp/build/get/some-file.txt`.
  - Downloaded files will be available to tasks under the folder specified by the
  tasks inputs. File names will be preserved.

### `out`: Upload a product to Pivotal Network.

Creates a new release on Pivotal Network with the provided version and metadata.

Also optionally uploads one or more files to the Pivotal Network bucket under
the provided `s3_filepath_prefix`, adding them both to the Pivotal Network as well as
to the newly-created release. The MD5 checksum of each file is taken locally, and
added to the file metadata in Pivotal Network.

If a product release already exists on Pivotal Network with the desired version, the
resource will exit with error without attempting to create the release or upload any
files.

#### Parameters

**Deprecation Warning**

**Parameters previously contained in separate files, like availability,
should now be provided via the metadata file.
Checking individual files for product metadata is now deprecated and will be
removed in a future release**

It is valid to provide both `file_glob` and `s3_filepath_prefix` or to provide
neither. If only one is present, release creation will fail. If neither are
present, file uploading is skipped.

If both `file_glob` and `s3_filepath_prefix` are present, then the source
configuration must also have `access_key_id` and `secret_access_key` or
release creation will fail.

* `file_glob`: *Optional.* Glob matching files to upload. If multiple files are
  matched by the glob, they are all uploaded. If no files are matched, release
  creation fails with error.

* `s3_filepath_prefix`: *Optional.* Case-sensitive prefix of the
  path in the S3 bucket.
  Generally similar to, but not the same as, `product_slug`. For example,
  a `product_slug` might be `pivotal-diego-pcf` (lower-case) but the
  `s3_filepath_prefix` could be `Pivotal-Diego-PCF` (mixed-case).

* `metadata_file`: *Optional.* File containing metadata for releases and product files. See [Metadata file](#metadata-file) for more details.

### Metadata file

Metadata is written in YAML and JSON format during `in`, and can be provided to
`out` via a YAML or JSON file.

The contents of this metadata (in YAML format) are as follows:

```yaml
---
release:
  version: "v1.0.0"
  release_type: All-In-One
  release_date: 1997-12-31
  eula_slug: "pivotal_beta_eula"
  description: |
    "wow this is a long description for this product"
  release_notes_url: http://example.com
  availability: Selected User Groups Only
  user_group_ids:
    - 8
    - 23
    - 42
  controlled: false
  eccn: "5D002"
  license_exception: "ENC Unrestricted"
  end_of_support_date: "2015-05-10"
  end_of_guidance_date: "2015-06-30"
  end_of_availability_date: "2015-07-04"
product_files:
- file: relative/path/to/some/product/file
  upload_as: some human-readable name
  description: |
    some
    multi-line
    description
dependencies:
- release:
    id: 1234
    version: v0.1.2
    product:
      id: 45
      name: Some product
```

The top-level `release` key is optional at present but will be required in a
later release, as it replaces the various files like `version_file`.

* `version`: *Required.* Version of the new release.

  Note, if sorting by
  semantic version in `source` params (i.e. `sort_by: semver`)
  then this version must be a valid semantic version.

  This is to prevent inconsistencies that would occur when creating a new
  version of a resource that cannot be discovered by the check for that resource.

* `release_type`: *Required.* See the
[official docs](https://network.pivotal.io/docs/api) for the supported types.

* `eula_slug`: *Required.* The EULA slug e.g. `pivotal_software_eula`. See the
[official docs](https://network.pivotal.io/docs/api#public/docs/api/v2/eulas.md)
for the supported values.

* `release_date`: *Optional.* Release date in the form of: `YYYY-MM-DD`.
  If it is not present, the release date will be set to the current date.

* `description`: *Optional.* Free-form description text.
  e.g.
  ```
  The description for this release.

  May contain line breaks.
  ```

* `release_notes_url`: *Optional.* The release notes URL
  e.g. `http://url.to/release/notes`.

* `availability`: *Optional.* Supported values are:
  - `Admins Only`
  - `All Users`
  - `Selected User Groups Only`

* `user_group_ids`: *Optional.* Comma-separated list of user
  group IDs. Each user group in the list will be added to the release.
  Will be used only if the availability is set to `Selected User Groups Only`.

* `controlled`: *Optional.* Boolean, defaults to `false`.

* `eccn`: *Optional.* String.

* `license_exception:`: *Optional.* String.

* `end_of_support_date`: *Optional.* Date in the form of: `YYYY-MM-DD`.

* `end_of_guidance_date`: *Optional.* Date in the form of: `YYYY-MM-DD`.

* `end_of_availability_date`: *Optional.* Date in the form of: `YYYY-MM-DD`.

The top-level `product_files` key is optional.
If provided, it is permitted to be an empty array.

Each element in `product_files` must have a non-empty value for the `file` key.
All other keys are optional. The purpose of the keys is as follows:

* `file` *Required.* Relative path to file. Must match exactly one file
  located via the out param `file_glob`, or the resource will exit with error.

* `description` *Optional.* The file description
  (also known as _File Notes_ in Pivotal Network).

* `upload_as` *Optional.* The display name for the file in Pivotal Network.
  This affects only the display name; the filename of the uploaded file remains
  the same as that of the local file.

The top-level `dependencies` key is currently write-only.

## Integration Environment

The Pivotal Network team maintain an integration environment at `https://pivnet-integration.cfapps.io/`
The credentials for this environment are the same as for production, and the
corresponding S3 bucket is `pivotal-network-staging`.

This environment is useful for teams to develop against, as changes to products
in this account are separated from the live account.

## Developing

### Prerequisites

A valid install of golang >= 1.5 is required.

### Dependencies

Dependencies are vendored in the `vendor` directory, according to the
[golang 1.5 vendor experiment](https://www.google.com/url?sa=t&rct=j&q=&esrc=s&source=web&cd=1&cad=rja&uact=8&ved=0ahUKEwi7puWg7ZrLAhUN1WMKHeT4A7oQFggdMAA&url=https%3A%2F%2Fgolang.org%2Fs%2Fgo15vendor&usg=AFQjCNEPCAjj1lnni5apHdA7rW0crWs7Zw).

If using golang 1.6, no action is required.

If using golang 1.5 run the following command:

```
export GO15VENDOREXPERIMENT=1
```

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
PIVNET_ENDPOINT= some-pivnet-endpoint \
S3_FILEPATH_PREFIX=Case-Sensitive-Path-eg-Pivotal-Diego-PCF \
./bin/test
```

### Contributing

Please make all pull requests to the `develop` branch, and
[ensure the tests pass locally](https://github.com/pivotal-cf-experimental/pivnet-resource#running-the-tests).

### Project management

The CI for this project can be found at https://sunrise.ci.cf-app.com and the
scripts can be found in the
[sunrise-ci repo](https://github.com/pivotal-cf-experimental/sunrise-ci).

The roadmap is captured in [Pivotal Tracker](https://www.pivotaltracker.com/projects/1474244).
