package iamv2_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	acc "github.com/terraform-providers/terraform-provider-nutanix/nutanix/acctest"
)

const datasourceNameUsers = "data.nutanix_users_v2.test"

func TestAccNutanixUsersV4Datasource_Basic(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("test-user-%d", r)
	path, _ := os.Getwd()
	filepath := path + "/../../../../test_config_v2.json"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testUsersDatasourceV4Config(filepath, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameUsers, "users.#"),
					resource.TestCheckResourceAttrSet(datasourceNameUsers, "users.0.username"),
					resource.TestCheckResourceAttrSet(datasourceNameUsers, "users.0.user_type"),
					resource.TestCheckResourceAttrSet(datasourceNameUsers, "users.0.ext_id"),
				),
			},
		},
	})
}

func TestAccNutanixUsersV4Datasource_WithFilter(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("test-user-%d", r)
	path, _ := os.Getwd()
	filepath := path + "/../../../../test_config_v2.json"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testUsersDatasourceV4WithFilterConfig(filepath, name, "userType eq Schema.Enums.UserType'LOCAL'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameUsers, "users.0.ext_id"),
					resource.TestCheckResourceAttr(datasourceNameUsers, "users.0.user_type", "LOCAL"),
				),
			},
			{
				Config: testUsersDatasourceV4WithFilterConfig(filepath, name, "username eq '"+name+"'"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(datasourceNameUsers, "users.0.ext_id"),
					resource.TestCheckResourceAttr(datasourceNameUsers, "users.0.username", name),
				),
			},
		},
	})
}

func TestAccNutanixUsersV4Datasource_WithLimit(t *testing.T) {
	r := acctest.RandInt()
	name := fmt.Sprintf("test-user-%d", r)
	limit := 1
	path, _ := os.Getwd()
	filepath := path + "/../../../../test_config_v2.json"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { acc.TestAccPreCheck(t) },
		Providers: acc.TestAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testUsersDatasourceV4WithLimitConfig(filepath, name, limit),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceNameUsers, "users.#", strconv.Itoa(limit)),
				),
			},
		},
	})
}

func testUsersDatasourceV4Config(filepath, name string) string {
	return fmt.Sprintf(`
		locals{
			config = (jsondecode(file("%[1]s")))
			users = local.config.iam.users
		}
		
		resource "nutanix_users_v2" "test" {
			username = "%[2]s"
			first_name = "first-name-%[2]s"
			middle_initial = "middle-initial-%[2]s"
			last_name = "last-name-%[2]s"
			email_id = local.users.email_id
			locale = local.users.locale
			region = local.users.region
			display_name = "display-name-%[2]s"
			password = local.users.password
			user_type = "LOCAL"
			status = "ACTIVE"  
			force_reset_password = local.users.force_reset_password   
		}

		data "nutanix_users_v2" "test"{
			depends_on = [nutanix_users_v2.test]
		}
	`, filepath, name)
}

func testUsersDatasourceV4WithFilterConfig(filepath, name, userQuery string) string {
	return fmt.Sprintf(`

	locals{
		config = (jsondecode(file("%[1]s")))
		users = local.config.iam.users
	}

	resource "nutanix_users_v2" "test" {
		username = "%[2]s"
		first_name = "first-name-%[2]s"
		middle_initial = "middle-initial-%[2]s"
		last_name = "last-name-%[2]s"
		email_id = local.users.email_id
		locale = local.users.locale
		region = local.users.region
		display_name = "display-name-%[2]s"
		password = local.users.password
		user_type = "LOCAL"
		status = "ACTIVE"  
		force_reset_password = local.users.force_reset_password   
	}
	
	data "nutanix_users_v2" "test" {
		filter = "%[3]s"
		depends_on = [nutanix_users_v2.test]
	}

	
	`, filepath, name, userQuery)
}

func testUsersDatasourceV4WithLimitConfig(filepath, name string, limit int) string {
	return fmt.Sprintf(`
		locals{
			config = (jsondecode(file("%[1]s")))
			users = local.config.iam.users
		}
		
		resource "nutanix_users_v2" "test" {
			username = "%[2]s"
			first_name = "first-name-%[2]s"
			middle_initial = "middle-initial-%[2]s"
			last_name = "last-name-%[2]s"
			email_id = local.users.email_id
			locale = local.users.locale
			region = local.users.region
			display_name = "display-name-%[2]s"
			password = local.users.password
			user_type = "LOCAL"
			status = "ACTIVE"  
			force_reset_password = local.users.force_reset_password   
		}
		
		data "nutanix_users_v2" "test" {
			limit     = %[3]d
			depends_on = [nutanix_users_v2.test]
		}
	`, filepath, name, limit)
}
