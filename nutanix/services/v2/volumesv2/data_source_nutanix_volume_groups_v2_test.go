package volumesv2_test

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	acc "github.com/terraform-providers/terraform-provider-nutanix/nutanix/acctest"
)

const dataSourceVolumeGroups = "data.nutanix_volume_groups_v2.test"

func TestAccNutanixVolumeGroupsV2DataSource_Basic(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("terraform-test-volume-group-%d", r)
	desc := "terraform test volume group description"
	path, _ := os.Getwd()
	filepath := path + "/../../../../test_config_v2.json"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeGroupsDataSourceConfig(filepath, name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceVolumeGroups, "volumes.#"),
					resource.TestCheckResourceAttrSet(dataSourceVolumeGroups, "volumes.0.name"),
				),
			},
		},
	})
}

func TestAccNutanixVolumeGroupsV4DataSource_WithFilter(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("terraform-test-volume-group-%d", r)
	desc := "terraform test volume group description"
	path, _ := os.Getwd()
	filepath := path + "/../../../../test_config_v2.json"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeGroupsDataSourceWithFilter(filepath, name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.name", name),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.description", desc),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.sharing_status", "SHARED"),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.created_by", "admin"),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.iscsi_features.0.enabled_authentications", "CHAP"),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.storage_features.0.flash_mode.0.is_enabled", "true"),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.usage_type", "USER"),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.0.is_hidden", "false"),
				),
			},
		},
	})
}

func TestAccNutanixVolumeGroupsV4DataSource_WithLimit(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("terraform-test-volume-group-%d", r)
	desc := "terraform test volume group description"
	limit := 3
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVolumeGroupsDataSourceWithLimit(name, desc, limit),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceAttrListNotEmpty(dataSourceVolumeGroups, "volumes", "name"),
					resource.TestCheckResourceAttr(dataSourceVolumeGroups, "volumes.#", strconv.Itoa(limit)),
				),
			},
		},
	})
}

func testAccVolumeGroupsDataSourceConfig(filepath, name, desc string) string {
	return testAccVolumeGroupResourceConfig(filepath, name, desc) + `
		data "nutanix_volume_groups_v2" "test" {
			depends_on = [resource.nutanix_volume_group_v2.test]
		}
	`
}

func testAccVolumeGroupsDataSourceWithFilter(filepath, name, desc string) string {
	return testAccVolumeGroupResourceConfig(filepath, name, desc) + fmt.Sprintf(`		
	data "nutanix_volume_groups_v2" "test" {
		filter = "name eq '%s'"
		depends_on = [resource.nutanix_volume_group_v2.test]
	}
	`, name)
}

func testAccVolumeGroupsDataSourceWithLimit(name, desc string, limit int) string {
	return fmt.Sprintf(
		`
			data "nutanix_clusters" "clusters" {}

			locals {
				cluster1 = [
					for cluster in data.nutanix_clusters.clusters.entities :
					cluster.metadata.uuid if cluster.service_list[0] != "PRISM_CENTRAL"
				][0]				
			}
	
			resource "nutanix_volume_group_v2" "test1" {
				name              = "%[1]s_1"
				cluster_reference = local.cluster1
				description       = "%[2]s"
			}
	
			resource "nutanix_volume_group_v2" "test2" {
				name              = "%[1]s_2"
				cluster_reference = local.cluster1
				description       = "%[2]s"
				depends_on        = [resource.nutanix_volume_group_v2.test1]
			}
	
			resource "nutanix_volume_group_v2" "test3" {
				name              = "%[1]s_3"
				cluster_reference = local.cluster1
				description       = "%[2]s"
				depends_on        = [resource.nutanix_volume_group_v2.test2]
			}
	
			data "nutanix_volume_groups_v2" "test" {
				filter     = "startswith(name, '%[1]s')"
				limit      = %[3]d
				depends_on = [resource.nutanix_volume_group_v2.test3]
			}
		`, name, desc, limit)
}
