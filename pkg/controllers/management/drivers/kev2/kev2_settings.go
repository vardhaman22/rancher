package kev2

import (
	"context"
	"reflect"

	v32 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/controllers/management/drivers/utils"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"k8s.io/apimachinery/pkg/api/errors"

	mgmtcontrollers "github.com/rancher/rancher/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/types/config"
	"github.com/rancher/rancher/pkg/wrangler"
)

const (
	AlibabaKEv2Operator = "alibaba"
)

type CredentialFields map[string]v32.Field

var CredentialsData = map[string]CredentialFields{
	AlibabaKEv2Operator: {
		"accessKeyId": v32.Field{
			Create: true,
			Update: true,
			Type:   "string",
		},
		"accessKeySecret": v32.Field{
			Create: true,
			Update: true,
			Type:   "password",
		},
	},
}

type KEv2SettingsController struct {
	Settings        mgmtcontrollers.SettingController
	schemaLister    v3.DynamicSchemaLister
	schemaClient    v3.DynamicSchemaInterface
	wranglerContext *wrangler.Context
	ctx             context.Context
}

func Register(ctx context.Context, management *config.ManagementContext, wContext *wrangler.Context) {
	alibabaCtrlr := &KEv2SettingsController{
		Settings:     wContext.Mgmt.Setting(),
		ctx:          ctx,
		schemaClient: management.Management.DynamicSchemas(""),
		schemaLister: management.Management.DynamicSchemas("").Controller().Lister(),
	}

	wContext.Mgmt.Setting().OnChange(ctx, "kev2-settings-handler", alibabaCtrlr.sync)
}

func (k *KEv2SettingsController) sync(_ string, setting *v32.Setting) (*v32.Setting, error) {
	if setting == nil {
		return nil, nil
	}

	switch setting.Name {
	// AlibabaKEv2Operator
	case settings.EnableACK.Name:
		if setting.Value == "true" {
			err := k.createCredSchema(AlibabaKEv2Operator, CredentialsData[AlibabaKEv2Operator])
			if err != nil {
				return nil, err
			}
		}
	default:
		return setting, nil
	}

	return setting, nil
}

func (m *KEv2SettingsController) createCredSchema(operatorName string, credFields map[string]v32.Field) error {
	name := utils.CredentialConfigSchemaName(operatorName)
	credSchema, err := m.schemaLister.Get("", name)
	if err != nil {
		if errors.IsNotFound(err) {
			credentialSchema := &v32.DynamicSchema{
				Spec: v32.DynamicSchemaSpec{
					ResourceFields: credFields,
				},
			}
			credentialSchema.Name = name
			_, err := m.schemaClient.Create(credentialSchema)
			return err
		}
		return err
	} else if !reflect.DeepEqual(credSchema.Spec.ResourceFields, credFields) {
		toUpdate := credSchema.DeepCopy()
		toUpdate.Spec.ResourceFields = credFields
		_, err := m.schemaClient.Update(toUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}
