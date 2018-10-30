---
layout: "template"
page_title: "Template: template_dir"
sidebar_current: "docs-template-datasource-dir"
description: |-
  Renders a list and a map of file paths.
---

# template_dir

Renders a directory and its sub-directories into a terraform list and map.

`template_dir` is similar to [`template_file`](../d/file.html) but it walks
a given source directory and treats every file it encounters as a template,
rendering it to a corresponding list and map.

~> **Note** When working with local files, Terraform will detect the resource
as having been deleted each time a configuration is applied on a new machine
where the destination dir is not present and will generate a diff to create
it. This may cause "noise" in diffs in environments where configurations are
routinely applied by many different users or within automation systems.

## Example Usage

The following example shows how one might use this resource to produce a
list of files from a directory and then upload them to Amazon S3.

```hcl
data "template_dir" "files" {
  source_dir      = "${path.module}/website"
  exclude         = "(debug/*|tmp$)"
}

resource "aws_s3_bucket" "filestore" {
  bucket = "terraform-s3-filestore-test"
}

resource "aws_s3_bucket_object" "files" {
  count = "${length(data.template_dir.files.list)}"

  source = "website/${element(data.template_dir.files.list,count.index)}"

  bucket = "${aws_s3_bucket.filestore}"
  key = "${element(data.template_dir.files.list,count.index)}"
}
```

## Argument Reference

The following arguments are supported:

* `source_dir` - (Required) Path to the directory where the files reside.

* `exclude` - (Optional) Regular expression that is applied to the relative
  path. Everything that matches the expression will be omitted from the list.

After rendering this resource remembers the content of the source directory
in the Terraform state, and will plan to recreate the upload if any changes
are detected during the plan phase.

## Attributes

This resource exports the following attributes:

* `list` - The list of files with relative path.

* `map` - The map of files with relative path. Key will contain the whole
path with the filename while Value will contain the file name. (In the future it might contain the rendered file.)
