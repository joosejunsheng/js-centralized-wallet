package model

import (
	"fmt"

	"gorm.io/gorm"
)

type Model struct {
	db *gorm.DB
}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) Setup() error {
	err := m.SetupDatabase()
	if err != nil {
		return fmt.Errorf("failed to setup storage: %w", err)
	}

	err = m.migrate()
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	return nil
}

func (m *Model) SetupDatabase() error {
	err := m.connectDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return nil
}
