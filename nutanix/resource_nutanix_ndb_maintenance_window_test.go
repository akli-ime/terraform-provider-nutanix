package nutanix

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const resourceMaintenaceWindowName = "nutanix_ndb_maintenance_window.acctest-managed"

func TestAccEra_MaintenanceWindow(t *testing.T) {
	name := "test-maintenance"
	desc := "this is desc"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccEraPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEraMaintenanceWindow(name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "name", name),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "description", desc),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "recurrence", "WEEKLY"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "duration", "2"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "day_of_week", "TUESDAY"),
				),
			},
		},
	})
}

func TestAccEra_MaintenanceWindowUpdate(t *testing.T) {
	name := "test-maintenance"
	updatedName := "test-maintenance-updated"
	desc := "this is desc"
	updatedDesc := "this desc is updated"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccEraPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEraMaintenanceWindow(name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "name", name),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "description", desc),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "recurrence", "WEEKLY"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "duration", "2"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "day_of_week", "TUESDAY"),
				),
			},
			{
				Config: testAccEraMaintenanceWindowUpdate(updatedName, updatedDesc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "description", updatedDesc),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "recurrence", "WEEKLY"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "duration", "4"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "day_of_week", "MONDAY"),
				),
			},
		},
	})
}

func TestAccEra_MaintenanceWindow_MonthlyRecurrence(t *testing.T) {
	name := "test-maintenance"
	desc := "this is desc"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccEraPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccEraMaintenanceWindowByMonthlyRecurrence(name, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "name", name),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "description", desc),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "recurrence", "MONTHLY"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "duration", "2"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "day_of_week", "TUESDAY"),
					resource.TestCheckResourceAttr(resourceMaintenaceWindowName, "week_of_month", "4"),
				),
			},
		},
	})
}

func testAccEraMaintenanceWindow(name, desc string) string {
	return fmt.Sprintf(`
		resource nutanix_ndb_maintenance_window acctest-managed{
			name = "%[1]s"
			description = "%[2]s"
			recurrence = "WEEKLY"
			duration = 2
			day_of_week = "TUESDAY"
			start_time = "17:04:47" 
		}
	`, name, desc)
}

func testAccEraMaintenanceWindowUpdate(name, desc string) string {
	return fmt.Sprintf(`
		resource nutanix_ndb_maintenance_window acctest-managed{
			name = "%[1]s"
			description = "%[2]s"
			recurrence = "WEEKLY"
			duration = 4
			day_of_week = "MONDAY"
			start_time = "17:04:47" 
		}
	`, name, desc)
}

func testAccEraMaintenanceWindowByMonthlyRecurrence(name, desc string) string {
	return fmt.Sprintf(`
		resource nutanix_ndb_maintenance_window acctest-managed{
			name = "%[1]s"
			description = "%[2]s"
			recurrence = "MONTHLY"
			duration = 2
			day_of_week = "TUESDAY"
			start_time = "17:04:47" 
			week_of_month= 4
		}
	`, name, desc)
}
