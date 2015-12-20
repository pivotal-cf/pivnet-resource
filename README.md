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

* `access_key_id`: *Optional.*  AWS access key id. Required for uploading products via `out`.

* `secret_access_key`: *Optional.*  AWS secret access key. Required for uploading products via `out`.

### Example Pipeline Configuration

#### Check

``` yaml
---
resources:
- name: p-gitlab-pivnet
  type: pivnet
  source:
    api_token: my-api-token
    product_name: p-gitlab
```

#### Get

Resource configuration as above for Check, with the following job configuration.

``` yaml
---
jobs:
- name: download-p-gitlab-pivnet
  plan:
  - get: p-gitlab-pivnet
```

#### Put

``` yaml
---
resources:
- name: p-gitlab-pivnet
  type: pivnet
  source:
    api_token: my-api-token
    product_name: p-gitlab
    access_key_id: my-aws-access-key-id
    secret_access_key: my-aws-secret-access-key

---
jobs:
- name: create-p-gitlab-pivnet
  plan:
  - put: p-gitlab-pivnet
    params:
      file: some-directory/*
      s3_filepath_prefix: P-Gitlab
```

## Behavior

### `check`: Check for new product versions on Pivotal Network.

Discovers all versions of the provided product.

### `in`: Download the product from Pivotal Network.

Downloads the provided product from Pivotal Network. Any EULAs must have already
been signed. Due to caching, it is advisable to sign the EULAs before the first
execution of `in` (i.e. the first `get` in the pipeline) as the resource will be
cached and therefore signing the EULA will have no effect.

#### Parameters

None.

### `out`: Upload a product to Pivotal Network.

Uploads a single file to the pivnet bucket.

#### Parameters

* `file`: *Required.* Path to the file to upload. If multiple files are
  matched by the glob, an error is raised.

* `s3_filepath_prefix`: *Required.* Case-sensitive prefix of the
  path in the S3 bucket.
  Generally similar to, but not the same as, `product_name`. For example,
  a `product_name` might be `pivotal-diego-pcf` (lower-case) but the
  `s3_filepath_prefix` could be `Pivotal-Diego-PCF`.

## Developing

### Prerequisites

A valid install of golang >= 1.4 is required.

### Dependencies

There are no external dependencies for the resource.
The test dependencies are vendored using [godep](https://github.com/tools/godep).
Install godep and the ginkgo executable with:

```
go get -u github.com/tools/godep
go get -u github.com/onsi/ginkgo/ginkgo
```

Restore dependencies with:

```
godep restore
```

### Running the tests

The tests require a valid Pivotal Network API token and valid AWS S3 configuration.

Refer to the
[official docs](https://network.pivotal.io/docs/api#how-to-authenticate)
for more details on obtaining a Pivotal Network API token.

For the AWS S3 configuration, as the tests will actually upload a small test
file to the specified S3 bucket, ensure the bucket is already created and
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
