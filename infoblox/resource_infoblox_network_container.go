package infoblox

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	ibclient "github.com/natemellendorf/infoblox-go-client"
)

func resourceNetworkContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetworkContainerCreate,
		Read:   resourceNetworkContainerRead,
		Update: resourceNetworkContainerUpdate,
		Delete: resourceNetworkContainerDelete,

		Schema: map[string]*schema.Schema{
			"network_view_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "Network view name available in NIOS Server.",
			},
			"network_name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of your network block.",
			},
			"cidr": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The network block in cidr format.",
			},
			"tenant_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique identifier of your tenant in cloud.",
			},
			"allocate_prefix_len": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Set parameter value>0 to allocate next available network with prefix=value from network container defined by parent_cidr.",
			},
			"parent_cidr": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The parent network container block in cidr format to allocate from.",
			},
		},
	}
}

func resourceNetworkContainerCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] %s: Beginning network container Creation", resourceNetworkContainerIDString(d))

	networkViewName := d.Get("network_view_name").(string)
	cidr := d.Get("cidr").(string)
	parent_cidr := d.Get("parent_cidr").(string)
	networkName := d.Get("network_name").(string)
	tenantID := d.Get("tenant_id").(string)
	connector := m.(*ibclient.Connector)
	prefixLen := d.Get("allocate_prefix_len").(int)

	objMgr := ibclient.NewObjectManager(connector, "Terraform", tenantID)

	var network_container *ibclient.NetworkContainer
	var err error

	if cidr == "" && parent_cidr != "" && prefixLen > 1 {
		log.Printf("[DEBUG] %s: Searching for network container: %s", resourceNetworkContainerIDString(d), parent_cidr)

		parent_network_container, err := objMgr.GetNetworkContainer(networkViewName, parent_cidr)
		if parent_network_container == nil {
			return fmt.Errorf(
				"Allocation of network container failed in network view (%s) : Parent network container %s not found.",
				networkViewName, parent_cidr)

		}

		log.Printf("[DEBUG] %s: Found parent container: %s", resourceNetworkContainerIDString(d), parent_network_container.Ref)
		log.Printf("[DEBUG] %s: Attempting to allocate next %v from parent...", resourceNetworkContainerIDString(d), prefixLen)

		network_container, err = objMgr.AllocateNetworkContainer(networkViewName, parent_cidr, uint(prefixLen), networkName)
		if err != nil {
			return fmt.Errorf("Allocation of network container failed in network view (%s) : %s", networkViewName, err)
		}

		log.Printf("[DEBUG] %s: Successfully allocated the new container: %s", resourceNetworkContainerIDString(d), network_container.Ref)

		d.Set("cidr", network_container.Cidr)
	} else if cidr != "" {
		network_container, err = objMgr.CreateNetworkContainer(networkViewName, cidr, networkName)
		if err != nil {
			return fmt.Errorf("Creation of network container failed in network view (%s) : %s", networkViewName, err)
		}
	} else {
		return fmt.Errorf("Creation of network container failed: neither cidr nor parent_cidr with allocate_prefix_len was specified.")
	}

	d.SetId(network_container.Ref)

	log.Printf("[DEBUG] %s: Creation on network container complete", resourceNetworkContainerIDString(d))
	return resourceNetworkContainerRead(d, m)
}
func resourceNetworkContainerRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] %s: Reading the required network container", resourceNetworkContainerIDString(d))

	networkViewName := d.Get("network_view_name").(string)
	tenantID := d.Get("tenant_id").(string)
	connector := m.(*ibclient.Connector)

	objMgr := ibclient.NewObjectManager(connector, "Terraform", tenantID)

	obj, err := objMgr.GetNetworkContainerwithref(d.Id())
	if err != nil {
		return fmt.Errorf("Getting Network container from network view (%s) failed : %s", networkViewName, err)
	}
	d.SetId(obj.Ref)
	log.Printf("[DEBUG] %s: Completed reading network container", resourceNetworkContainerIDString(d))
	return nil
}
func resourceNetworkContainerUpdate(d *schema.ResourceData, m interface{}) error {

	return fmt.Errorf("container updation is not supported")
}

func resourceNetworkContainerDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] %s: Beginning Deletion of network container", resourceNetworkContainerIDString(d))

	networkViewName := d.Get("network_view_name").(string)
	tenantID := d.Get("tenant_id").(string)
	connector := m.(*ibclient.Connector)

	objMgr := ibclient.NewObjectManager(connector, "Terraform", tenantID)

	_, err := objMgr.DeleteNetworkContainer(d.Id(), d.Get("network_view_name").(string))
	if err != nil {
		return fmt.Errorf("Deletion of Network container failed from network view(%s): %s", networkViewName, err)
	}
	d.SetId("")

	log.Printf("[DEBUG] %s: Deletion of network container complete", resourceNetworkContainerIDString(d))
	return nil
}

type resourceNetworkContainerIDStringInterface interface {
	Id() string
}

func resourceNetworkContainerIDString(d resourceNetworkContainerIDStringInterface) string {
	id := d.Id()
	if id == "" {
		id = "<new resource>"
	}
	return fmt.Sprintf("infoblox_network_container (ID = %s)", id)
}
