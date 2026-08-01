package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminRepository struct {
	db *pgxpool.Pool
}

// NewAdminRepository creates a new instance of AdminRepository.
func NewAdminRepository(db *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{db: db}
}

// -----------------------------------------------------------------------------
// Zone Repository Implementation
// -----------------------------------------------------------------------------

func (r *AdminRepository) ListZones(ctx context.Context) ([]*domain.Zone, error) {
	query := `SELECT id, name, code, requires_in_out, created_at FROM zones ORDER BY id ASC;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones: %w", err)
	}
	defer rows.Close()

	zones := make([]*domain.Zone, 0)
	for rows.Next() {
		var z domain.Zone
		if err := rows.Scan(&z.ID, &z.Name, &z.Code, &z.RequiresInOut, &z.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan zone row: %w", err)
		}
		zones = append(zones, &z)
	}
	return zones, rows.Err()
}

func (r *AdminRepository) GetZoneByID(ctx context.Context, id int) (*domain.Zone, error) {
	query := `SELECT id, name, code, requires_in_out, created_at FROM zones WHERE id = $1;`
	var z domain.Zone
	err := r.db.QueryRow(ctx, query, id).Scan(&z.ID, &z.Name, &z.Code, &z.RequiresInOut, &z.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone by id %d: %w", id, err)
	}
	return &z, nil
}

func (r *AdminRepository) CreateZone(ctx context.Context, z *domain.Zone) error {
	query := `
		INSERT INTO zones (name, code, requires_in_out)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`
	err := r.db.QueryRow(ctx, query, z.Name, z.Code, z.RequiresInOut).Scan(&z.ID, &z.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create zone: %w", err)
	}
	return nil
}

func (r *AdminRepository) UpdateZone(ctx context.Context, z *domain.Zone) error {
	query := `
		UPDATE zones
		SET name = $1, code = $2, requires_in_out = $3
		WHERE id = $4;
	`
	cmd, err := r.db.Exec(ctx, query, z.Name, z.Code, z.RequiresInOut, z.ID)
	if err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("zone not found")
	}
	return nil
}

func (r *AdminRepository) DeleteZone(ctx context.Context, id int) error {
	query := `DELETE FROM zones WHERE id = $1;`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("zone not found")
	}
	return nil
}

// -----------------------------------------------------------------------------
// Category Repository Implementation
// -----------------------------------------------------------------------------

func (r *AdminRepository) ListCategories(ctx context.Context) ([]*domain.CategoryWithZones, error) {
	query := `SELECT id, name, color_code, can_eat, created_at FROM categories ORDER BY id ASC;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*domain.CategoryWithZones, 0)
	for rows.Next() {
		var c domain.CategoryWithZones
		if err := rows.Scan(&c.ID, &c.Name, &c.ColorCode, &c.CanEat, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		c.AllowedZoneIDs = make([]int, 0)
		categories = append(categories, &c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch allowed zones for each category
	for i, cat := range categories {
		zoneIDs, err := r.getCategoryAllowedZoneIDs(ctx, cat.ID)
		if err != nil {
			return nil, err
		}
		categories[i].AllowedZoneIDs = zoneIDs
	}

	return categories, nil
}

func (r *AdminRepository) GetCategoryByID(ctx context.Context, id int) (*domain.CategoryWithZones, error) {
	query := `SELECT id, name, color_code, can_eat, created_at FROM categories WHERE id = $1;`
	var c domain.CategoryWithZones
	err := r.db.QueryRow(ctx, query, id).Scan(&c.ID, &c.Name, &c.ColorCode, &c.CanEat, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get category by id %d: %w", id, err)
	}

	zoneIDs, err := r.getCategoryAllowedZoneIDs(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.AllowedZoneIDs = zoneIDs
	return &c, nil
}

func (r *AdminRepository) getCategoryAllowedZoneIDs(ctx context.Context, categoryID int) ([]int, error) {
	query := `SELECT zone_id FROM category_allowed_zones WHERE category_id = $1 ORDER BY zone_id ASC;`
	rows, err := r.db.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query category allowed zones: %w", err)
	}
	defer rows.Close()

	zoneIDs := make([]int, 0)
	for rows.Next() {
		var zid int
		if err := rows.Scan(&zid); err != nil {
			return nil, err
		}
		zoneIDs = append(zoneIDs, zid)
	}
	return zoneIDs, rows.Err()
}

func (r *AdminRepository) CreateCategory(ctx context.Context, c *domain.Category, zoneIDs []int) error {
	query := `
		INSERT INTO categories (name, color_code, can_eat)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`
	err := r.db.QueryRow(ctx, query, c.Name, c.ColorCode, c.CanEat).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	if len(zoneIDs) > 0 {
		if err := r.SetCategoryAllowedZones(ctx, c.ID, zoneIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepository) UpdateCategory(ctx context.Context, c *domain.Category, zoneIDs []int) error {
	query := `
		UPDATE categories
		SET name = $1, color_code = $2, can_eat = $3
		WHERE id = $4;
	`
	cmd, err := r.db.Exec(ctx, query, c.Name, c.ColorCode, c.CanEat, c.ID)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	if zoneIDs != nil {
		if err := r.SetCategoryAllowedZones(ctx, c.ID, zoneIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepository) DeleteCategory(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = $1;`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func (r *AdminRepository) SetCategoryAllowedZones(ctx context.Context, categoryID int, zoneIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM category_allowed_zones WHERE category_id = $1;`, categoryID)
	if err != nil {
		return fmt.Errorf("failed to clear existing category zones: %w", err)
	}

	for _, zid := range zoneIDs {
		_, err = tx.Exec(ctx, `INSERT INTO category_allowed_zones (category_id, zone_id) VALUES ($1, $2);`, categoryID, zid)
		if err != nil {
			return fmt.Errorf("failed to insert category zone (%d, %d): %w", categoryID, zid, err)
		}
	}

	return tx.Commit(ctx)
}

// -----------------------------------------------------------------------------
// Meal Schedule Repository Implementation
// -----------------------------------------------------------------------------

func (r *AdminRepository) ListMealSchedules(ctx context.Context) ([]*domain.MealScheduleWithCategories, error) {
	query := `SELECT id, date::text, meal_type, start_time::text, end_time::text, created_at FROM meal_schedules ORDER BY date DESC, start_time ASC;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query meal schedules: %w", err)
	}
	defer rows.Close()

	schedules := make([]*domain.MealScheduleWithCategories, 0)
	for rows.Next() {
		var m domain.MealScheduleWithCategories
		var createdAt time.Time
		if err := rows.Scan(&m.ID, &m.Date, &m.MealType, &m.StartTime, &m.EndTime, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan meal schedule row: %w", err)
		}
		m.AllowedCategoryIDs = make([]int, 0)
		schedules = append(schedules, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, ms := range schedules {
		catIDs, err := r.getMealScheduleCategoryIDs(ctx, ms.ID)
		if err != nil {
			return nil, err
		}
		schedules[i].AllowedCategoryIDs = catIDs
	}

	return schedules, nil
}

func (r *AdminRepository) GetMealScheduleByID(ctx context.Context, id int) (*domain.MealScheduleWithCategories, error) {
	query := `SELECT id, date::text, meal_type, start_time::text, end_time::text, created_at FROM meal_schedules WHERE id = $1;`
	var m domain.MealScheduleWithCategories
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.Date, &m.MealType, &m.StartTime, &m.EndTime, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get meal schedule by id %d: %w", id, err)
	}

	catIDs, err := r.getMealScheduleCategoryIDs(ctx, m.ID)
	if err != nil {
		return nil, err
	}
	m.AllowedCategoryIDs = catIDs
	return &m, nil
}

func (r *AdminRepository) getMealScheduleCategoryIDs(ctx context.Context, mealScheduleID int) ([]int, error) {
	query := `SELECT category_id FROM meal_schedule_categories WHERE meal_schedule_id = $1 ORDER BY category_id ASC;`
	rows, err := r.db.Query(ctx, query, mealScheduleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query meal schedule categories: %w", err)
	}
	defer rows.Close()

	categoryIDs := make([]int, 0)
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, cid)
	}
	return categoryIDs, rows.Err()
}

func (r *AdminRepository) CreateMealSchedule(ctx context.Context, m *domain.MealSchedule, categoryIDs []int) error {
	query := `
		INSERT INTO meal_schedules (date, meal_type, start_time, end_time)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`
	err := r.db.QueryRow(ctx, query, m.Date, m.MealType, m.StartTime, m.EndTime).Scan(&m.ID)
	if err != nil {
		return fmt.Errorf("failed to create meal schedule: %w", err)
	}

	if len(categoryIDs) > 0 {
		if err := r.SetMealScheduleCategories(ctx, m.ID, categoryIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepository) UpdateMealSchedule(ctx context.Context, m *domain.MealSchedule, categoryIDs []int) error {
	query := `
		UPDATE meal_schedules
		SET date = $1, meal_type = $2, start_time = $3, end_time = $4
		WHERE id = $5;
	`
	cmd, err := r.db.Exec(ctx, query, m.Date, m.MealType, m.StartTime, m.EndTime, m.ID)
	if err != nil {
		return fmt.Errorf("failed to update meal schedule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("meal schedule not found")
	}

	if categoryIDs != nil {
		if err := r.SetMealScheduleCategories(ctx, m.ID, categoryIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepository) DeleteMealSchedule(ctx context.Context, id int) error {
	query := `DELETE FROM meal_schedules WHERE id = $1;`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete meal schedule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("meal schedule not found")
	}
	return nil
}

func (r *AdminRepository) SetMealScheduleCategories(ctx context.Context, mealScheduleID int, categoryIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM meal_schedule_categories WHERE meal_schedule_id = $1;`, mealScheduleID)
	if err != nil {
		return fmt.Errorf("failed to clear existing meal schedule categories: %w", err)
	}

	for _, cid := range categoryIDs {
		_, err = tx.Exec(ctx, `INSERT INTO meal_schedule_categories (meal_schedule_id, category_id) VALUES ($1, $2);`, mealScheduleID, cid)
		if err != nil {
			return fmt.Errorf("failed to insert meal schedule category (%d, %d): %w", mealScheduleID, cid, err)
		}
	}

	return tx.Commit(ctx)
}

// -----------------------------------------------------------------------------
// RBAC Repository Implementation
// -----------------------------------------------------------------------------

func (r *AdminRepository) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	query := `SELECT id, name, coalesce(description, ''), created_at FROM roles ORDER BY id ASC;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role row: %w", err)
		}
		roles = append(roles, &role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, role := range roles {
		perms, err := r.getPermissionsForRole(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		roles[i].Permissions = perms
		permIDs := make([]int, len(perms))
		for j, p := range perms {
			permIDs[j] = p.ID
		}
		roles[i].PermissionIDs = permIDs
	}

	return roles, nil
}

func (r *AdminRepository) GetRoleByID(ctx context.Context, id int) (*domain.Role, error) {
	query := `SELECT id, name, coalesce(description, ''), created_at FROM roles WHERE id = $1;`
	var role domain.Role
	err := r.db.QueryRow(ctx, query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by id %d: %w", id, err)
	}

	perms, err := r.getPermissionsForRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}
	role.Permissions = perms
	permIDs := make([]int, len(perms))
	for j, p := range perms {
		permIDs[j] = p.ID
	}
	role.PermissionIDs = permIDs

	return &role, nil
}

func (r *AdminRepository) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	query := `SELECT id, name, coalesce(description, ''), created_at FROM roles WHERE name = $1;`
	var role domain.Role
	err := r.db.QueryRow(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name %s: %w", name, err)
	}
	return &role, nil
}

func (r *AdminRepository) getPermissionsForRole(ctx context.Context, roleID int) ([]*domain.Permission, error) {
	query := `
		SELECT p.id, p.name, coalesce(p.description, ''), p.created_at
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.id ASC;
	`
	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	perms := make([]*domain.Permission, 0)
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, &p)
	}
	return perms, rows.Err()
}

func (r *AdminRepository) GetPermissionNamesForRole(ctx context.Context, roleID int) ([]string, error) {
	query := `
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1;
	`
	rows, err := r.db.Query(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (r *AdminRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	query := `
		INSERT INTO roles (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at;
	`
	err := r.db.QueryRow(ctx, query, role.Name, role.Description).Scan(&role.ID, &role.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

func (r *AdminRepository) DeleteRole(ctx context.Context, id int) error {
	query := `DELETE FROM roles WHERE id = $1;`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("role not found")
	}
	return nil
}

func (r *AdminRepository) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	query := `SELECT id, name, coalesce(description, ''), created_at FROM permissions ORDER BY id ASC;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	perms := make([]*domain.Permission, 0)
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan permission row: %w", err)
		}
		perms = append(perms, &p)
	}
	return perms, rows.Err()
}

func (r *AdminRepository) CreatePermission(ctx context.Context, p *domain.Permission) error {
	query := `
		INSERT INTO permissions (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at;
	`
	err := r.db.QueryRow(ctx, query, p.Name, p.Description).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

func (r *AdminRepository) AssignPermissionsToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1;`, roleID)
	if err != nil {
		return fmt.Errorf("failed to clear existing role permissions: %w", err)
	}

	for _, pid := range permissionIDs {
		_, err = tx.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2);`, roleID, pid)
		if err != nil {
			return fmt.Errorf("failed to insert role permission (%d, %d): %w", roleID, pid, err)
		}
	}

	return tx.Commit(ctx)
}

// -----------------------------------------------------------------------------
// User Repository Implementation
// -----------------------------------------------------------------------------

func (r *AdminRepository) ListUsers(ctx context.Context) ([]*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users ORDER BY created_at DESC;`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, u := range users {
		roles, err := r.getRolesForUser(ctx, u.ID)
		if err != nil {
			return nil, err
		}
		users[i].Roles = roles
		rIDs := make([]int, len(roles))
		for j, role := range roles {
			rIDs[j] = role.ID
		}
		users[i].RoleIDs = rIDs
	}

	return users, nil
}

func (r *AdminRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1;`
	var u domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id %s: %w", id, err)
	}

	roles, err := r.getRolesForUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	rIDs := make([]int, len(roles))
	for j, role := range roles {
		rIDs[j] = role.ID
	}
	u.RoleIDs = rIDs

	return &u, nil
}

func (r *AdminRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1;`
	var u domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}

	roles, err := r.getRolesForUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	u.Roles = roles
	rIDs := make([]int, len(roles))
	for j, role := range roles {
		rIDs[j] = role.ID
	}
	u.RoleIDs = rIDs

	return &u, nil
}

func (r *AdminRepository) getRolesForUser(ctx context.Context, userID string) ([]*domain.Role, error) {
	query := `
		SELECT r.id, r.name, coalesce(r.description, ''), r.created_at
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
		ORDER BY r.id ASC;
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]*domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, &role)
	}
	return roles, rows.Err()
}

func (r *AdminRepository) CreateUser(ctx context.Context, u *domain.User, roleIDs []int) error {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at;
	`
	err := r.db.QueryRow(ctx, query, u.Email, u.PasswordHash).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if len(roleIDs) > 0 {
		if err := r.AssignRolesToUser(ctx, u.ID, roleIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepository) UpdateUser(ctx context.Context, u *domain.User, roleIDs []int) error {
	var err error
	var cmd pgx.TxOptions

	if u.PasswordHash != "" {
		query := `
			UPDATE users
			SET email = $1, password_hash = $2, updated_at = NOW()
			WHERE id = $3;
		`
		_, err = r.db.Exec(ctx, query, u.Email, u.PasswordHash, u.ID)
	} else {
		query := `
			UPDATE users
			SET email = $1, updated_at = NOW()
			WHERE id = $2;
		`
		_, err = r.db.Exec(ctx, query, u.Email, u.ID)
	}

	_ = cmd
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if roleIDs != nil {
		if err := r.AssignRolesToUser(ctx, u.ID, roleIDs); err != nil {
			return err
		}
	}
	return nil
}

func (r *AdminRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1;`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *AdminRepository) AssignRolesToUser(ctx context.Context, userID string, roleIDs []int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1;`, userID)
	if err != nil {
		return fmt.Errorf("failed to clear existing user roles: %w", err)
	}

	for _, rid := range roleIDs {
		_, err = tx.Exec(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2);`, userID, rid)
		if err != nil {
			return fmt.Errorf("failed to insert user role (%s, %d): %w", userID, rid, err)
		}
	}

	return tx.Commit(ctx)
}
