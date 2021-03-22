package infoblox

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	ibclient "github.com/natemellendorf/infoblox-go-client"
)

func TestAccresourceNetworkContainer(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkContainerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccresourceNetworkContainerCreate,
				Check: resource.ComposeTestCheckFunc(
					testAccCreateNetworkContainerExists(t, "infoblox_network_container.foo", "10.10.0.0/24", "default", "demo-network"),
				),
			},
			resource.TestStep{
				Config: testAccresourceNetworkContainerUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCreateNetworkContainerExists(t, "infoblox_network_container.foo", "10.10.0.0/24", "default", "demo-network"),
				),
			},
		},
	})
}

func TestAccresourceNetworkContainer_Allocate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkContainerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccresourceNetworkContainerAllocate,
				Check: resource.ComposeTestCheckFunc(
					testAccCreateNetworkContainerExists(t, "infoblox_network_container.foo0", "10.0.0.0/24", "default", "demo-network"),
					testAccCreateNetworkContainerExists(t, "infoblox_network_container.foo1", "10.0.1.0/24", "default", "demo-network"),
				),
			},
		},
	})
}

func TestAccresourceNetworkContainer_Allocate_Fail(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkContainerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      testAccresourceNetworkContainerAllocateFail,
				ExpectError: regexp.MustCompile(`Allocation of network container failed in network view \(default\) : Parent network container 11.11.0.0/16 not found.`),
			},
		},
	})
}

func testAccCheckNetworkContainerDestroy(s *terraform.State) error {
	meta := testAccProvider.Meta()
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "infoblox_network" {
			continue
		}
		Connector := meta.(*ibclient.Connector)
		objMgr := ibclient.NewObjectManager(Connector, "terraform_test", "test")
		networkName, _ := objMgr.GetNetworkContainer("demo-network", "10.10.0.0/24")
		if networkName != nil {
			return fmt.Errorf("Network not found")
		}

	}
	return nil
}

func testAccCreateNetworkContainerExists(t *testing.T, n string, cidr string, networkViewName string, networkName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found:%s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID i set")
		}
		meta := testAccProvider.Meta()
		Connector := meta.(*ibclient.Connector)
		objMgr := ibclient.NewObjectManager(Connector, "terraform_test", "test")

		networkName, _ := objMgr.GetNetworkContainer(networkName, cidr)
		if networkName != nil {
			return fmt.Errorf("Network Container not found")
		}
		return nil
	}
}

var testAccresourceNetworkContainerCreate = fmt.Sprintf(`
resource "infoblox_network_container" "foo"{
	network_view_name="default"
	network_name="demo-network"
	cidr="10.10.0.0/24"
	tenant_id="foo"
	}`)

var testAccresourceNetworkContainerAllocate = fmt.Sprintf(`
resource "infoblox_network_container" "foo0"{
	network_view_name="default"
	network_name="demo-network"
	tenant_id="foo"
	allocate_prefix_len=24
	parent_cidr="10.0.0.0/16"
	}
resource "infoblox_network_container" "foo1"{
	network_view_name="default"
	network_name="demo-network"
	tenant_id="foo"
	allocate_prefix_len=24
	parent_cidr="10.0.0.0/16"
	}`)

/* Network container 11.11.0.0 should NOT exists to pass this test */
var testAccresourceNetworkContainerAllocateFail = fmt.Sprintf(`
resource "infoblox_network_container" "foo0"{
	network_view_name="default"
	network_name="demo-network"
	tenant_id="foo"
	allocate_prefix_len=24
	parent_cidr="11.11.0.0/16"
	}`)

var testAccresourceNetworkContainerUpdate = fmt.Sprintf(`
resource "infoblox_network_container" "foo"{
	network_view_name="default"
	network_name="demo-network"
	cidr="10.10.0.0/24"
	tenant_id="foo"
	}`)
