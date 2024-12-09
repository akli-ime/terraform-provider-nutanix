package dataprotectionv2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	acc "github.com/terraform-providers/terraform-provider-nutanix/nutanix/acctest"
)

const resourceNameRecoveryPointReplicate = "nutanix_recovery_point_replicate_v2.test"

func TestAccNutanixRecoveryPointReplicateV2Resource_basic(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("tf-test-recovery-point-%d", r)
	//clsName := fmt.Sprintf("tf-test-cluster-rp-%d", r)
	vmName := fmt.Sprintf("tf-test-vm-rp-%d", r)
	// End time is two week later
	expirationTime := time.Now().Add(14 * 24 * time.Hour)

	expirationTimeFormatted := expirationTime.UTC().Format(time.RFC3339)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccFoundationPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testVmConfig(vmName) +
					testRecoveryPointReplicateResourceConfig(name, expirationTimeFormatted),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceNameRecoveryPointReplicate, "ext_id"),
					resource.TestCheckResourceAttr(resourceNameRecoveryPointReplicate, "pc_ext_id", testVars.DataProtection.PcExtID),
					resource.TestCheckResourceAttr(resourceNameRecoveryPointReplicate, "cluster_ext_id", testVars.DataProtection.ClusterExtID),
					resource.TestCheckResourceAttrSet(resourceNameRecoveryPointReplicate, "replicated_rp_ext_id"),
				),
			},
		},
	})
}

func testRecoveryPointReplicateResourceConfig(name, expirationTime string) string {
	return testRecoveryPointsResourceConfigWithVmRecoveryPoints(name, expirationTime) + `
	resource "nutanix_recovery_point_replicate_v2" "test" {
	  ext_id         = nutanix_recovery_points_v2.test.id
	  cluster_ext_id = local.data_protection.cluster_ext_id
	  pc_ext_id      = local.data_protection.pc_ext_id
	  depends_on     = [nutanix_recovery_points_v2.test]
	}`
}
