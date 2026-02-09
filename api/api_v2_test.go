package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rainbowmga/timetravel/entity"
	"github.com/rainbowmga/timetravel/repository"
)

type mockService struct {
	records map[int][]entity.VersionedRecord
}

func newMock() *mockService {
	return &mockService{records: make(map[int][]entity.VersionedRecord)}
}

func (m *mockService) GetLatestRecord(ctx context.Context, id int) (entity.VersionedRecord, error) {
	recs := m.records[id]
	if len(recs) == 0 {
		return entity.VersionedRecord{}, ErrInternal
	}
	return recs[len(recs)-1], nil
}

func (m *mockService) GetRecordAtVersion(ctx context.Context, id, ver int) (entity.VersionedRecord, error) {
	for _, r := range m.records[id] {
		if r.Version == ver {
			return r, nil
		}
	}
	return entity.VersionedRecord{}, ErrInternal
}

func (m *mockService) GetRecordAtTime(ctx context.Context, id int, at time.Time) (entity.VersionedRecord, error) {
	recs := m.records[id]
	if len(recs) == 0 {
		return entity.VersionedRecord{}, ErrInternal
	}
	return recs[len(recs)-1], nil
}

func (m *mockService) ListVersions(ctx context.Context, id int) ([]repository.VersionInfo, error) {
	recs := m.records[id]
	if len(recs) == 0 {
		return nil, ErrInternal
	}
	var result []repository.VersionInfo
	for _, r := range recs {
		result = append(result, repository.VersionInfo{Version: r.Version, CreatedAt: r.CreatedAt})
	}
	return result, nil
}

func (m *mockService) CreateOrUpdateRecord(ctx context.Context, id int, updates map[string]*string) (entity.VersionedRecord, error) {
	recs := m.records[id]
	nextVer := len(recs) + 1

	data := make(map[string]string)
	if len(recs) > 0 {
		for k, v := range recs[len(recs)-1].Data {
			data[k] = v
		}
	}
	for k, v := range updates {
		if v == nil {
			delete(data, k)
		} else {
			data[k] = *v
		}
	}

	rec := entity.VersionedRecord{
		ID: id, Version: nextVer, Data: data, CreatedAt: time.Now().UTC(),
	}
	m.records[id] = append(m.records[id], rec)
	return rec, nil
}

func (m *mockService) add(id, ver int, data map[string]string) {
	m.records[id] = append(m.records[id], entity.VersionedRecord{
		ID: id, Version: ver, Data: data, CreatedAt: time.Now(),
	})
}

func TestV2_GetRecord(t *testing.T) {
	mock := newMock()
	mock.add(1, 1, map[string]string{"key": "value"})

	api := NewAPIV2(mock)
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	req := httptest.NewRequest("GET", "/api/v2/records/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestV2_GetRecord_InvalidID(t *testing.T) {
	api := NewAPIV2(newMock())
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	req := httptest.NewRequest("GET", "/api/v2/records/bad", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestV2_PostRecord_Create(t *testing.T) {
	api := NewAPIV2(newMock())
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	body := bytes.NewBufferString(`{"name":"Test"}`)
	req := httptest.NewRequest("POST", "/api/v2/records/1", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 201 {
		t.Errorf("status = %d, want 201 for new record", rec.Code)
	}
}

func TestV2_PostRecord_Update(t *testing.T) {
	mock := newMock()
	mock.add(1, 1, map[string]string{"status": "pending"})

	api := NewAPIV2(mock)
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	body := bytes.NewBufferString(`{"status":"done"}`)
	req := httptest.NewRequest("POST", "/api/v2/records/1", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200 for update", rec.Code)
	}
}

func TestV2_PostRecord_BadJSON(t *testing.T) {
	api := NewAPIV2(newMock())
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	body := bytes.NewBufferString(`{not valid}`)
	req := httptest.NewRequest("POST", "/api/v2/records/1", body)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestV2_ListVersions(t *testing.T) {
	mock := newMock()
	mock.add(1, 1, map[string]string{"v": "1"})
	mock.add(1, 2, map[string]string{"v": "2"})

	api := NewAPIV2(mock)
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	req := httptest.NewRequest("GET", "/api/v2/records/1/versions", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp VersionListResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if len(resp.Versions) != 2 {
		t.Errorf("got %d versions, want 2", len(resp.Versions))
	}
}

func TestV2_GetRecordVersion(t *testing.T) {
	mock := newMock()
	mock.add(1, 1, map[string]string{"state": "first"})
	mock.add(1, 2, map[string]string{"state": "second"})

	api := NewAPIV2(mock)
	r := mux.NewRouter()
	api.CreateRoutes(r.PathPrefix("/api/v2").Subrouter())

	req := httptest.NewRequest("GET", "/api/v2/records/1/versions/1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp entity.VersionedRecord
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Version != 1 {
		t.Errorf("version = %d, want 1", resp.Version)
	}
}
