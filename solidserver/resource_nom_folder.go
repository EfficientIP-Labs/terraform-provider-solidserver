package solidserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/url"
	"path/filepath"
)

func resourcenomfolder() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcenomfolderCreate,
		ReadContext:   resourcenomfolderRead,
		UpdateContext: resourcenomfolderUpdate,
		DeleteContext: resourcenomfolderDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcenomfolderImportState,
		},

		Description: heredoc.Doc(`
			Folder resource allows to create and manage the folders in the SOLIDserver's Network Object Manager (NOM) module,
			a lightweight network assets repository.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the folder.",
				Computed:    true,
			},
			"path": {
				Type:        schema.TypeString,
				Description: "The path of the folder to create.",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A short description of the folder to create.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"space": {
				Type:        schema.TypeString,
				Description: "The name of the IP space associated to the folder.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the folder.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to the folder.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcenomfolderCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("nomfolder_path", d.Get("path").(string))
	parameters.Add("nomfolder_description", d.Get("description").(string))
	parameters.Add("nomfolder_site_name", d.Get("space").(string))
	parameters.Add("nomfolder_class_name", d.Get("class").(string))
	parameters.Add("nomfolder_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending creation request
	resp, body, err := s.Request("post", "rest/nom_folder_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created folder (oid): %s\n", oid))
				d.SetId(oid)
				d.Set("name", filepath.Base(d.Get("path").(string)))
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to create folder: %s (%s)", d.Get("path").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create folder: %s\n", d.Get("path").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomfolderUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomfolder_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("nomfolder_description", d.Get("description").(string))
	parameters.Add("nomfolder_site_name", d.Get("space").(string))
	parameters.Add("nomfolder_class_name", d.Get("class").(string))
	parameters.Add("nomfolder_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending the update request
	resp, body, err := s.Request("put", "rest/nom_folder_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated folder (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to update folder: %s (%s)", d.Get("path").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update folder: %s\n", d.Get("path").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomfolderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomfolder_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/nom_folder_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete folder: %s (%s)", d.Get("path").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete folder: %s", d.Get("path").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted folder (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomfolderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomfolder_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/nom_folder_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("description", buf[0]["nomfolder_description"].(string))

			if nomSpace, vnomSpaceExist := buf[0]["nomfolder_site_name"].(string); vnomSpaceExist && nomSpace != "#" {
				d.Set("space", buf[0]["nomfolder_site_name"].(string))
			}

			d.Set("class", buf[0]["nomfolder_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["nomfolder_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			for ck := range currentClassParameters {
				if rv, rvExist := retrievedClassParameters[ck]; rvExist {
					computedClassParameters[ck] = rv[0]
				} else {
					computedClassParameters[ck] = ""
				}
			}

			d.Set("class_parameters", computedClassParameters)

			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to find folder: %s (%s)\n", d.Get("path"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find folder (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find folder: %s\n", d.Get("path").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomfolderImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomfolder_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/nom_folder_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("description", buf[0]["nomfolder_description"].(string))

			if nomSpace, vnomSpaceExist := buf[0]["nomfolder_site_name"].(string); vnomSpaceExist && nomSpace != "#" {
				d.Set("space", buf[0]["nomfolder_site_name"].(string))
			}

			d.Set("class", buf[0]["nomfolder_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["nomfolder_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			for ck := range currentClassParameters {
				if rv, rvExist := retrievedClassParameters[ck]; rvExist {
					computedClassParameters[ck] = rv[0]
				} else {
					computedClassParameters[ck] = ""
				}
			}

			d.Set("class_parameters", computedClassParameters)

			return []*schema.ResourceData{d}, nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(ctx, fmt.Sprintf("Unable to import folder(oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import folder (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("Unable to find and import folder (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
