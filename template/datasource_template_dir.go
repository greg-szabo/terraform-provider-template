package template

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/hashicorp/terraform/helper/pathorcontents"
	"github.com/hashicorp/terraform/helper/schema"
	"os"
	"path/filepath"
	"regexp"
)

func dataSourceDir() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDirRead,

		Schema: map[string]*schema.Schema{
			"source_dir": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Directory to read",
			},
			"exclude": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "regular expression to exclude",
			},
			"vars": &schema.Schema{
				Type:         schema.TypeMap,
				Optional:     true,
				Default:      make(map[string]interface{}),
				Description:  "variables to substitute",
				ValidateFunc: validateVarsAttribute,
			},
			"render": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "render file content as template",
			},
			"rendered": &schema.Schema{
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "rendered files",
			},
		},
	}
}

func dataSourceDirRead(d *schema.ResourceData, meta interface{}) error {
//	dirMap, err := renderDir(d)
//	if err != nil {
//		return err
//	}
	dirMap := make(map[string]string)
	dirMap["a.b"] = "apple"
	d.Set("rendered", dirMap)
	d.SetId(generateHash(dirMap))
	return nil
}

func renderDir(d *schema.ResourceData) (map[string]interface{}, error) {
	rootDir := d.Get("source_dir").(string)
	exclude := d.Get("exclude").(string)
	render := d.Get("render").(bool)
	vars := d.Get("vars").(map[string]interface{})
	dirMap := make(map[string]interface{})
	//Thanks to https://github.com/saymedia/terraform-s3-dir for the initial idea
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %s\n", path, err)
			// Skip stuff we can't read.
			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed make %s relative: %s\n", path, err)
			return nil
		}
		path, err = filepath.EvalSymlinks(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to resolve symlink %s: %s\n", path, err)
			return nil
		}

		if info.IsDir() {
			// Don't need to create directories since they are implied
			// by the files within.
			return nil
		}

		if exclude != "" {
			if matched, _ := regexp.MatchString(exclude, relPath); matched {
				// This path should be excluded.
				return nil
			}
		}

		if render {
			data, _, err := pathorcontents.Read(path)
			if err != nil {
				return err
			}

			rendered, err := execute(data, vars)
			if err != nil {
				return templateRenderError(
					fmt.Errorf("failed to render %v: %v", path, err),
				)
			}
			dirMap[relPath] = rendered
		} else {
			dirMap[relPath] = ""
		}
		return nil
	})
	return dirMap, nil
}

func generateHash(dirMap map[string]string) string {
	var nameChecksums []byte
	var contentChecksums []byte
	//ioutil.WriteFile("debug.log", []byte(fmt.Sprintf("%v",dirMap)), 0644)
	for name, content := range dirMap {
		nameHash := sha1.Sum([]byte(name))
		contentHash := sha1.Sum([]byte(content))
		nameChecksums = append(nameChecksums, nameHash[:]...)
		contentChecksums = append(contentChecksums, contentHash[:]...)
	}
	checksum := sha1.Sum(append(nameChecksums,contentChecksums...))
	return hex.EncodeToString(checksum[:])
}
