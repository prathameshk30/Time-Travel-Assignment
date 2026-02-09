package entity

import "time"

type VersionedRecord struct {
	ID        int               `json:"id"`
	Version   int               `json:"version"`
	Data      map[string]string `json:"data"`
	CreatedAt time.Time         `json:"created_at"`
}

func (r *VersionedRecord) Copy() VersionedRecord {
	dataCopy := make(map[string]string, len(r.Data))
	for k, v := range r.Data {
		dataCopy[k] = v
	}
	return VersionedRecord{
		ID:        r.ID,
		Version:   r.Version,
		Data:      dataCopy,
		CreatedAt: r.CreatedAt,
	}
}
