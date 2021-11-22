package kubernetes

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"

	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

func dataSourceKubernetesNamespacedResources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceKubernetesNamespacedResourcesRead,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:        schema.TypeString,
				Description: "namespace to list resource from.",
				Required:    true,
			},
			"group": {
				Type:        schema.TypeString,
				Description: "API group to which the resource belongs",
				Required:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: "version of the resource",
				Required:    true,
			},
			"resource_name": {
				Type:        schema.TypeString,
				Description: "Name of resource to list",
				Required:    true,
			},
			"resources": {
				Type:        schema.TypeList,
				Description: "List of all resource names in the specified namespace",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceKubernetesNamespacedResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientsets := meta.(kubeClientsets)

	log.Printf("[INFO] Listing resources")
	dyn, err := dynamic.NewForConfig(clientsets.config)
	if err != nil {
		return diag.FromErr(err)
	}
	r := k8sschema.GroupVersionResource{}
	r.Group = d.Get("group").(string)
	r.Version = d.Get("version").(string)
	r.Resource = d.Get("resource").(string)
	resource, err := dyn.Resource(r).Namespace(d.Get("namespace").(string)).List(ctx, v1.ListOptions{})
	if err != nil {
		return diag.FromErr(err)
	}
	names := make([]string, len(resource.Items))
	for i, el := range resource.Items {
		names[i] = el.GetName()
	}
	err = d.Set("resources", names)
	if err != nil {
		return diag.FromErr(err)
	}
	idsum := sha256.New()
	for _, v := range names {
		_, err := idsum.Write([]byte(v))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	id := fmt.Sprintf("%x", idsum.Sum(nil))
	d.SetId(id)
	return nil
}
