
# Metadata

Metadata is written in YAML and JSON format during `in`, and can be provided to
`out` via a YAML or JSON file.

The contents of this metadata (in YAML format) are as follows:

```yaml

---
release:
  version: "v1.0.0"
  id: 12345
  release_type: All-In-One
  release_date: 1997-12-31
  eula_slug: "vmware-prerelease-eula"
  description: |
    "wow this is a long description for this product (1000 chars max)"
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
  file_type: "Software"
  docs_url: "http://foobar.com/readme.html"
  system_requirements: ["spinning platters", "das blinkenlights"]
  platforms: ["Linux"]
  included_files: ["Component 1", "Another component"]
file_groups:
- id: 2345
  name: "some file group"
  product_files:
  - id: 5432
artifact_references:
- id: 4567
  name: my artifact
  artifact_path: repo_name:tag
  digest: sha256:mydigest
  description: |
    some
    multi-line
    description
  docs_url: "http://foobar.com/readme.html"
  system_requirements: ["requirement1", "requirement2"]
dependency_specifiers:
- specifier: 1.8.*
  product_slug: some-product
- specifier: ~>1.9.1
  product_slug: some-product
- specifier: 2.3.4
  product_slug: some-product
upgrade_path_specifiers:
- specifier: 0.2.*
- specifier: ~>0.0.5
- specifier: 0.2.0-build.2050
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

* `id`: *Optional.* Written during `in` and ignored during `out`.

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

* `file_version` *Optional.* The version number which will be displayed next to
  the filename in Pivotal Network. Defaults to `release.version` if not
  specified.

* `upload_as` *Optional.* The display name for the file in Pivotal Network.
  This affects only the display name; the filename of the uploaded file remains
  the same as that of the local file.

* `file_type` *Optional.* The type of file. Must be one of `Software`, `Documentation`
  or `Open Source License`. If not specified, defaults to `Software`

* `docs_url` *Optional.* A URL for documentation relevant to this file.

* `system_requirements` *Optional.* Additional list of requirements for using
  this file. For example: JDK version or system resources such as memory or storage
  requirements.

* `platforms` *Optional.* A list of platforms supported by this file. Valid values are:
  `Android`, `AWS`, `Apt-Get`, `BOSH`, `Brew`, `CentOS`, `Chef`, `GVM`, `Generic`,
  `Google Compute Engine`, `iOS`, `Linux`, `MSI`, `Maven` `Repo`, `Microsoft Hyper-V`,
  `OS X`, `OVM`, `OpenStack`, `Oracle` `Linux`, `Pivotal CF`, `Puppet`, `RHEL`,
  `RedHat KVM`, `SLES`, `Solaris`, `Ubuntu`, `VHCS`, `VM`, `Virtual Appliance`,
  `Windows`, `Windows Server`, `Yum` and `vSphere`.

* `included_files` *Optional.* A list of files or components included with this file.

## File Groups

The top-level `file_groups` key is optional.
If provided, it is permitted to be an empty array.

* `id` *Optional.* If provided, it will add the existing file group to the release.

* `name` *Optional.* Ignored if `id` is provided. Required otherwise. Creates a new file group with the name.

* `product_files` *Optional.* Ignored if `id` is provided. Creates a new file group with
  the provided product file ids attached to it. The ids listed must be of existing product files.

## Artifact References

This can only be used by products that use the harbor registry.

The top-level `artifact_references` key is optional.
If provided, it is permitted to be an empty array.

* `id` *Optional.* If provided, it will add the existing artifact reference to the release.

* `name` *Optional.* Ignored if `id` is provided. Required otherwise. Creates a new artifact reference with the name.

* `artifact_path` *Optional.* Ignored if `id` is provided. Required otherwise. Path of the artifact in harbor.

* `digest` *Optional.* Ignored if `id` is provided. Required otherwise. Digest must match the artifact digest on harbor.

* `description` *Optional.* Ignored if `id` is provided. Creates a new artifact reference with the description.

* `docs_url` *Optional.* Ignored if `id` is provided. A URL for documentation relevant to this artifact.

* `system_requirements` *Optional.* Ignored if `id` is provided. Additional list of requirements for using this artifact reference.

## Dependency Specifiers

The top-level `dependency_specifiers` key is optional.
If provided, it is permitted to be an empty array.

Each element in `dependency_specifiers` must have a non-empty value for both
the `specifier` key and the `product_slug` key.

See supported specifier formats in the [Pivnet API docs](https://network.pivotal.io/docs/api#public/docs/api/v2/release_dependency_specifiers.md)

## Upgrade Path Specifiers

The top-level `upgrade_path_specifiers` key is optional.
If provided, it is permitted to be an empty array.

Each element in `upgrade_path_specifiers` must have a non-empty value for
the `specifier` key.

See supported specifier formats in the [Pivnet API docs](https://network.pivotal.io/docs/api#public/docs/api/v2/release_upgrade_path_specifiers.md)

## Updating A Release

The contents of this metadata (in YAML format) are as follows. Only permits uploading additional files to an
existing release with availability not set to 'All Users':

```yaml

---
existing_release:
  id: 12345
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
  file_type: "Software"
  docs_url: "http://foobar.com/readme.html"
  system_requirements: ["spinning platters", "das blinkenlights"]
  platforms: ["Linux"]
  included_files: ["Component 1", "Another component"]
```
