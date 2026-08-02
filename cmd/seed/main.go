package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/internal/repository/postgres"
	"github.com/Devesanoff/olympic-sport-backend/pkg/database"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Initialize logger
	log.Info().Msg("Starting Database Seeder...")

	// 1. Load Configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load environment configuration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2. Connect PostgreSQL
	dbPool, err := database.NewPostgresPool(ctx, &cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}
	defer dbPool.Close()

	// 3. Initialize Admin Repository
	repo := postgres.NewAdminRepository(dbPool)

	// 4. Seed Permissions
	permissionsToSeed := map[string]string{
		"dashboard:view":         "Access Admin Dashboard",
		"scanner:access":         "Allow QR code scan operations",
		"participants:write":     "Create/edit participants",
		"participants:read":      "View participant data",
		"reports:read":           "Read analytical reports",
		"zones:read":             "View zone details",
		"zones:write":            "Modify zones",
		"categories:read":        "View participant categories",
		"categories:write":       "Modify categories",
		"meal_schedules:read":    "View meal schedules",
		"meal_schedules:write":   "Modify meal schedules",
		"roles:read":             "View roles and permissions",
		"roles:write":            "Modify roles and assign permissions",
		"users:read":             "View administrative users",
		"users:write":            "Modify administrative users",
	}

	existingPerms, err := repo.ListPermissions(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list permissions")
	}
	permMap := make(map[string]int)
	for _, p := range existingPerms {
		permMap[p.Name] = p.ID
	}

	for name, desc := range permissionsToSeed {
		if _, exists := permMap[name]; !exists {
			p := &domain.Permission{Name: name, Description: desc}
			if err := repo.CreatePermission(ctx, p); err != nil {
				log.Fatal().Err(err).Msgf("Failed to create permission: %s", name)
			}
			permMap[name] = p.ID
			log.Info().Str("permission", name).Msg("Seeded permission")
		}
	}

	// 5. Seed Roles
	rolesToSeed := map[string]string{
		"SuperAdmin":     "Super Administrator with full privileges",
		"Guard":          "Security Guard for scanning gates",
		"KitchenManager": "Kitchen Manager for catering gates",
		"ADMIN":          "Administrator (Compatibility Role)",
		"SCANNER":        "Scanner Service (Compatibility Role)",
	}

	existingRoles, err := repo.ListRoles(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list roles")
	}
	roleMap := make(map[string]int)
	for _, r := range existingRoles {
		roleMap[r.Name] = r.ID
	}

	for name, desc := range rolesToSeed {
		if _, exists := roleMap[name]; !exists {
			role := &domain.Role{Name: name, Description: desc}
			if err := repo.CreateRole(ctx, role); err != nil {
				log.Fatal().Err(err).Msgf("Failed to create role: %s", name)
			}
			roleMap[name] = role.ID
			log.Info().Str("role", name).Msg("Seeded role")
		}
	}

	// 6. Assign Permissions to Roles
	rolePermissions := map[string][]string{
		"SuperAdmin": {
			"dashboard:view", "scanner:access", "participants:write", "participants:read",
			"reports:read", "zones:read", "zones:write", "categories:read", "categories:write",
			"meal_schedules:read", "meal_schedules:write", "roles:read", "roles:write",
			"users:read", "users:write",
		},
		"ADMIN": {
			"dashboard:view", "scanner:access", "participants:write", "participants:read",
			"reports:read", "zones:read", "zones:write", "categories:read", "categories:write",
			"meal_schedules:read", "meal_schedules:write", "roles:read", "roles:write",
			"users:read", "users:write",
		},
		"Guard": {
			"scanner:access", "participants:read", "zones:read",
		},
		"SCANNER": {
			"scanner:access", "participants:read", "zones:read",
		},
		"KitchenManager": {
			"scanner:access", "participants:read", "meal_schedules:read", "meal_schedules:write",
		},
	}

	for roleName, permNames := range rolePermissions {
		roleID, ok := roleMap[roleName]
		if !ok {
			log.Fatal().Msgf("Role %s not found in mapping", roleName)
		}

		var permIDs []int
		for _, pName := range permNames {
			pID, ok := permMap[pName]
			if !ok {
				log.Fatal().Msgf("Permission %s not found in mapping", pName)
			}
			permIDs = append(permIDs, pID)
		}

		if err := repo.AssignPermissionsToRole(ctx, roleID, permIDs); err != nil {
			log.Fatal().Err(err).Msgf("Failed to assign permissions to role %s", roleName)
		}
		log.Info().Str("role", roleName).Int("permissions_count", len(permIDs)).Msg("Assigned permissions to role")
	}

	// 7. Seed Initial SuperAdmin User
	adminEmail := os.Getenv("SEED_ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@olympic.org"
	}
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "AdminPassword123!"
	}

	_, err = repo.GetUserByEmail(ctx, adminEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to hash user password")
			}

			u := &domain.User{
				Email:        adminEmail,
				PasswordHash: string(hashedPassword),
			}

			roleIDs := []int{roleMap["SuperAdmin"]}
			if adminID, ok := roleMap["ADMIN"]; ok {
				roleIDs = append(roleIDs, adminID)
			}

			if err := repo.CreateUser(ctx, u, roleIDs); err != nil {
				log.Fatal().Err(err).Msg("Failed to create initial SuperAdmin user")
			}
			log.Info().Str("email", adminEmail).Msg("Successfully seeded SuperAdmin user")
		} else {
			log.Fatal().Err(err).Msg("Failed to check SuperAdmin user existence")
		}
	} else {
		log.Info().Str("email", adminEmail).Msg("SuperAdmin user already exists, skipping creation")
	}

	// 8. Seed test Category, Zone, and Participant for Performance Testing
	var catID int
	err = dbPool.QueryRow(ctx, `
		INSERT INTO categories (id, name, color_code, can_eat)
		VALUES (1, 'Athlete', '#00FF00', true)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&catID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to seed test category")
	}

	var zoneID int
	err = dbPool.QueryRow(ctx, `
		INSERT INTO zones (id, name, code, requires_in_out)
		VALUES (1, 'Olympic Village Entrance', 'ZONE_A', true)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`).Scan(&zoneID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to seed test zone")
	}

	_, err = dbPool.Exec(ctx, `
		INSERT INTO category_allowed_zones (category_id, zone_id)
		VALUES (1, 1)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to seed category allowed zone")
	}

	// Generate QR token using HMAC signing helper
	hmacHelper := hmac.NewHelper(cfg.HMAC.Secret)
	testParticipantID := "123e4567-e89b-12d3-a456-426614174000"
	qrToken, err := hmacHelper.GenerateQRToken(testParticipantID, time.Now().Add(365*24*time.Hour).Unix())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to generate test QR token")
	}

	_, err = dbPool.Exec(ctx, `
		INSERT INTO participants (id, full_name, category_id, qr_token, status)
		VALUES ($1, 'Usain Bolt', 1, $2, 'ACTIVE')
		ON CONFLICT (id) DO UPDATE SET qr_token = EXCLUDED.qr_token
	`, testParticipantID, qrToken)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to seed test participant")
	}
	log.Info().Str("qr_token", qrToken).Msg("Seeded test participant Usain Bolt")

	// Save qr_token to a local text file so scripts can read it
	err = os.WriteFile("test_qr_token.txt", []byte(qrToken), 0644)
	if err != nil {
		log.Warn().Err(err).Msg("Could not save qr_token to test_qr_token.txt")
	}

	log.Info().Msg("Database Seeding Completed Successfully!")
}
