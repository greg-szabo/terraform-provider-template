package template

import (
	"fmt"
	"reflect"
	"testing"

	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"os"
)

const datasourceTemplateDirRenderingConfig = `
data "template_dir" "dir" {
  source_dir = "%s"
  exclude = "%s"
  vars = %s
  render = %s
}
output "result" {
  value = "${data.template_dir.dir.rendered}"
}`

func mapStringInterfaceToMapStringString(in interface{}) map[string]string {
	out := make(map[string]string)
	v := reflect.ValueOf(in)
	for _, key := range v.MapKeys() {
		content := v.MapIndex(key).Interface().(string)
		out[key.String()]=content
	}
	return out
}

func TestDatasourceTemplateDirRendering(t *testing.T) {
	var cases = []struct {
		vars  string
		files map[string]testTemplate
	}{
		{
			files: map[string]testTemplate{
				"website/foo":           {"${bar}", "bar"},
				"website/monkey": {"ooh-ooh-ooh-eee-eee", "ooh-ooh-ooh-eee-eee"},
				"website/maths":         {"${1+2+3}", "6"},
			},
			vars: `{bar = "bar"}`,
		},
	}

	for _, tt := range cases {
		// Write the desired templates in a temporary directory.
		in, _, err := testTemplateDirWriteFiles(tt.files)
		if err != nil {
			t.Skipf("could not write templates to temporary directory: %s", err)
			continue
		}
		defer os.RemoveAll(in)

		// Run test case.
		r.UnitTest(t, r.TestCase{
			Providers: testProviders,
			Steps: []r.TestStep{
				{
					Config: fmt.Sprintf(datasourceTemplateDirRenderingConfig, in, "", tt.vars, "true"),
					Check: func(s *terraform.State) error {
						result := mapStringInterfaceToMapStringString(s.RootModule().Outputs["result"].Value)
						for name, file := range tt.files {
							if result[name] != file.want {
								return fmt.Errorf("template:\n%s\nvars:\n%s\ngot:\n%s\nwant:\n%s\n", file.template, tt.vars, result[name], file.want)
							}
						}
						return nil
					},
				},
			},
			CheckDestroy: func(*terraform.State) error {
				return nil
			},
		})
	}
}
