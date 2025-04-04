package model

import "time"

type Base struct {
	Id        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PageInfo struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
