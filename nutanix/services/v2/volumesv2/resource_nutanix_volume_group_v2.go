package volumesv2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	taskPoll "github.com/nutanix-core/ntnx-api-golang-sdk-internal/prism-go-client/v16/models/prism/v4/config"
	volumesPrism "github.com/nutanix-core/ntnx-api-golang-sdk-internal/volumes-go-client/v16/models/prism/v4/config"
	volumesClient "github.com/nutanix-core/ntnx-api-golang-sdk-internal/volumes-go-client/v16/models/volumes/v4/config"

	conns "github.com/terraform-providers/terraform-provider-nutanix/nutanix"
	"github.com/terraform-providers/terraform-provider-nutanix/nutanix/sdks/v4/prism"
	"github.com/terraform-providers/terraform-provider-nutanix/utils"
)

// CRUD for Volume Group.
func ResourceNutanixVolumeGroupV2() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates a new Volume Group.",
		CreateContext: ResourceNutanixVolumeGroupV2Create,
		ReadContext:   ResourceNutanixVolumeGroupV2Read,
		UpdateContext: ResourceNutanixVolumeGroupV2Update,
		DeleteContext: ResourceNutanixVolumeGroupV2Delete,

		Schema: map[string]*schema.Schema{
			"ext_id": {
				Description: "A globally unique identifier of an instance that is suitable for external consumption.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Volume Group name. This is an Required field.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Volume Group description. This is an optional field.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"should_load_balance_vm_attachments": {
				Description: "Indicates whether to enable Volume Group load balancing for VM attachments. This cannot be enabled if there are iSCSI client attachments already associated with the Volume Group, and vice-versa. This is an optional field.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"sharing_status": {
				Description:  "Indicates whether the Volume Group can be shared across multiple iSCSI initiators. The mode cannot be changed from SHARED to NOT_SHARED on a Volume Group with multiple attachments. Similarly, a Volume Group cannot be associated with more than one attachment as long as it is in exclusive mode. This is an optional field.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"NOT_SHARED", "SHARED"}, false),
			},
			"target_prefix": {
				Description: "The specifications contain the target prefix for external clients as the value. This is an optional field.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"target_name": {
				Description: "Name of the external client target that will be visible and accessible to the client. This is an optional field.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"enabled_authentications": {
				Description:  "The authentication type enabled for the Volume Group. This is an optional field. If omitted, authentication is not configured for the Volume Group. If this is set to CHAP, the target/client secret must be provided.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"CHAP", "NONE"}, false),
			},
			"iscsi_features": {
				Description: "iSCSI specific settings for the Volume Group. This is an optional field.",
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_secret": {
							Description: "Target secret in case of a CHAP authentication. This field must only be provided in case the authentication type is not set to CHAP. This is an optional field and it cannot be retrieved once configured.",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"enabled_authentications": {
							Description:  "The authentication type enabled for the Volume Group. This is an optional field. If omitted, authentication is not configured for the Volume Group. If this is set to CHAP, the target/client secret must be provided.",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"CHAP", "NONE"}, false),
						},
					},
				},
			},
			"created_by": {
				Description: "Service/user who created this Volume Group. This is an optional field.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"cluster_reference": {
				Description: "The UUID of the cluster that will host the Volume Group. This is a mandatory field for creating a Volume Group on Prism Central.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"storage_features": {
				Description: "Storage optimization features which must be enabled on the Volume Group. This is an optional field.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flash_mode": {
							Description: "Once configured, this field will avoid down migration of data from the hot tier unless the overrides field is specified for the virtual disks.",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"is_enabled": {
										Description: "Indicates whether the flash mode is enabled for the Volume Group.",
										Type:        schema.TypeBool,
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"usage_type": {
				Description:  "Expected usage type for the Volume Group. This is an indicative hint on how the caller will consume the Volume Group. This is an optional field.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"USER", "INTERNAL", "TEMPORARY", "BACKUP_TARGET"}, false),
			},
			"is_hidden": {
				Description: "Indicates whether the Volume Group is meant to be hidden or not. This is an optional field. If omitted, the VG will not be hidden.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func ResourceNutanixVolumeGroupV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO_VG] Creating Volume Group")
	conn := meta.(*conns.Client).VolumeAPI

	body := volumesClient.VolumeGroup{}

	// Required field
	if name, nok := d.GetOk("name"); nok {
		body.Name = utils.StringPtr(name.(string))
	}
	if desc, ok := d.GetOk("description"); ok {
		body.Description = utils.StringPtr(desc.(string))
	}
	if shouldLoadBalanceVmAttachments, ok := d.GetOk("should_load_balance_vm_attachments"); ok {
		body.ShouldLoadBalanceVmAttachments = utils.BoolPtr(shouldLoadBalanceVmAttachments.(bool))
	}
	if sharingStatus, ok := d.GetOk("sharing_status"); ok {
		sharingStatusMap := map[string]interface{}{
			"SHARED":     2,
			"NOT_SHARED": 3,
		}
		pVal := sharingStatusMap[sharingStatus.(string)]
		p := volumesClient.SharingStatus(pVal.(int))
		body.SharingStatus = &p
	}
	if targetPrefix, ok := d.GetOk("target_prefix"); ok {
		body.TargetPrefix = utils.StringPtr(targetPrefix.(string))
	}
	if targetName, ok := d.GetOk("target_name"); ok {
		body.TargetName = utils.StringPtr(targetName.(string))
	}
	// if enabledAuthentications, ok := d.GetOk("enabled_authentications"); ok {
	// 	enabledAuthenticationsMap := map[string]interface{}{
	// 		"CHAP": 2,
	// 		"NONE": 3,
	// 	}
	// 	pVal := enabledAuthenticationsMap[enabledAuthentications.(string)]
	// 	p := volumesClient.AuthenticationType(pVal.(int))
	// 	body.EnabledAuthentications = &p
	// } else {
	// 	p := volumesClient.AuthenticationType(0) // Replace 0 with the appropriate default value
	// 	body.EnabledAuthentications = &p
	// }
	if iscsiFeatures, ok := d.GetOk("iscsi_features"); ok {
		body.IscsiFeatures = expandIscsiFeatures(iscsiFeatures.([]interface{}))
	}
	if createdBy, ok := d.GetOk("created_by"); ok {
		body.CreatedBy = utils.StringPtr(createdBy.(string))
	}
	// Required field
	if clusterReference, ok := d.GetOk("cluster_reference"); ok {
		body.ClusterReference = utils.StringPtr(clusterReference.(string))
	}
	if storageFeatures, ok := d.GetOk("storage_features"); ok {
		body.StorageFeatures = expandStorageFeatures(storageFeatures.([]interface{}))
	}
	if usageType, ok := d.GetOk("usage_type"); ok {
		usageTypeMap := map[string]interface{}{
			"USER":          2,
			"INTERNAL":      3,
			"TEMPORARY":     4,
			"BACKUP_TARGET": 5,
		}
		pInt := usageTypeMap[usageType.(string)]
		p := volumesClient.UsageType(pInt.(int))
		body.UsageType = &p
	}
	if isHidden, ok := d.GetOk("is_hidden"); ok {
		body.IsHidden = utils.BoolPtr(isHidden.(bool))
	}

	resp, err := conn.VolumeAPIInstance.CreateVolumeGroup(&body)
	if err != nil {
		var errordata map[string]interface{}
		e := json.Unmarshal([]byte(err.Error()), &errordata)
		if e != nil {
			return diag.FromErr(e)
		}
		log.Printf("[INFO_VG] Error Data: %v", errordata)
		data := errordata["data"].(map[string]interface{})
		errorList := data["error"].([]interface{})
		errorMessage := errorList[0].(map[string]interface{})
		return diag.Errorf("error while creating Volume Group : %v", errorMessage["message"])
	}

	TaskRef := resp.Data.GetValue().(volumesPrism.TaskReference)
	taskUUID := TaskRef.ExtId

	taskconn := meta.(*conns.Client).PrismAPI
	// Wait for the VM to be available
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PENDING", "RUNNING", "QUEUED"},
		Target:  []string{"SUCCEEDED"},
		Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
		return diag.Errorf("error waiting for template (%s) to create: %s", utils.StringValue(taskUUID), errWaitTask)
	}

	// Get UUID from TASK API

	resourceUUID, err := taskconn.TaskRefAPI.GetTaskById(taskUUID, nil)
	if err != nil {
		var errordata map[string]interface{}
		e := json.Unmarshal([]byte(err.Error()), &errordata)
		if e != nil {
			return diag.FromErr(e)
		}
		data := errordata["data"].(map[string]interface{})
		errorList := data["error"].([]interface{})
		errorMessage := errorList[0].(map[string]interface{})
		return diag.Errorf("error while fetching Volume Group UUID : %v", errorMessage["message"])
	}
	rUUID := resourceUUID.Data.GetValue().(taskPoll.Task)

	uuid := rUUID.EntitiesAffected[0].ExtId
	d.SetId(*uuid)
	d.Set("ext_id", *uuid)

	return nil
}

func ResourceNutanixVolumeGroupV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.Client).VolumeAPI

	resp, err := conn.VolumeAPIInstance.GetVolumeGroupById(utils.StringPtr(d.Id()))
	if err != nil {
		var errordata map[string]interface{}
		e := json.Unmarshal([]byte(err.Error()), &errordata)
		if e != nil {
			return diag.FromErr(e)
		}
		data := errordata["data"].(map[string]interface{})
		errorList := data["error"].([]interface{})
		errorMessage := errorList[0].(map[string]interface{})
		return diag.Errorf("error while fetching Volume Group : %v", errorMessage["message"])
	}

	getResp := resp.Data.GetValue().(volumesClient.VolumeGroup)

	if err := d.Set("name", getResp.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", getResp.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("should_load_balance_vm_attachments", getResp.ShouldLoadBalanceVmAttachments); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sharing_status", flattenSharingStatus(getResp.SharingStatus)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_prefix", getResp.TargetPrefix); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("target_name", getResp.TargetName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("enabled_authentications", flattenEnabledAuthentications(getResp.EnabledAuthentications)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("iscsi_features", flattenIscsiFeatures(getResp.IscsiFeatures)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_by", getResp.CreatedBy); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cluster_reference", getResp.ClusterReference); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("storage_features", flattenStorageFeatures(getResp.StorageFeatures)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("usage_type", flattenUsageType(getResp.UsageType)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("is_hidden", getResp.IsHidden); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceNutanixVolumeGroupV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func ResourceNutanixVolumeGroupV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.Client).VolumeAPI

	resp, err := conn.VolumeAPIInstance.DeleteVolumeGroupById(utils.StringPtr(d.Id()))
	if err != nil {
		var errordata map[string]interface{}
		e := json.Unmarshal([]byte(err.Error()), &errordata)
		if e != nil {
			return diag.FromErr(e)
		}
		data := errordata["data"].(map[string]interface{})
		errorList := data["error"].([]interface{})
		errorMessage := errorList[0].(map[string]interface{})
		return diag.Errorf("error while Deleting Volume group : %v", errorMessage["message"])
	}

	TaskRef := resp.Data.GetValue().(volumesPrism.TaskReference)
	taskUUID := TaskRef.ExtId

	// calling group API to poll for completion of task
	taskconn := meta.(*conns.Client).PrismAPI
	// Wait for the VM to be available
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PENDING", "RUNNING", "QUEUED"},
		Target:  []string{"SUCCEEDED"},
		Refresh: taskStateRefreshPrismTaskGroupFunc(ctx, taskconn, utils.StringValue(taskUUID)),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	if _, errWaitTask := stateConf.WaitForStateContext(ctx); errWaitTask != nil {
		return diag.Errorf("error waiting for template (%s) to create: %s", utils.StringValue(taskUUID), errWaitTask)
	}
	return nil
}

func expandIscsiFeatures(IscsiFeaturesList interface{}) *volumesClient.IscsiFeatures {
	if len(IscsiFeaturesList.([]interface{})) > 0 {
		iscsiFeature := &volumesClient.IscsiFeatures{}
		iscsiFeaturesI := IscsiFeaturesList.([]interface{})
		if iscsiFeaturesI[0] == nil {
			return nil
		}
		val := iscsiFeaturesI[0].(map[string]interface{})

		if targetSecret, ok := val["target_secret"]; ok {
			iscsiFeature.TargetSecret = utils.StringPtr(targetSecret.(string))
		}

		if enabledAuthentications, ok := val["enabled_authentications"]; ok {
			enabledAuthenticationsMap := map[string]interface{}{
				"CHAP": 2,
				"NONE": 3,
			}
			pVal := enabledAuthenticationsMap[enabledAuthentications.(string)]
			p := volumesClient.AuthenticationType(pVal.(int))
			iscsiFeature.EnabledAuthentications = &p
		}
		log.Printf("[INFO_VG] iscsiFeature.EnabledAuthentications: %v", *iscsiFeature.EnabledAuthentications)
		log.Printf("[INFO_VG] iscsiFeature.TargetSecret: %v", *iscsiFeature.TargetSecret)
		return iscsiFeature
	}
	return nil
}

func expandStorageFeatures(storageFeaturesList []interface{}) *volumesClient.StorageFeatures {
	if len(storageFeaturesList) > 0 {
		storageFeature := volumesClient.StorageFeatures{}

		val := storageFeaturesList[0].(map[string]interface{})

		if flashMode, ok := val["flash_mode"]; ok {
			storageFeature.FlashMode = expandFlashMode(flashMode.([]interface{}))
		}
		return &storageFeature
	}
	return nil
}

func expandFlashMode(flashModeList []interface{}) *volumesClient.FlashMode {
	if len(flashModeList) > 0 {
		flashMode := volumesClient.FlashMode{}

		val := flashModeList[0].(map[string]interface{})

		if isEnabled, ok := val["is_enabled"]; ok {
			flashMode.IsEnabled = utils.BoolPtr(isEnabled.(bool))
		}
		return &flashMode
	}
	return nil
}

func taskStateRefreshPrismTaskGroupFunc(ctx context.Context, client *prism.Client, taskUUID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		vresp, err := client.TaskRefAPI.GetTaskById(utils.StringPtr(taskUUID), nil)

		if err != nil {
			var errordata map[string]interface{}
			e := json.Unmarshal([]byte(err.Error()), &errordata)
			if e != nil {
				return nil, "", e
			}
			data := errordata["data"].(map[string]interface{})
			errorList := data["error"].([]interface{})
			errorMessage := errorList[0].(map[string]interface{})
			return "", "", (fmt.Errorf("error while polling prism task: %v", errorMessage["message"]))
		}

		// get the group results

		v := vresp.Data.GetValue().(taskPoll.Task)

		if getTaskStatus(v.Status) == "CANCELED" || getTaskStatus(v.Status) == "FAILED" {
			return v, getTaskStatus(v.Status),
				fmt.Errorf("error_detail: %s, progress_message: %d", utils.StringValue(v.ErrorMessages[0].Message), utils.IntValue(v.ProgressPercentage))
		}
		return v, getTaskStatus(v.Status), nil
	}
}

func getTaskStatus(taskStatus *taskPoll.TaskStatus) string {
	if taskStatus != nil {
		if *taskStatus == taskPoll.TaskStatus(6) {
			return "FAILED"
		}
		if *taskStatus == taskPoll.TaskStatus(7) {
			return "CANCELED"
		}
		if *taskStatus == taskPoll.TaskStatus(2) {
			return "QUEUED"
		}
		if *taskStatus == taskPoll.TaskStatus(3) {
			return "RUNNING"
		}
		if *taskStatus == taskPoll.TaskStatus(5) {
			return "SUCCEEDED"
		}
	}
	return "UNKNOWN"
}
