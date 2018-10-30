package template

import (
	"fmt"
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
				Description: "golb patterns to exclude",
			},
			"list": &schema.Schema{
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "directory listing as list",
			},
			"map": &schema.Schema{
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "directory listing as map",
			},
		},
	}
}

func dataSourceDirRead(d *schema.ResourceData, meta interface{}) error {
	dirList, dirMap, err := renderDir(d)
	if err != nil {
		return err
	}
	d.Set("list", dirList)
	d.Set("map", dirMap)
	hash, err := generateDirHash(d.Get("source_dir").(string))
	if err != nil {
		return err
	}
	d.SetId(hash)
	return nil
}

func renderDir(d *schema.ResourceData) ([]string, map[string]string, error) {
	rootDir := d.Get("source_dir").(string)
	exclude := d.Get("exclude").(string)
	dirList := make([]string, 0)
	dirMap := make(map[string]string)
	//Thanks to https://github.com/saymedia/terraform-s3-dir
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

		dirList = append(dirList, relPath)
		dirMap[relPath] = filepath.Base(relPath)

		return nil
	})
	return dirList, dirMap, nil
}
