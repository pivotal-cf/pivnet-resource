
# Metadata

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
  product_files:
  - id: 9283
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
  id: 9283
  upload_as: some human-readable name
  description: |
    some
    multi-line
    description
- file: another/relative/path/to/some/other/product/file
  id: 5432
  upload_as: some other human-readable name
  description: |
    some
    multi-line
    description
file_groups:
- id: 2345
  name: "some file group"
  product_files:
  - id: 5432
dependencies:
- release:
    id: 1234
    version: v0.1.2
    product:
      id: 45
      name: Some product
      slug: Some product
upgrade_paths:
- id: 2345
  version: v3.1.2
```

## Release

The top-level `release` key is required.

* `version`: *Required.* Version of the new release.

  Note, if sorting by semantic version in the `source` config
  (i.e. `sort_by: semver`) then this version must be a valid semantic version.

  Also, if `product_version` is a regex in the `source` config
  then this version must conform to that regex.

  These constraints prevent inconsistencies that would occur when creating a new
  version of a resource that cannot be discovered by the check for that resource.

* `release_type`: *Required.* See the
[official docs](https://network.pivotal.io/docs/api) for the supported types.

  Note, if filtering by `release_type` in the `source` config
  then this release type must be the same.

  This is to prevent inconsistencies that would occur when creating a new
  version of a resource that cannot be discovered by the check for that resource.

* `eula_slug`: *Required.* The EULA slug e.g. `pivotal_software_eula`.

  See the
  [official docs](https://network.pivotal.io/docs/api#public/docs/api/v2/eulas.md)
  for the supported values.

* `release_date`: *Optional.*
  Release date in the form of: `YYYY-MM-DD`.

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

* `product_files`: *Optional.* Written during `in` and ignored during `out`.

* `user_group_ids`: *Optional.* Comma-separated list of user
  group IDs.

  Each user group in the list will be added to the release.
  Will be used only if the availability is set to `Selected User Groups Only`.

* `controlled`: *Optional.* Boolean, defaults to `false`.

* `eccn`: *Optional.* String.

* `license_exception:`: *Optional.* String.

* `end_of_support_date`: *Optional.* Date in the form of: `YYYY-MM-DD`.

* `end_of_guidance_date`: *Optional.* Date in the form of: `YYYY-MM-DD`.

* `end_of_availability_date`: *Optional.* Date in the form of: `YYYY-MM-DD`.

## Product files

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

## File Groups

The top-level `file_groups` key is written to during `in` but is not read from
during `out`. Therefore it cannot be used to set file groups when creating
or updating a release.

## Dependencies

The top-level `dependencies` key is optional.
If provided, it is permitted to be an empty array.

Each element in `dependencies` must have a non-empty value for the `release` key.
Within each `release` element either:

* `id` must be present and non-zero

or:

* `version` and `product.slug` must be present and non-empty.

## Upgrade paths

The top-level `upgrade_paths` key is optional.
If provided, it is permitted to be an empty array.

Each element in `upgrade_paths` must have either:

* `id` - must be present and non-zero

or:

* `version` - must be present and non-empty.
  - Regular expressions are permitted - all matching
    upgrade paths will be added.
  - If an upgrade path matches multiple regular expressions,
    it will only be added once.
