---
layout: "template"
page_title: "Template: template_dir"
sidebar_current: "docs-template-datasource-dir"
description: |-
  Renders a directory of templates.
---

# template_dir

Renders a directory containing templates into a map of
corresponding rendered files.

`template_dir` is similar to [`template_file`](../d/file.html) but it walks
a given source directory and treats every file it encounters as a template,
rendering it to a corresponding map in memory.

Note: empty folders are omitted.

## Example Usage

The following example shows how one might use this resource to produce a
list of files from a directory and upload them to Amazon S3.

Option 1: By content:

This example will give errors if a file is empty, because `aws_s3_bucket_object.content`
cannot be empty. Use this example if you have valid configuration
files that you want to interpolate.
This example renders all files as templates before uploading them to S3.

```hcl
data "template_dir" "files" {
  source_dir      = "${path.module}/website"
  exclude         = "(debug/*|tmp$)"
  vars {
    "title"       = "My Website"
  }
}

resource "aws_s3_bucket" "filestore" {
  bucket = "terraform-s3-filestore-example1"
}

resource "aws_s3_bucket_object" "files" {
  count = "${length(keys(data.template_dir.files.rendered))}"

  bucket = "${aws_s3_bucket.filestore}"
  key = "${element(keys(data.template_dir.files.rendered),count.index)}"
  content = "${data.template_dir.files.rendered[element(keys(data.template_dir.files.rendered),count.index)]}"  
}
```

Option 2: By filename:

This example is quicker because it does not interpolate the files
and it works for empty files too.
However, it does not work when the S3 bucket is encrypted because the
`etag` property is necessary to track file changes.

```hcl
data "template_dir" "files" {
  source_dir      = "${path.module}/website"
  exclude         = "(debug/*|tmp$)"
  render          = false
}

resource "aws_s3_bucket" "filestore" {
  bucket = "terraform-s3-filestore-example2"
}

resource "aws_s3_bucket_object" "files" {
  count = "${length(keys(data.template_dir.files.rendered))}"

  bucket = "${aws_s3_bucket.filestore}"
  key = "${element(keys(data.template_dir.files.rendered),count.index)}"
  source = "${data.template_dir.files.source_dir}/${element(keys(data.template_dir.files.rendered),count.index)}"
  etag = "${md5(file(format("%s/%s",data.template_dir.files.source_dir,element(keys(data.template_dir.files.rendered),count.index))))}"
}
```

## Argument Reference

The following arguments are supported:

* `source_dir` - (Required) Path to the directory where the files reside.

* `exclude` - (Optional) Regular expression that is applied to the relative
  path. Everything that matches the expression will be omitted.

* `render` - (Optional) Specify if the files should be rendered as templates.
  Default to `true`

* `vars` - (Optional) Variables for interpolation within the templates. Note
  that variables must all be primitives. Direct references to lists or maps
  will cause a validation error.

After rendering this resource remembers the content of the source directory
in the Terraform state, and will plan to recreate the upload if any changes
are detected during the plan phase.

## Template Syntax

The syntax of the template files is the same as
[standard interpolation syntax](/docs/configuration/interpolation.html),
but you only have access to the variables defined in the `vars` section.

To access interpolations that are normally available to Terraform
configuration (such as other variables, resource attributes, module
outputs, etc.) you can expose them via `vars` as shown below:

```hcl
data "template_dir" "init" {
  # ...

  vars {
    foo  = "${var.foo}"
    attr = "${aws_instance.foo.private_ip}"
  }
}
```

## Attributes Reference

This resource exports the following attributes:

* `rendered` - A map of template files rendered. The key is the relative path
and filename from `source_dir` and the value is the rendered file. The value
is empty, if `render = false`.
