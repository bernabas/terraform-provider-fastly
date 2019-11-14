package fastly

import (
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// WAFSchema the WAF block schema
var wafSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"response_object": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The web firewall's response object",
			},
			"prefetch_condition": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The web firewall's prefetch condition",
			},
			"waf_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The web firewall id",
			},
		},
	},
}

func processWAF(d *schema.ResourceData, conn *gofastly.Client, v int) error {

	serviceID := d.Id()
	serviceVersion := strconv.Itoa(v)
	oldWAFVal, newWAFVal := d.GetChange("waf")

	if len(oldWAFVal.([]interface{})) > 0 && len(newWAFVal.([]interface{})) > 0 {
		wf := newWAFVal.([]interface{})[0].(map[string]interface{})
		copts, err := buildCreateWAF(wf, serviceID, serviceVersion)
		if err != nil {
			log.Printf("[DEBUG] Error building create WAF input: %s", err)
			return err
		}

		log.Printf("[DEBUG] Fastly WAF update opts: %#v", copts)
		// check if WAF exists first
		if !wafExists(conn, gofastly.GetWAFInput{
			Version: serviceVersion,
			Service: serviceID,
			ID:      wf["waf_id"].(string),
		}) {
			log.Printf("[WARN] WAF not found, creating one with update opts: %#v", copts)
			if err := createWAF(wf, conn, copts); err != nil {
				return err
			}
		}
		uopts, err := buildUpdateWAF(wf, serviceID, serviceVersion)
		if err != nil {
			log.Printf("[DEBUG] Error building update WAF input: %s", err)
			return err
		}
		_, err = conn.UpdateWAF(uopts)
		if err != nil {
			return err
		}

	} else if len(newWAFVal.([]interface{})) > 0 {
		wf := newWAFVal.([]interface{})[0].(map[string]interface{})
		opts, err := buildCreateWAF(wf, serviceID, serviceVersion)
		if err != nil {
			log.Printf("[DEBUG] Error building WAF: %s", err)
			return err
		}

		if err := createWAF(wf, conn, opts); err != nil {
			return err
		}

	} else if len(oldWAFVal.([]interface{})) > 0 {
		wf := oldWAFVal.([]interface{})[0].(map[string]interface{})

		opts := gofastly.DeleteWAFInput{
			Version: serviceVersion,
			ID:      wf["waf_id"].(string),
		}

		log.Printf("[DEBUG] Fastly WAF Removal opts: %#v", opts)
		err := conn.DeleteWAF(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func createWAF(df map[string]interface{}, conn *gofastly.Client, i *gofastly.CreateWAFInput) error {

	log.Printf("[DEBUG] Fastly WAF Addition opts: %#v", i)
	w, err := conn.CreateWAF(i)
	if err != nil {
		return err
	}
	df["waf_id"] = w.ID
	return nil
}

func wafExists(conn *gofastly.Client, i gofastly.GetWAFInput) bool {

	_, err := conn.GetWAF(&i)
	if err != nil {
		return false
	}
	return true
}

func flattenWAFs(wafList []*gofastly.WAF) []map[string]interface{} {

	var wl []map[string]interface{}
	if len(wafList) == 0 {
		return wl
	}

	w := wafList[0]
	WAFMapString := map[string]interface{}{
		"waf_id":             w.ID,
		"response_object":    w.Response,
		"prefetch_condition": w.PrefetchCondition,
	}

	// prune any empty values that come from the default string value in structs
	for k, v := range WAFMapString {
		if v == "" {
			delete(WAFMapString, k)
		}
	}
	return append(wl, WAFMapString)
}

func buildCreateWAF(WAFMap interface{}, serviceID string, ServiceVersion string) (*gofastly.CreateWAFInput, error) {
	df := WAFMap.(map[string]interface{})
	opts := gofastly.CreateWAFInput{
		Service:           serviceID,
		Version:           ServiceVersion,
		ID:                df["waf_id"].(string),
		PrefetchCondition: df["prefetch_condition"].(string),
		Response:          df["response_object"].(string),
	}
	return &opts, nil
}

func buildUpdateWAF(wafMap interface{}, serviceID string, ServiceVersion string) (*gofastly.UpdateWAFInput, error) {
	df := wafMap.(map[string]interface{})
	opts := gofastly.UpdateWAFInput{
		Service:           serviceID,
		Version:           ServiceVersion,
		ID:                df["waf_id"].(string),
		PrefetchCondition: df["prefetch_condition"].(string),
		Response:          df["response_object"].(string),
	}
	return &opts, nil
}
