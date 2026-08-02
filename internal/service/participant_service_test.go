package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
)

type mockParticipantRepository struct {
	participants map[string]*domain.Participant
	createErr    error
	getErr       error
}

func (m *mockParticipantRepository) Create(ctx context.Context, p *domain.Participant) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.participants[p.ID] = p
	return nil
}

func (m *mockParticipantRepository) GetByID(ctx context.Context, id string) (*domain.Participant, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.participants[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *mockParticipantRepository) List(ctx context.Context, limit, offset int) ([]*domain.Participant, int, error) {
	var list []*domain.Participant
	for _, p := range m.participants {
		list = append(list, p)
	}
	return list, len(m.participants), nil
}

func (m *mockParticipantRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Participant, error) {
	var list []*domain.Participant
	for _, id := range ids {
		if p, ok := m.participants[id]; ok {
			list = append(list, p)
		}
	}
	return list, nil
}

func (m *mockParticipantRepository) GetByCategoryID(ctx context.Context, categoryID int) ([]*domain.Participant, error) {
	var list []*domain.Participant
	for _, p := range m.participants {
		if p.CategoryID == categoryID && p.Status == domain.ParticipantStatusActive {
			list = append(list, p)
		}
	}
	return list, nil
}

func TestParticipantService(t *testing.T) {
	hmacHelper := hmac.NewHelper("testsecretsigningkey")
	repo := &mockParticipantRepository{
		participants: make(map[string]*domain.Participant),
	}
	svc := NewParticipantService(repo, hmacHelper)

	// 1. Test Create Participant
	p, err := svc.Create(context.Background(), "John Doe", 1)
	if err != nil {
		t.Fatalf("Failed to create participant: %v", err)
	}

	if p.FullName != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", p.FullName)
	}
	if p.CategoryID != 1 {
		t.Errorf("Expected CategoryID 1, got %d", p.CategoryID)
	}
	if p.QRToken == "" {
		t.Error("QRToken is empty")
	}
	if p.Status != domain.ParticipantStatusActive {
		t.Errorf("Expected status ACTIVE, got %v", p.Status)
	}

	// Verify QR token contains participant ID and is valid
	parsedID, err := hmacHelper.ValidateQRToken(p.QRToken)
	if err != nil {
		t.Fatalf("Generated QR token failed validation: %v", err)
	}
	if parsedID != p.ID {
		t.Errorf("Expected QR token to parse to participant ID %s, got %s", p.ID, parsedID)
	}

	// 2. Test GetByID
	retrieved, err := svc.GetByID(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("Failed to get participant: %v", err)
	}
	if retrieved.ID != p.ID {
		t.Errorf("Expected participant ID %s, got %s", p.ID, retrieved.ID)
	}

	// 3. Test List
	list, total, err := svc.List(context.Background(), 10, 0)
	if err != nil {
		t.Fatalf("Failed to list participants: %v", err)
	}
	if total != 1 {
		t.Errorf("Expected total count 1, got %d", total)
	}
	if len(list) != 1 {
		t.Errorf("Expected list length 1, got %d", len(list))
	}
}
