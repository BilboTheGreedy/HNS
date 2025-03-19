package models

import (
	"time"
)

// TemplateGroup represents a group within a template
type TemplateGroup struct {
	ID          int64  `json:"id" db:"id"`
	TemplateID  int64  `json:"template_id" db:"template_id"`
	Name        string `json:"name" db:"name"`
	Length      int    `json:"length" db:"length"`
	Position    int    `json:"position" db:"position"`
	IsRequired  bool   `json:"is_required" db:"is_required"`
	ValidationType string `json:"validation_type" db:"validation_type"`
	ValidationValue string `json:"validation_value" db:"validation_value"`
}

// TemplateValidationType represents the type of validation for a template group
type TemplateValidationType string

const (
	ValidationTypeRegex   TemplateValidationType = "regex"
	ValidationTypeList    TemplateValidationType = "list"
	ValidationTypeFixed   TemplateValidationType = "fixed"
	ValidationTypeSequence TemplateValidationType = "sequence"
)

// Template represents a hostname template
type Template struct {
	ID                int64         `json:"id" db:"id"`
	Name              string        `json:"name" db:"name"`
	Description       string        `json:"description" db:"description"`
	MaxLength         int           `json:"max_length" db:"max_length"`
	Groups            []TemplateGroup `json:"groups,omitempty"`
	SequenceStart     int           `json:"sequence_start" db:"sequence_start"`
	SequenceLength    int           `json:"sequence_length" db:"sequence_length"`
	SequencePadding   bool          `json:"sequence_padding" db:"sequence_padding"`
	SequenceIncrement int           `json:"sequence_increment" db:"sequence_increment"`
	SequencePosition  int           `json:"sequence_position" db:"sequence_position"`
	CreatedBy         string        `json:"created_by" db:"created_by"`
	CreatedAt         time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at" db:"updated_at"`
	IsActive          bool          `json:"is_active" db:"is_active"`
}

// TemplateCreateRequest represents a request to create a new template
type TemplateCreateRequest struct {
	Name              string        `json:"name" binding:"required"`
	Description       string        `json:"description"`
	MaxLength         int           `json:"max_length" binding:"required,min=1,max=64"`
	Groups            []TemplateGroupRequest `json:"groups" binding:"required,dive"`
	SequenceStart     int           `json:"sequence_start" binding:"required,min=0"`
	SequenceLength    int           `json:"sequence_length" binding:"required,min=1,max=10"`
	SequencePadding   bool          `json:"sequence_padding"`
	SequenceIncrement int           `json:"sequence_increment" binding:"required,min=1"`
	CreatedBy         string        `json:"created_by" binding:"required"`
}

// TemplateGroupRequest represents a request to create or update a template group
type TemplateGroupRequest struct {
	Name             string `json:"name" binding:"required"`
	Length           int    `json:"length" binding:"required,min=1"`
	IsRequired       bool   `json:"is_required"`
	ValidationType   string `json:"validation_type" binding:"required,oneof=regex list fixed sequence"`
	ValidationValue  string `json:"validation_value"`
}

// TemplateUpdateRequest represents a request to update an existing template
type TemplateUpdateRequest struct {
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	MaxLength         int           `json:"max_length" binding:"min=1,max=64"`
	Groups            []TemplateGroupRequest `json:"groups" binding:"dive"`
	SequenceStart     int           `json:"sequence_start" binding:"min=0"`
	SequenceLength    int           `json:"sequence_length" binding:"min=1,max=10"`
	SequencePadding   bool          `json:"sequence_padding"`
	SequenceIncrement int           `json:"sequence_increment" binding:"min=1"`
	UpdatedBy         string        `json:"updated_by" binding:"required"`
}