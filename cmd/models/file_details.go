package models

import "time"

type FileDetails struct {
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
}
