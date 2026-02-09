package service

import (
	"context"
	"errors"
	"time"

	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/repository"
)

var ErrVersionNotFound = errors.New("version not found")

// VersionedRecordService is what the v2 API uses - exposes full version history
type VersionedRecordService interface {
	GetLatestRecord(ctx context.Context, id int) (entity.VersionedRecord, error)
	GetRecordAtVersion(ctx context.Context, id, version int) (entity.VersionedRecord, error)
	GetRecordAtTime(ctx context.Context, id int, at time.Time) (entity.VersionedRecord, error)
	ListVersions(ctx context.Context, id int) ([]repository.VersionInfo, error)
	CreateOrUpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.VersionedRecord, error)
}

type SQLiteVersionedService struct {
	repo repository.RecordRepository
}

func NewSQLiteVersionedService(repo repository.RecordRepository) *SQLiteVersionedService {
	return &SQLiteVersionedService{repo: repo}
}

func (s *SQLiteVersionedService) GetLatestRecord(ctx context.Context, id int) (entity.VersionedRecord, error) {
	if id <= 0 {
		return entity.VersionedRecord{}, ErrRecordIDInvalid
	}

	row, err := s.repo.GetLatestVersion(ctx, id)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return entity.VersionedRecord{}, ErrRecordDoesNotExist
	}
	if err != nil {
		return entity.VersionedRecord{}, err
	}

	return toVersionedRecord(row), nil
}

func (s *SQLiteVersionedService) GetRecordAtVersion(ctx context.Context, id, version int) (entity.VersionedRecord, error) {
	if id <= 0 {
		return entity.VersionedRecord{}, ErrRecordIDInvalid
	}
	if version <= 0 {
		return entity.VersionedRecord{}, ErrVersionNotFound
	}

	row, err := s.repo.GetVersion(ctx, id, version)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return entity.VersionedRecord{}, ErrVersionNotFound
	}
	if err != nil {
		return entity.VersionedRecord{}, err
	}

	return toVersionedRecord(row), nil
}

func (s *SQLiteVersionedService) GetRecordAtTime(ctx context.Context, id int, at time.Time) (entity.VersionedRecord, error) {
	if id <= 0 {
		return entity.VersionedRecord{}, ErrRecordIDInvalid
	}

	row, err := s.repo.GetVersionAtTime(ctx, id, at)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return entity.VersionedRecord{}, ErrRecordDoesNotExist
	}
	if err != nil {
		return entity.VersionedRecord{}, err
	}

	return toVersionedRecord(row), nil
}

func (s *SQLiteVersionedService) ListVersions(ctx context.Context, id int) ([]repository.VersionInfo, error) {
	if id <= 0 {
		return nil, ErrRecordIDInvalid
	}

	versions, err := s.repo.ListVersions(ctx, id)
	if errors.Is(err, repository.ErrRecordNotFound) {
		return nil, ErrRecordDoesNotExist
	}
	return versions, err
}

// CreateOrUpdateRecord handles both create and update - always makes a new version
func (s *SQLiteVersionedService) CreateOrUpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.VersionedRecord, error) {
	if id <= 0 {
		return entity.VersionedRecord{}, ErrRecordIDInvalid
	}

	current, err := s.repo.GetLatestVersion(ctx, id)

	var newData map[string]string

	if errors.Is(err, repository.ErrRecordNotFound) {
		// brand new record
		newData = make(map[string]string)
		for k, v := range updates {
			if v != nil {
				newData[k] = *v
			}
		}
	} else if err != nil {
		return entity.VersionedRecord{}, err
	} else {
		// updating existing - start with current data
		newData = make(map[string]string, len(current.Data))
		for k, v := range current.Data {
			newData[k] = v
		}
		for k, v := range updates {
			if v == nil {
				delete(newData, k)
			} else {
				newData[k] = *v
			}
		}
	}

	row, err := s.repo.CreateVersion(ctx, id, newData)
	if err != nil {
		return entity.VersionedRecord{}, err
	}

	return toVersionedRecord(row), nil
}

func toVersionedRecord(row *repository.RecordRow) entity.VersionedRecord {
	return entity.VersionedRecord{
		ID:        row.ID,
		Version:   row.Version,
		Data:      row.Data,
		CreatedAt: row.CreatedAt,
	}
}
