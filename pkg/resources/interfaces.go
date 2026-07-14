// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package resources

import "github.com/itential/ipctl/pkg/services"

// AccountResourcer defines operations for account business logic.
// It provides methods for retrieving accounts with business logic applied.
type AccountResourcer interface {
	GetAll() ([]services.Account, error)
	Get(id string) (*services.Account, error)
	GetByName(name string) (*services.Account, error)
	Activate(id string) error
	Deactivate(id string) error
}

// AdapterResourcer defines operations for adapter business logic.
// It handles adapter management with business rules and validations.
type AdapterResourcer interface {
	GetAll() ([]services.Adapter, error)
	Get(name string) (*services.Adapter, error)
	Create(in services.Adapter) (*services.Adapter, error)
	Update(in services.Adapter) (*services.Adapter, error)
	Delete(name string) error
	Start(name string) error
	Stop(name string) error
	Restart(name string) error
	Import(in services.Adapter) (*services.Adapter, error)
	Export(name string) (*services.Adapter, error)
}

// AutomationResourcer defines operations for automation business logic.
// It provides methods for managing automations with validation and transformation.
type AutomationResourcer interface {
	GetAll() ([]*services.Automation, error)
	Get(id string) (*services.Automation, error)
	GetByName(name string) (*services.Automation, error)
	Create(in services.Automation) (*services.Automation, error)
	Delete(id string) error
	Import(in services.Automation) (*services.Automation, error)
	ImportTransformed(automations []any) (*services.Automation, error)
	Export(id string) (*services.Automation, error)
	Clear() error
}

// ConfigurationTemplateResourcer defines operations for configuration template business logic.
// It handles template management with business rules applied.
type ConfigurationTemplateResourcer interface {
	GetByName(name string) (*services.ConfigurationTemplate, error)
}

// DeviceGroupResourcer defines operations for device group business logic.
// It provides methods for managing device groups with business rules.
type DeviceGroupResourcer interface {
	GetByName(name string) (*services.DeviceGroup, error)
}

// GroupResourcer defines operations for authorization group business logic.
// It handles group management with business rules and validations.
type GroupResourcer interface {
	GetAll() ([]services.Group, error)
	Get(id string) (*services.Group, error)
	GetByName(name string) (*services.Group, error)
	Create(in services.Group) (*services.Group, error)
	Delete(id string) error
}

// JsonFormResourcer defines operations for JSON Form business logic.
// It provides methods for managing JSON Forms with business rules applied.
type JsonFormResourcer interface {
	GetAll() ([]services.JsonForm, error)
	Get(id string) (*services.JsonForm, error)
	GetByName(name string) (*services.JsonForm, error)
	Create(in services.JsonForm) (*services.JsonForm, error)
	Delete(ids []string) error
	Import(in services.JsonForm) (*services.JsonForm, error)
	Clear() error
}

// ModelResourcer defines operations for lifecycle manager model business logic.
// It handles model management with business rules and action execution.
type ModelResourcer interface {
	GetAll() ([]services.Model, error)
	GetByName(name string) (*services.Model, error)
	Create(in services.Model) (*services.Model, error)
	Delete(id string, deleteInstances bool) error
	DeleteWithOptions(model *services.Model, opts DeleteOptions) error
	GetInstances(modelId string) ([]services.Instance, error)
	Export(id string) (*services.Model, error)
	Import(in services.Model) (*services.Model, error)
}

// ProjectResourcer defines operations for project business logic.
// It provides methods for managing projects with transformation and member management.
type ProjectResourcer interface {
	GetAll() ([]services.Project, error)
	Get(id string) (*services.Project, error)
	GetByName(name string) (*services.Project, error)
	Create(name string) (*services.Project, error)
	Delete(id string) error
	Import(in services.Project, cfg services.ProjectImportConfig) (*services.Project, error)
	ImportTransformed(data map[string]interface{}) (*services.Project, error)
	Export(id string) (*services.Project, error)
	UpdateMembers(projectId string, members []services.ProjectMember) error
	AddMembers(projectId string, members []services.ProjectMember) error
}

// TemplateResourcer defines operations for template business logic.
// It handles template management with business rules applied.
type TemplateResourcer interface {
	GetAll() ([]services.Template, error)
	Get(id string) (*services.Template, error)
	GetByName(name string) (*services.Template, error)
	Create(in services.Template) (*services.Template, error)
	Delete(id string) error
	Import(in services.Template) (*services.Template, error)
	Export(id string) (*services.Template, error)
}

// TransformationResourcer defines operations for transformation business logic.
// It provides methods for managing transformations with business rules.
type TransformationResourcer interface {
	GetAll() ([]services.Transformation, error)
	Get(name string) (*services.Transformation, error)
	GetByName(name string) (*services.Transformation, error)
	Create(in services.Transformation) (*services.Transformation, error)
	Delete(id string) error
	Import(in services.Transformation) (*services.Transformation, error)
	Clear() error
}

// WorkflowResourcer defines operations for workflow business logic.
// It handles workflow management with business rules and bulk operations.
type WorkflowResourcer interface {
	GetAll() ([]services.Workflow, error)
	Get(name string) (*services.Workflow, error)
	GetById(id string) (*services.Workflow, error)
	Create(in services.Workflow) (*services.Workflow, error)
	Update(in services.Workflow) (*services.Workflow, error)
	Delete(name string) error
	Import(in services.Workflow) (*services.Workflow, error)
	Export(name string) (*services.Workflow, error)
	ExportById(id string) (*services.Workflow, error)
	Clear() error
}
