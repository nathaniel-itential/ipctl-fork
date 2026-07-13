// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

// Service defines the base interface for all service types.
// It provides a minimal contract that all services must satisfy.
type Service interface {
	Export(string) (any, error)
}

// AccountServicer defines operations for managing Itential Platform user accounts.
// It provides methods for retrieving, activating, and deactivating accounts.
type AccountServicer interface {
	GetAll() ([]Account, error)
	Get(id string) (*Account, error)
	Activate(id string) error
	Deactivate(id string) error
}

// AdapterServicer defines operations for managing adapter instances.
// It handles CRUD operations, lifecycle management, and import/export functionality.
type AdapterServicer interface {
	GetAll() ([]Adapter, error)
	Get(name string) (*Adapter, error)
	Create(in Adapter) (*Adapter, error)
	Update(in Adapter) (*Adapter, error)
	Delete(name string) error
	Start(name string) error
	Stop(name string) error
	Restart(name string) error
	Import(in Adapter) (*Adapter, error)
	Export(name string) (*Adapter, error)
}

// AdapterModelServicer defines operations for managing adapter models.
// It provides methods for retrieving available adapter models and configurations.
type AdapterModelServicer interface {
	GetAll() ([]string, error)
}

// AnalyticTemplateServicer defines operations for managing analytic templates.
// It handles CRUD operations for templates used in analytics and reporting.
type AnalyticTemplateServicer interface {
	GetAll() ([]AnalyticTemplate, error)
	Get(id string) (*AnalyticTemplate, error)
	Create(in AnalyticTemplate) (*AnalyticTemplate, error)
	Delete(id string) error
	Import(in AnalyticTemplate) (*AnalyticTemplate, error)
	Export(id string) (*AnalyticTemplate, error)
}

// ApplicationServicer defines operations for managing applications.
// It provides methods for application lifecycle management and configuration.
type ApplicationServicer interface {
	GetAll() ([]Application, error)
	Get(id string) (*Application, error)
	Create(in Application) (*Application, error)
	Delete(id string) error
}

// AutomationServicer defines operations for managing automations.
// It handles CRUD operations and import/export for automation assets.
type AutomationServicer interface {
	GetAll() ([]*Automation, error)
	Get(id string) (*Automation, error)
	Create(in Automation) (*Automation, error)
	Delete(id string) error
	ImportTransformed(automations []any) (*Automation, error)
	Export(id string) (*Automation, error)
}

// CommandTemplateServicer defines operations for managing command templates.
// It provides methods for creating and managing reusable command templates.
type CommandTemplateServicer interface {
	GetAll() ([]CommandTemplate, error)
	Get(id string) (*CommandTemplate, error)
	Create(in CommandTemplate) (*CommandTemplate, error)
	Delete(id string) error
	Import(in CommandTemplate) (*CommandTemplate, error)
	Export(id string) (*CommandTemplate, error)
}

// ConfigurationParserServicer defines operations for managing configuration parsers.
// It handles CRUD operations for parsers used in configuration management.
type ConfigurationParserServicer interface {
	GetAll() ([]ConfigurationParser, error)
	Get(id string) (*ConfigurationParser, error)
	Create(in ConfigurationParser) (*ConfigurationParser, error)
	Delete(id string) error
}

// ConfigurationTemplateServicer defines operations for managing configuration templates.
// It provides methods for template CRUD operations and import/export.
type ConfigurationTemplateServicer interface {
	GetAll() ([]ConfigurationTemplate, error)
	Get(id string) (*ConfigurationTemplate, error)
	Create(in ConfigurationTemplate) (*ConfigurationTemplate, error)
	Update(id string, in ConfigurationTemplate) (*ConfigurationTemplate, error)
	Delete(id string) error
	Import(in ConfigurationTemplate) (*ConfigurationTemplate, error)
	Export(id string) (*ConfigurationTemplate, error)
}

// DeviceGroupServicer defines operations for managing device groups.
// It handles grouping and organization of network devices.
type DeviceGroupServicer interface {
	GetAll() ([]DeviceGroup, error)
	Get(id string) (*DeviceGroup, error)
	Create(in DeviceGroup) (*DeviceGroup, error)
	Delete(id string) error
}

// DeviceServicer defines operations for managing network devices.
// It provides CRUD operations for device inventory management.
type DeviceServicer interface {
	GetAll() ([]Device, error)
	Get(id string) (*Device, error)
	Create(in Device) (*Device, error)
	Delete(id string) error
}

// GoldenConfigServicer defines operations for managing golden configurations.
// It handles configuration templates and standards for network devices.
type GoldenConfigServicer interface {
	GetTrees() ([]GoldenConfigTree, error)
	GetTree(name string) (*GoldenConfigTree, error)
	Create(in GoldenConfigTree) (*GoldenConfigTree, error)
	Delete(id string) error
}

// GroupServicer defines operations for managing authorization groups.
// It provides methods for group CRUD operations and membership management.
type GroupServicer interface {
	GetAll() ([]Group, error)
	Get(id string) (*Group, error)
	Create(in Group) (*Group, error)
	Delete(id string) error
}

// HealthServicer defines operations for checking platform health and status.
// It provides methods for monitoring system health and component status.
type HealthServicer interface {
	GetHealth() (*HealthStatus, error)
	CheckHealth() error
}

// InstanceServicer defines operations for managing lifecycle manager instances.
// It handles CRUD operations for resource instances and their states.
type InstanceServicer interface {
	GetAll(modelId string) ([]Instance, error)
}

// IntegrationServicer defines operations for managing integrations.
// It handles integration configuration and lifecycle management.
type IntegrationServicer interface {
	GetAll() ([]Integration, error)
	Get(name string) (*Integration, error)
	Create(in Integration) (*Integration, error)
	Update(name string, in Integration) (*Integration, error)
	Delete(name string) error
	Start(name string) error
	Stop(name string) error
	Restart(name string) error
	Import(integrations []Integration) error
	Export(name string) (*Integration, error)
}

// IntegrationModelServicer defines operations for managing integration models.
// It provides methods for retrieving available integration models and schemas.
type IntegrationModelServicer interface {
	GetAll() ([]IntegrationModel, error)
	Get(name string) (*IntegrationModel, error)
}

// JsonFormServicer defines operations for managing JSON Form assets.
// It handles CRUD operations for dynamic form definitions.
type JsonFormServicer interface {
	GetAll() ([]JsonForm, error)
	Get(id string) (*JsonForm, error)
	Create(in JsonForm) (*JsonForm, error)
	Delete(ids []string) error
	Import(in JsonForm) (*JsonForm, error)
}

// ModelServicer defines operations for managing lifecycle manager models.
// It provides CRUD operations and action execution for resource models.
type ModelServicer interface {
	GetAll() ([]Model, error)
	Get(id string) (*Model, error)
	Create(in Model) (*Model, error)
	Delete(id string, deleteInstances bool) error
	Import(in Model) (*Model, error)
	Export(id string) (*Model, error)
	RunAction(modelId string, req RunActionRequest) (*RunActionResponse, error)
}

// PrebuiltServicer defines operations for managing prebuilt assets.
// It handles installation and management of pre-packaged automation content.
type PrebuiltServicer interface {
	GetAll() ([]Prebuilt, error)
	Get(id string) (*Prebuilt, error)
	Install(id string) error
	Uninstall(id string) error
}

// ProfileServicer defines operations for managing user profiles.
// It provides methods for profile configuration and preferences.
type ProfileServicer interface {
	GetAll() ([]Profile, error)
	Get(id string) (*Profile, error)
	Create(in Profile) (*Profile, error)
	Update(id string, in Profile) (*Profile, error)
	Delete(id string) error
}

// ProjectServicer defines operations for managing automation studio projects.
// It handles CRUD operations, import/export, and member management.
type ProjectServicer interface {
	GetAll() ([]Project, error)
	Get(id string) (*Project, error)
	Create(name string) (*Project, error)
	Delete(id string) error
	Import(data map[string]interface{}) (*Project, error)
	Export(id string) (*Project, error)
	UpdateProject(projectId string, data map[string]interface{}) error
}

// RoleServicer defines operations for managing authorization roles.
// It provides methods for role CRUD operations and permission management.
type RoleServicer interface {
	GetAll() ([]Role, error)
	Get(id string) (*Role, error)
	Create(in Role) (*Role, error)
	Delete(id string) error
}

// TagServicer defines operations for managing tags and labels.
// It handles tag CRUD operations and asset tagging.
type TagServicer interface {
	GetAll() ([]Tag, error)
	Get(id string) (*Tag, error)
	Create(in Tag) (*Tag, error)
	Delete(id string) error
}

// TemplateServicer defines operations for managing automation studio templates.
// It handles CRUD operations and import/export for template assets.
type TemplateServicer interface {
	GetAll() ([]Template, error)
	Get(id string) (*Template, error)
	Create(in Template) (*Template, error)
	Delete(id string) error
	Import(in Template) (*Template, error)
	Export(id string) (*Template, error)
}

// TransformationServicer defines operations for managing data transformations.
// It provides CRUD operations for transformation definitions and logic.
type TransformationServicer interface {
	GetAll() ([]Transformation, error)
	Get(id string) (*Transformation, error)
	Create(in Transformation) (*Transformation, error)
	Delete(id string) error
	Import(in Transformation) (*Transformation, error)
}

// TriggerServicer defines operations for managing automation triggers.
// It handles trigger configuration and event-based automation activation.
type TriggerServicer interface {
	GetAll() ([]Trigger, error)
	Get(id string) (*Trigger, error)
	Create(in Trigger) (*Trigger, error)
	Update(id string, in Trigger) (*Trigger, error)
	Delete(id string) error
}

// AgentProjectServicer defines operations for managing agent projects.
// It handles retrieval, CRUD, and import/export of agent project bundles.
type AgentProjectServicer interface {
	GetAll() ([]AgentProject, error)
	Get(id string) (*AgentProject, error)
	GetByName(name string) (*AgentProject, error)
	Create(name string, description string) (*AgentProject, error)
	Export(id string) (*AgentProjectBundle, error)
	Import(bundle AgentProjectBundle, conflictMode string) (*AgentProjectBundle, error)
	UpdateProject(projectId string, data map[string]interface{}) error
	Delete(id string) error
}

// WorkflowServicer defines operations for managing workflow assets.
// It handles CRUD operations, import/export, and workflow execution.
type WorkflowServicer interface {
	GetAll() ([]Workflow, error)
	Get(name string) (*Workflow, error)
	GetById(id string) (*Workflow, error)
	Create(in Workflow) (*Workflow, error)
	Update(in Workflow) (*Workflow, error)
	Delete(name string) error
	Import(in Workflow) (*Workflow, error)
	Export(name string) (*Workflow, error)
	ExportById(id string) (*Workflow, error)
}
