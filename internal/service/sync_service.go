package service

import (
	"context"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"golang.org/x/sync/errgroup"
)

type syncService struct {
	repo domain.SyncRepository
}

// NewSyncService creates a new instance of syncService.
func NewSyncService(repo domain.SyncRepository) domain.SyncService {
	return &syncService{repo: repo}
}

func (s *syncService) GetOfflinePackage(ctx context.Context) (*domain.OfflinePackage, error) {
	return s.repo.GetOfflinePackage(ctx)
}

func (s *syncService) UploadLogs(ctx context.Context, req *domain.SyncUploadLogsRequest) error {
	g, ctx := errgroup.WithContext(ctx)

	if len(req.AccessLogs) > 0 {
		g.Go(func() error {
			return s.repo.BulkInsertAccessLogs(ctx, req.AccessLogs)
		})
	}

	if len(req.MealLogs) > 0 {
		g.Go(func() error {
			return s.repo.BulkInsertMealLogs(ctx, req.MealLogs)
		})
	}

	return g.Wait()
}
