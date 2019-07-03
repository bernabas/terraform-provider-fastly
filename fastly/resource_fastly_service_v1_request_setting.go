package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform/helper/schema"
	"strings"
)

var requestSettingSchema = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			// Required fields
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique name to refer to this Request Setting",
			},
			// Optional fields
			"request_condition": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Name of a request condition to apply. If there is no condition this setting will always be applied.",
			},
			"max_stale_age": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "How old an object is allowed to be, in seconds. Default `60`",
			},
			"force_miss": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Force a cache miss for the request",
			},
			"force_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Forces the request use SSL",
			},
			"action": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allows you to terminate request handling and immediately perform an action",
			},
			"bypass_busy_wait": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable collapsed forwarding",
			},
			"hash_keys": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Comma separated list of varnish request object fields that should be in the hash key",
			},
			"xff": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "append",
				Description: "X-Forwarded-For options",
			},
			"timer_support": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Injects the X-Timer info into the request",
			},
			"geo_headers": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Inject Fastly-Geo-Country, Fastly-Geo-City, and Fastly-Geo-Region",
			},
			"default_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "the host header",
			},
		},
	},
}

func flattenRequestSettings(rsList []*fastly.RequestSetting) []map[string]interface{} {
	var rl []map[string]interface{}
	for _, r := range rsList {
		// Convert Request Settings to a map for saving to state.
		nrs := map[string]interface{}{
			"name":              r.Name,
			"max_stale_age":     r.MaxStaleAge,
			"force_miss":        r.ForceMiss,
			"force_ssl":         r.ForceSSL,
			"action":            r.Action,
			"bypass_busy_wait":  r.BypassBusyWait,
			"hash_keys":         r.HashKeys,
			"xff":               r.XForwardedFor,
			"timer_support":     r.TimerSupport,
			"geo_headers":       r.GeoHeaders,
			"default_host":      r.DefaultHost,
			"request_condition": r.RequestCondition,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range nrs {
			if v == "" {
				delete(nrs, k)
			}
		}

		rl = append(rl, nrs)
	}

	return rl
}

func buildRequestSetting(requestSettingMap interface{}) (*fastly.CreateRequestSettingInput, error) {
	df := requestSettingMap.(map[string]interface{})
	opts := fastly.CreateRequestSettingInput{
		Name:             df["name"].(string),
		MaxStaleAge:      uint(df["max_stale_age"].(int)),
		ForceMiss:        fastly.CBool(df["force_miss"].(bool)),
		ForceSSL:         fastly.CBool(df["force_ssl"].(bool)),
		BypassBusyWait:   fastly.CBool(df["bypass_busy_wait"].(bool)),
		HashKeys:         df["hash_keys"].(string),
		TimerSupport:     fastly.CBool(df["timer_support"].(bool)),
		GeoHeaders:       fastly.CBool(df["geo_headers"].(bool)),
		DefaultHost:      df["default_host"].(string),
		RequestCondition: df["request_condition"].(string),
	}

	act := strings.ToLower(df["action"].(string))
	switch act {
	case "lookup":
		opts.Action = fastly.RequestSettingActionLookup
	case "pass":
		opts.Action = fastly.RequestSettingActionPass
	}

	xff := strings.ToLower(df["xff"].(string))
	switch xff {
	case "clear":
		opts.XForwardedFor = fastly.RequestSettingXFFClear
	case "leave":
		opts.XForwardedFor = fastly.RequestSettingXFFLeave
	case "append":
		opts.XForwardedFor = fastly.RequestSettingXFFAppend
	case "append_all":
		opts.XForwardedFor = fastly.RequestSettingXFFAppendAll
	case "overwrite":
		opts.XForwardedFor = fastly.RequestSettingXFFOverwrite
	}

	return &opts, nil
}