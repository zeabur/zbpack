package types

import "fmt"

// FieldInfo is the information of a field in a struct.
type FieldInfo struct {
	// Key is the key of the field in the struct.
	Key string
	// Name is the name of the field in the struct.
	Name string
	// Description is the description of the field in the struct.
	Description string
	// Icon is the icon of the field in the struct.
	//
	// It should be a URL of the icon.
	Icon string
}

// NewPlanTypeFieldInfo creates a new generic FieldInfo for the PlanType (provider).
//
// It creates a generic Provider field information with the icon of the provider.
func NewPlanTypeFieldInfo(planType PlanType) FieldInfo {
	return FieldInfo{
		Key:         "_provider",
		Name:        "Provider",
		Description: "The programing language or runtime detected in the source code",
		Icon:        fmt.Sprintf("https://raw.githubusercontent.com/zeabur/service-icons/main/git/%s/default.svg", planType),
	}
}

// NewFrameworkFieldInfo creates a new generic FieldInfo for the framework.
//
// It creates a generic Framework field information with the icon of the framework.
func NewFrameworkFieldInfo(key string, planType PlanType, framework string) FieldInfo {
	return FieldInfo{
		Key:         key,
		Name:        "Framework",
		Description: "The framework detected in the source code",
		Icon:        fmt.Sprintf("https://raw.githubusercontent.com/zeabur/service-icons/main/git/%s/%s.svg", planType, framework),
	}
}

// NewOutputDirFieldInfo creates a new generic FieldInfo for the output directory.
func NewOutputDirFieldInfo(key string) FieldInfo {
	return FieldInfo{
		Key:         key,
		Name:        "Output Directory",
		Description: "The directory where Zeabur will use as the output static files.",
	}
}

// NewInstallCmdFieldInfo creates a new generic FieldInfo for the installation command.
func NewInstallCmdFieldInfo(key string) FieldInfo {
	return FieldInfo{
		Key:         key,
		Name:        "Install Command",
		Description: "The command to install the dependencies",
	}
}

// NewBuildCmdFieldInfo creates a new generic FieldInfo for the build command.
func NewBuildCmdFieldInfo(key string) FieldInfo {
	return FieldInfo{
		Key:         key,
		Name:        "Build Command",
		Description: "The command to build the source code",
	}
}

// NewStartCmdFieldInfo creates a new generic FieldInfo for the start command.
func NewStartCmdFieldInfo(key string) FieldInfo {
	return FieldInfo{
		Key:         key,
		Name:        "Start Command",
		Description: "The command to start the service",
	}
}

// NewServerlessFieldInfo creates a new generic FieldInfo for the serverless field.
func NewServerlessFieldInfo(key string) FieldInfo {
	return FieldInfo{
		Key:         key,
		Name:        "Serverless",
		Description: "Whether the project is serverless",
	}
}
