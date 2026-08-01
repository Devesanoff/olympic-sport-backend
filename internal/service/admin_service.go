package service

import (
	"context"
	"fmt"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	adminRepo *domain.AdminRepoBundle
	rdb       *redis.Client
}

func NewAdminService(bundle *domain.AdminRepoBundle, rdb *redis.Client) *AdminService {
	return &AdminService{
		adminRepo: bundle,
		rdb:       rdb,
	}
}

// -----------------------------------------------------------------------------
// Zone Service Implementation
// -----------------------------------------------------------------------------

func (s *AdminService) ListZones(ctx context.Context) ([]*domain.Zone, error) {
	return s.adminRepo.ZoneRepo.ListZones(ctx)
}

func (s *AdminService) GetZoneByID(ctx context.Context, id int) (*domain.Zone, error) {
	return s.adminRepo.ZoneRepo.GetZoneByID(ctx, id)
}

func (s *AdminService) CreateZone(ctx context.Context, req *domain.CreateZoneRequest) (*domain.Zone, error) {
	z := &domain.Zone{
		Name:          req.Name,
		Code:          req.Code,
		RequiresInOut: req.RequiresInOut,
	}
	if err := s.adminRepo.ZoneRepo.CreateZone(ctx, z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *AdminService) UpdateZone(ctx context.Context, id int, req *domain.UpdateZoneRequest) (*domain.Zone, error) {
	z, err := s.adminRepo.ZoneRepo.GetZoneByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		z.Name = req.Name
	}
	if req.Code != "" {
		z.Code = req.Code
	}
	if req.RequiresInOut != nil {
		z.RequiresInOut = *req.RequiresInOut
	}

	if err := s.adminRepo.ZoneRepo.UpdateZone(ctx, z); err != nil {
		return nil, err
	}
	return z, nil
}

func (s *AdminService) DeleteZone(ctx context.Context, id int) error {
	return s.adminRepo.ZoneRepo.DeleteZone(ctx, id)
}

// -----------------------------------------------------------------------------
// Category Service Implementation
// -----------------------------------------------------------------------------

func (s *AdminService) ListCategories(ctx context.Context) ([]*domain.CategoryWithZones, error) {
	return s.adminRepo.CategoryRepo.ListCategories(ctx)
}

func (s *AdminService) GetCategoryByID(ctx context.Context, id int) (*domain.CategoryWithZones, error) {
	return s.adminRepo.CategoryRepo.GetCategoryByID(ctx, id)
}

func (s *AdminService) CreateCategory(ctx context.Context, req *domain.CreateCategoryRequest) (*domain.CategoryWithZones, error) {
	c := &domain.Category{
		Name:      req.Name,
		ColorCode: req.ColorCode,
		CanEat:    req.CanEat,
	}

	if err := s.adminRepo.CategoryRepo.CreateCategory(ctx, c, req.AllowedZoneIDs); err != nil {
		return nil, err
	}

	return s.adminRepo.CategoryRepo.GetCategoryByID(ctx, c.ID)
}

func (s *AdminService) UpdateCategory(ctx context.Context, id int, req *domain.UpdateCategoryRequest) (*domain.CategoryWithZones, error) {
	cwz, err := s.adminRepo.CategoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c := &cwz.Category
	if req.Name != "" {
		c.Name = req.Name
	}
	if req.ColorCode != "" {
		c.ColorCode = req.ColorCode
	}
	if req.CanEat != nil {
		c.CanEat = *req.CanEat
	}

	zoneIDs := req.AllowedZoneIDs
	if err := s.adminRepo.CategoryRepo.UpdateCategory(ctx, c, zoneIDs); err != nil {
		return nil, err
	}

	return s.adminRepo.CategoryRepo.GetCategoryByID(ctx, id)
}

func (s *AdminService) DeleteCategory(ctx context.Context, id int) error {
	return s.adminRepo.CategoryRepo.DeleteCategory(ctx, id)
}

func (s *AdminService) SetCategoryAllowedZones(ctx context.Context, categoryID int, zoneIDs []int) error {
	return s.adminRepo.CategoryRepo.SetCategoryAllowedZones(ctx, categoryID, zoneIDs)
}

// -----------------------------------------------------------------------------
// Meal Schedule Service Implementation
// -----------------------------------------------------------------------------

func (s *AdminService) ListMealSchedules(ctx context.Context) ([]*domain.MealScheduleWithCategories, error) {
	return s.adminRepo.MealScheduleRepo.ListMealSchedules(ctx)
}

func (s *AdminService) GetMealScheduleByID(ctx context.Context, id int) (*domain.MealScheduleWithCategories, error) {
	return s.adminRepo.MealScheduleRepo.GetMealScheduleByID(ctx, id)
}

func (s *AdminService) CreateMealSchedule(ctx context.Context, req *domain.CreateMealScheduleRequest) (*domain.MealScheduleWithCategories, error) {
	m := &domain.MealSchedule{
		Date:      req.Date,
		MealType:  req.MealType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := s.adminRepo.MealScheduleRepo.CreateMealSchedule(ctx, m, req.AllowedCategoryIDs); err != nil {
		return nil, err
	}

	return s.adminRepo.MealScheduleRepo.GetMealScheduleByID(ctx, m.ID)
}

func (s *AdminService) UpdateMealSchedule(ctx context.Context, id int, req *domain.UpdateMealScheduleRequest) (*domain.MealScheduleWithCategories, error) {
	ms, err := s.adminRepo.MealScheduleRepo.GetMealScheduleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	m := &ms.MealSchedule
	if req.Date != "" {
		m.Date = req.Date
	}
	if req.MealType != "" {
		m.MealType = req.MealType
	}
	if req.StartTime != "" {
		m.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		m.EndTime = req.EndTime
	}

	if err := s.adminRepo.MealScheduleRepo.UpdateMealSchedule(ctx, m, req.AllowedCategoryIDs); err != nil {
		return nil, err
	}

	return s.adminRepo.MealScheduleRepo.GetMealScheduleByID(ctx, id)
}

func (s *AdminService) DeleteMealSchedule(ctx context.Context, id int) error {
	return s.adminRepo.MealScheduleRepo.DeleteMealSchedule(ctx, id)
}

func (s *AdminService) SetMealScheduleCategories(ctx context.Context, mealScheduleID int, categoryIDs []int) error {
	return s.adminRepo.MealScheduleRepo.SetMealScheduleCategories(ctx, mealScheduleID, categoryIDs)
}

// -----------------------------------------------------------------------------
// RBAC Service Implementation
// -----------------------------------------------------------------------------

func (s *AdminService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.adminRepo.RBACRepo.ListRoles(ctx)
}

func (s *AdminService) GetRoleByID(ctx context.Context, id int) (*domain.Role, error) {
	return s.adminRepo.RBACRepo.GetRoleByID(ctx, id)
}

func (s *AdminService) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
	role := &domain.Role{
		Name:        name,
		Description: description,
	}
	if err := s.adminRepo.RBACRepo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *AdminService) DeleteRole(ctx context.Context, id int) error {
	role, err := s.adminRepo.RBACRepo.GetRoleByID(ctx, id)
	if err == nil && role != nil && s.rdb != nil {
		cacheKey := fmt.Sprintf("role:permissions:%s", role.Name)
		s.rdb.Del(ctx, cacheKey)
	}
	return s.adminRepo.RBACRepo.DeleteRole(ctx, id)
}

func (s *AdminService) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.adminRepo.RBACRepo.ListPermissions(ctx)
}

func (s *AdminService) CreatePermission(ctx context.Context, name, description string) (*domain.Permission, error) {
	p := &domain.Permission{
		Name:        name,
		Description: description,
	}
	if err := s.adminRepo.RBACRepo.CreatePermission(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *AdminService) AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	role, err := s.adminRepo.RBACRepo.GetRoleByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if err := s.adminRepo.RBACRepo.AssignPermissionsToRole(ctx, roleID, permissionIDs); err != nil {
		return err
	}

	// Invalidate Redis permissions cache for this role name so RBACMiddleware immediately picks up changes
	if s.rdb != nil {
		cacheKey := fmt.Sprintf("role:permissions:%s", role.Name)
		if err := s.rdb.Del(ctx, cacheKey).Err(); err != nil {
			log.Warn().Err(err).Str("role", role.Name).Msg("Failed to invalidate role permissions Redis cache")
		} else {
			log.Info().Str("role", role.Name).Msg("Invalidated role permissions Redis cache successfully")
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// User Management Service Implementation
// -----------------------------------------------------------------------------

func (s *AdminService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.adminRepo.UserRepo.ListUsers(ctx)
}

func (s *AdminService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.adminRepo.UserRepo.GetUserByID(ctx, id)
}

func (s *AdminService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.adminRepo.UserRepo.CreateUser(ctx, u, req.RoleIDs); err != nil {
		return nil, err
	}

	return s.adminRepo.UserRepo.GetUserByID(ctx, u.ID)
}

func (s *AdminService) UpdateUser(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	u, err := s.adminRepo.UserRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		u.PasswordHash = string(hashedPassword)
	} else {
		u.PasswordHash = ""
	}

	if err := s.adminRepo.UserRepo.UpdateUser(ctx, u, req.RoleIDs); err != nil {
		return nil, err
	}

	return s.adminRepo.UserRepo.GetUserByID(ctx, id)
}

func (s *AdminService) DeleteUser(ctx context.Context, id string) error {
	return s.adminRepo.UserRepo.DeleteUser(ctx, id)
}

func (s *AdminService) AssignRolesToUser(ctx context.Context, userID string, roleIDs []int) error {
	return s.adminRepo.UserRepo.AssignRolesToUser(ctx, userID, roleIDs)
}
