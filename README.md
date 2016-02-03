# Pivnet Resource

## Installing

The recommended way to add this resource to a Concourse instance is via the
[BOSH release](https://github.com/pivotal-cf-experimental/pivnet-resource-boshrelease)

The rootfs of the docker image is available with each release on the
[releases page](https://github.com/pivotal-cf-experimental/pivnet-resource/releases).

The docker image is `pivotalcf/pivnet-resource`; the images are available on
[dockerhub](https://hub.docker.com/r/pivotalcf/pivnet-resource).

Both the docker images and the BOSH releases are semantically versioned;
they have the same version. These versions correspond to the tags in this
repository and in the
[BOSH release](https://github.com/pivotal-cf-experimental/pivnet-resource-boshrelease)
repository.

BOSH releases are available on
[bosh.io](http://bosh.io/releases/github.com/pivotal-cf-experimental/pivnet-resource-boshrelease).

## Source Configuration

* `api_token`: *Required.*  Token from your pivnet profile.

* `product_slug`: *Required.*  Name of product on Pivotal Network.

* `access_key_id`: *Optional.*  AWS access key id. Required for uploading products via `out`.

* `secret_access_key`: *Optional.*  AWS secret access key. Required for uploading products via `out`.

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
      version_file: some-metadata-files/version
      release_type_file: some-metadata-files/release_type
      eula_slug_file: some-metadata-files/eula_slug
      file_glob: some-source-files/*
      s3_filepath_prefix: P-Gitlab
```

## Behavior

### `check`: Check for new product versions on Pivotal Network.

Discovers all versions of the provided product.

### `in`: Download the product from Pivotal Network.

Downloads the provided product from Pivotal Network. **Any EULAs that have not
already been accepted will be automatically accepted at this point.**

#### Parameters

* `globs`: *Optional.* Array of globs matching files to download.
  If multiple files are matched, they are all downloaded. If one or more globs
  fails to match any files the release download fails with error.
  The globs match on the actual *file names*, not the display names in Pivotal
  Network. This is to provide a more consistent experience between uploading and
  downloading files.
  If `globs` is not provided, no files will be downloaded.

### `out`: Upload a product to Pivotal Network.

Creates a new release on Pivotal Network with the provided version and metadata.

Also optionally uploads one or more files to the Pivotal Network bucket under
the provided `s3_filepath_prefix`, adding them both to the Pivotal Network as well as to
the newly-created release.

#### Parameters

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
  `s3_filepath_prefix` could be `Pivotal-Diego-PCF`.

* `version_file`: *Required.* File containing the version string.
  Will be read to determine the new release version.

* `release_type_file`: *Required.* File containing the release type.
  Will be read to determine the release type. Valid file contents are:
  - All-In-One
  - Major Release
  - Minor Release
  - Service Release
  - Maintenance Release
  - Security Release

* `release_date_file`: *Optional.* File containing the release date in the form
  `YYYY-MM-DD`.
  If it is not present, the release date will be set to the current date.

* `eula_slug_file`: *Required.* File containing the EULA slug
  e.g. `pivotal_software_eula`

* `description_file`: *Optional.* File containing the free-form description text.
  e.g.
  ```
  The description for this release.

  May contain line breaks.
  ```

* `release_notes_url_file`: *Optional.* File containing the release notes URL
  e.g. `http://url.to/release/notes`

* `availability`: *Optional.* File containing the availability.
  Will be read to determine the availability. Valid file contents are:
  - Admins Only
  - All Users
  - Selected User Groups Only

## Developing

### Prerequisites

A valid install of golang >= 1.4 is required.

### Dependencies

Dependencies are vendored using [godep](https://github.com/tools/godep).

Restore the dependencies with:

```
godep restore
```

### Running the tests

Install godep and the ginkgo executable with:

```
go get -u github.com/tools/godep
go get -u github.com/onsi/ginkgo/ginkgo
```

The tests require a valid Pivotal Network API token and valid AWS S3 configuration.

Refer to the
[official docs](https://network.pivotal.io/docs/api#how-to-authenticate)
for more details on obtaining a Pivotal Network API token.

The tests also require that you build the s3 resource out as a binary.
The source for that can be found [here](https://github.com/concourse/s3-resource).
`S3_OUT_LOCATION` should be set to the location of the compiled binary.

For the AWS S3 configuration, as the tests will actually upload a few small test
files to the specified S3 bucket, ensure the bucket is already created and
permissions are set correctly such that the user associated with the provided
credentials can upload, download and delete.

Run the tests with the following command:

```
API_TOKEN=my-token \
AWS_ACCESS_KEY_ID=my-aws-access-key-id \
AWS_SECRET_ACCESS_KEY=my-aws-secret-access-key \
S3_OUT_LOCATION=location-of-s3-out-binary \
PIVNET_S3_REGION=region-of-pivnet-eg-us-east-1 \
PIVNET_BUCKET_NAME=bucket-of-pivnet-eg-pivnet-bucket \
S3_FILEPATH_PREFIX=Case-Sensitive-Path-eg-Pivotal-Diego-PCF \
./scripts/test
```

### Project management

The CI for this project can be found at https://sunrise.ci.cf-app.com and the
scripts can be found in the
[sunrise-ci repo](https://github.com/pivotal-cf-experimental/sunrise-ci).

The roadmap is captured in [Pivotal Tracker](https://www.pivotaltracker.com/projects/1474244).
