package models

import (
	"time"
)

// HostnameStatus represents the status of a hostname
type HostnameStatus string

const (
	StatusAvailable HostnameStatus = "available"
	StatusReserved  HostnameStatus = "reserved"
	StatusCommitted HostnameStatus = "committed"
	StatusReleased  HostnameStatus = "released"
)

// Hostname represents a generated hostname record
type Hostname struct {
	ID          int64          `json:"id" db:"id"`
	Name        string         `json:"name" db:"name"`
	TemplateID  int64          `json:"template_id" db:"template_id"`
	Status      HostnameStatus `json:"status" db:"status"`
	SequenceNum int            `json:"sequence_num" db:"sequence_num"`
	ReservedBy  string         `json:"reserved_by" db:"reserved_by"`
	ReservedAt  time.Time      `json:"reserved_at" db:"reserved_at"`
	CommittedBy string         `json:"committed_by,omitempty" db:"committed_by"`
	CommittedAt *time.Time     `json:"committed_at,omitempty" db:"committed_at"`
	ReleasedBy  string         `json:"released_by,omitempty" db:"released_by"`
	ReleasedAt  *time.Time     `json:"released_at,omitempty" db:"released_at"`
	DNSVerified bool           `json:"dns_verified" db:"dns_verified"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// HostnameReservationRequest represents a request to reserve a hostname
type HostnameReservationRequest struct {
	TemplateID  int64             `json:"template_id" binding:"required"`
	Params      map[string]string `json:"params,omitempty"`
	RequestedBy string            `json:"requested_by" binding:"required"`
}

// HostnameCommitRequest represents a request to commit a reserved hostname
type HostnameCommitRequest struct {
	HostnameID  int64  `json:"hostname_id" binding:"required"`
	CommittedBy string `json:"committed_by" binding:"required"`
}

// HostnameReleaseRequest represents a request to release a committed hostname
type HostnameReleaseRequest struct {
	HostnameID int64  `json:"hostname_id" binding:"required"`
	ReleasedBy string `json:"released_by" binding:"required"`
}

// DNSVerificationResult represents a DNS verification result
type DNSVerificationResult struct {
	Hostname   string    `json:"hostname"`
	Exists     bool      `json:"exists"`
	IPAddress  string    `json:"ip_address,omitempty"`
	VerifiedAt time.Time `json:"verified_at"`
}

// HostnameScanResponse represents a response to a hostname scan request
type HostnameScanResponse struct {
	Template      string   `json:"template"`
	Total         int      `json:"total"`
	Available     int      `json:"available"`
	Used          int      `json:"used"`
	UsedList      []string `json:"used_list,omitempty"`
	AvailableList []string `json:"available_list,omitempty"`
}
