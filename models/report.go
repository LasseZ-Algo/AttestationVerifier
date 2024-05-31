package models

import "time"

type AttestationReport struct {
	ID        uint      `gorm:"primaryKey"`
	Policy    string    `json:"policy"`
	CertChain string    `json:"cert_chain"`
	Signature string    `json:"signature"`
	CreatedAt time.Time `json:"created_at"`
}
