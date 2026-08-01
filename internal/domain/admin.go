package domain

import (
	"context"
	"time"
)

// User represents an administrative user / operator in the system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Password     string    `json:"password,omitempty"` // Used for request DTO
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RoleIDs      []int     `json:"role_ids,omitempty"`
	Roles        []*Role   `json:"roles,omitempty"`
}

// Role represents an RBAC role entity.
type Role struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	PermissionIDs []int         `json:"permission_ids,omitempty"`
	Permissions   []*Permission `json:"permissions,omitempty"`
}

// Permission represents an RBAC permission entity.
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CategoryWithZones represents a Category along with its assigned allowed zone IDs.
type CategoryWithZones struct {
	Category
	AllowedZoneIDs []int `json:"allowed_zone_ids"`
}

// MealScheduleWithCategories represents a MealSchedule along with its allowed category IDs.
type MealScheduleWithCategories struct {
	MealSchedule
	AllowedCategoryIDs []int `json:"allowed_category_ids"`
}

// Request DTOs
type CreateZoneRequest struct {
	Name          string `json:"name" binding:"required"`
	Code          string `json:"code" binding:"required"`
	RequiresInOut bool   `json:"requires_in_out"`
}

type UpdateZoneRequest struct {
	Name          string `json:"name"`
	Code          string `json:"code"`
	RequiresInOut *bool  `json:"requires_in_out"`
}

type CreateCategoryRequest struct {
	Name           string `json:"name" binding:"required"`
	ColorCode      string `json:"color_code" binding:"required"`
	CanEat         bool   `json:"can_eat"`
	AllowedZoneIDs []int  `json:"allowed_zone_ids"`
}

type UpdateCategoryRequest struct {
	Name           string `json:"name"`
	ColorCode      string `json:"color_code"`
	CanEat         *bool  `json:"can_eat"`
	AllowedZoneIDs []int  `json:"allowed_zone_ids"`
}

type CreateMealScheduleRequest struct {
	Date               string   `json:"date" binding:"required"`
	MealType           MealType `json:"meal_type" binding:"required"`
	StartTime          string   `json:"start_time" binding:"required"`
	EndTime            string   `json:"end_time" binding:"required"`
	AllowedCategoryIDs []int    `json:"allowed_category_ids"`
}

type UpdateMealScheduleRequest struct {
	Date               string   `json:"date"`
	MealType           MealType `json:"meal_type"`
	StartTime          string   `json:"start_time"`
	EndTime            string   `json:"end_time"`
	AllowedCategoryIDs []int    `json:"allowed_category_ids"`
}

type AssignRolePermissionsRequest struct {
	PermissionIDs []int `json:"permission_ids" binding:"required"`
}

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	RoleIDs  []int  `json:"role_ids"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleIDs  []int  `json:"role_ids"`
}

type AssignUserRolesRequest struct {
	RoleIDs []int `json:"role_ids" binding:"required"`
}

// AdminRepoBundle wraps all administrative repository interfaces for clean service initialization.
type AdminRepoBundle struct {
	ZoneRepo         ZoneAdminRepository
	CategoryRepo     CategoryAdminRepository
	MealScheduleRepo MealScheduleAdminRepository
	RBACRepo         RBACRepository
	UserRepo         UserAdminRepository
}

// Repository Interfaces

type ZoneAdminRepository interface {
	ListZones(ctx context.Context) ([]*Zone, error)
	GetZoneByID(ctx context.Context, id int) (*Zone, error)
	CreateZone(ctx context.Context, z *Zone) error
	UpdateZone(ctx context.Context, z *Zone) error
	DeleteZone(ctx context.Context, id int) error
}

type CategoryAdminRepository interface {
	ListCategories(ctx context.Context) ([]*CategoryWithZones, error)
	GetCategoryByID(ctx context.Context, id int) (*CategoryWithZones, error)
	CreateCategory(ctx context.Context, c *Category, zoneIDs []int) error
	UpdateCategory(ctx context.Context, c *Category, zoneIDs []int) error
	DeleteCategory(ctx context.Context, id int) error
	SetCategoryAllowedZones(ctx context.Context, categoryID int, zoneIDs []int) error
}

type MealScheduleAdminRepository interface {
	ListMealSchedules(ctx context.Context) ([]*MealScheduleWithCategories, error)
	GetMealScheduleByID(ctx context.Context, id int) (*MealScheduleWithCategories, error)
	CreateMealSchedule(ctx context.Context, m *MealSchedule, categoryIDs []int) error
	UpdateMealSchedule(ctx context.Context, m *MealSchedule, categoryIDs []int) error
	DeleteMealSchedule(ctx context.Context, id int) error
	SetMealScheduleCategories(ctx context.Context, mealScheduleID int, categoryIDs []int) error
}

type RBACRepository interface {
	ListRoles(ctx context.Context) ([]*Role, error)
	GetRoleByID(ctx context.Context, id int) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	CreateRole(ctx context.Context, r *Role) error
	DeleteRole(ctx context.Context, id int) error
	ListPermissions(ctx context.Context) ([]*Permission, error)
	CreatePermission(ctx context.Context, p *Permission) error
	AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error
	GetPermissionNamesForRole(ctx context.Context, roleID int) ([]string, error)
}

type UserAdminRepository interface {
	ListUsers(ctx context.Context) ([]*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, u *User, roleIDs []int) error
	UpdateUser(ctx context.Context, u *User, roleIDs []int) error
	DeleteUser(ctx context.Context, id string) error
	AssignRolesToUser(ctx context.Context, userID string, roleIDs []int) error
}

// Service Interfaces

type ZoneAdminService interface {
	ListZones(ctx context.Context) ([]*Zone, error)
	GetZoneByID(ctx context.Context, id int) (*Zone, error)
	CreateZone(ctx context.Context, req *CreateZoneRequest) (*Zone, error)
	UpdateZone(ctx context.Context, id int, req *UpdateZoneRequest) (*Zone, error)
	DeleteZone(ctx context.Context, id int) error
}

type CategoryAdminService interface {
	ListCategories(ctx context.Context) ([]*CategoryWithZones, error)
	GetCategoryByID(ctx context.Context, id int) (*CategoryWithZones, error)
	CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*CategoryWithZones, error)
	UpdateCategory(ctx context.Context, id int, req *UpdateCategoryRequest) (*CategoryWithZones, error)
	DeleteCategory(ctx context.Context, id int) error
	SetCategoryAllowedZones(ctx context.Context, categoryID int, zoneIDs []int) error
}

type MealScheduleAdminService interface {
	ListMealSchedules(ctx context.Context) ([]*MealScheduleWithCategories, error)
	GetMealScheduleByID(ctx context.Context, id int) (*MealScheduleWithCategories, error)
	CreateMealSchedule(ctx context.Context, req *CreateMealScheduleRequest) (*MealScheduleWithCategories, error)
	UpdateMealSchedule(ctx context.Context, id int, req *UpdateMealScheduleRequest) (*MealScheduleWithCategories, error)
	DeleteMealSchedule(ctx context.Context, id int) error
	SetMealScheduleCategories(ctx context.Context, mealScheduleID int, categoryIDs []int) error
}

type RBACService interface {
	ListRoles(ctx context.Context) ([]*Role, error)
	GetRoleByID(ctx context.Context, id int) (*Role, error)
	CreateRole(ctx context.Context, name, description string) (*Role, error)
	DeleteRole(ctx context.Context, id int) error
	ListPermissions(ctx context.Context) ([]*Permission, error)
	CreatePermission(ctx context.Context, name, description string) (*Permission, error)
	AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error
}

type UserAdminService interface {
	ListUsers(ctx context.Context) ([]*User, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
	CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	UpdateUser(ctx context.Context, id string, req *UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	AssignRolesToUser(ctx context.Context, userID string, roleIDs []int) error
}
