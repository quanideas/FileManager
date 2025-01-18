package constants

const (
	UserLogin = "/account/login"

	// Company
	CompanyCreate  = "/create"
	CompanyGetByID = "/get"
	CompanyUpdate  = "/update"
	CompanyGetAll  = "/get-all"

	// User
	UserCreate                = "/create"
	UserAddPermission         = "/add-permission"
	UserRemovePermission      = "/remove-permission"
	UserGetByID               = "/get"
	UserGetByCompany          = "/getlist"
	UserGetSpecificPermission = "/get-permission"
	UserUpdate                = "/update"

	// Role
	RoleCreate           = "/create"
	RoleUpdate           = "/update"
	RoleGetAll           = "/get-all"
	RoleGetByID          = "/get"
	RoleDelete           = "/delete"
	RoleAddUser          = "/add-user"
	RoleRemoveUser       = "/remove-user"
	RoleAddPermission    = "/add-permission"
	RoleRemovePermission = "/remove-permission"

	// Project
	ProjectGetByID                 = "/get"
	ProjectGetCompanyIDByProjectID = "/get-company-id"
	ProjectCreate                  = "/create"
	ProjectGetAll                  = "/get-all"
	ProjectIterationGet            = "/get-iteration"
	ProjectIterationCreate         = "/create-iteration"
	ProjectIterationUpdate         = "/update-iteration"
	ProjectIterationDelete         = "/delete-iteration"

	// Permission
	PermissionGetAll   = "/get-all"
	PermissionValidate = "/validate-permission"

	// Dashboard
	DashboardStats = "/stats"

	// Token
	TokenValidation = "/validate-token"
)
