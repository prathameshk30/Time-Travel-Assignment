package service

import (
	"context"
	"errors"

	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/repository"
)

// SQLiteRecordService is the v1 API service backed by sqlite.
// It hides versioning details from v1 clients.
type SQLiteRecordService struct {
	repo repository.RecordRepository
}

func NewSQLiteRecordService(repo repository.RecordRepository) *SQLiteRecordService {
	return &SQLiteRecordService{repo: repo}
}

func (s *SQLiteRecordService) GetRecord(ctx context.Context, id int) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrRecordIDInvalid
	}

	row, err := s.repo.GetLatestVersion(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return entity.Record{}, ErrRecordDoesNotExist
		}
		return entity.Record{}, err
	}

	return entity.Record{ID: row.ID, Data: row.Data}, nil
}

func (s *SQLiteRecordService) CreateRecord(ctx context.Context, record entity.Record) error {
	if record.ID <= 0 {
		return ErrRecordIDInvalid
	}

	// make sure it doesn't already exist
	_, err := s.repo.GetLatestVersion(ctx, record.ID)
	if err == nil {
		return ErrRecordAlreadyExists
	}
	if !errors.Is(err, repository.ErrRecordNotFound) {
		return err
	}

	_, err = s.repo.CreateVersion(ctx, record.ID, record.Data)
	return err
}

func (s *SQLiteRecordService) UpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.Record, error) {
	if id <= 0 {
		return entity.Record{}, ErrRecordIDInvalid
	}

	current, err := s.repo.GetLatestVersion(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return entity.Record{}, ErrRecordDoesNotExist
		}
		return entity.Record{}, err
	}

	// merge updates into current state
	newData := make(map[string]string, len(current.Data))
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

	row, err := s.repo.CreateVersion(ctx, id, newData)
	if err != nil {
		return entity.Record{}, err
	}

	return entity.Record{ID: row.ID, Data: row.Data}, nil
}
