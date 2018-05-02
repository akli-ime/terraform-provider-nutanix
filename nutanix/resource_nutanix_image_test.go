package nutanix

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccNutanixImage_basic(t *testing.T) {
	r := rand.Int31()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNutanixImageDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNutanixImageConfig(r),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNutanixImageExists("nutanix_image.test"),
				),
			},
		},
	})
}

func testAccCheckNutanixImageExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		return nil
	}
}

func testAccCheckNutanixImageDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*NutanixClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "nutanix_image" {
			continue
		}
		for {
			_, err := conn.API.V3.GetImage(rs.Primary.ID)
			if err != nil {
				if strings.Contains(fmt.Sprint(err), "ENTITY_NOT_FOUND") {
					return nil
				}
				return err
			}
			time.Sleep(3000 * time.Millisecond)
		}

	}

	return nil
}

func testAccNutanixImageConfig(r int32) string {
	return fmt.Sprintf(`
provider "nutanix" {
  username = "admin"
  password = "Nutanix/1234"
  endpoint = "10.5.81.134"
  insecure = true
  port = 9440
}

resource "nutanix_image" "test" {
	metadata = {
		kind = "image"
	}

	name = "dou_image_%d"
	description = "Dou Image Test"
	name = "CentOS7-ISO"
	source_uri = "http://endor.dyn.nutanix.com/acro_images/DISKs/cirros-0.3.4-x86_64-disk.img"

	checksum = {
		checksum_algorithm = "SHA_256"
	}
}
`, r)
}
