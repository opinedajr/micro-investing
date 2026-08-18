package stock

import (
	"regexp"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	tickerRegex = regexp.MustCompile(`^[A-Z0-9]{4,6}$`)
)

type Stock struct {
	ID        string `gorm:"type:varchar(36);primaryKey" json:"id"`
	Ticker    string `gorm:"type:varchar(6);uniqueIndex;not null" json:"ticker"`
	Name      string `gorm:"type:varchar(100);not null" json:"name"`
	Sector    string `gorm:"type:varchar(50);not null" json:"sector"`
	Rank      int8   `gorm:"type:tinyint;not null;check:rank >= 0 AND rank <= 10" json:"rank"`
	Website   *string `gorm:"type:text" json:"website"`
	CreatedAt string `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt string `gorm:"type:datetime;not null" json:"updated_at"`
}

func (s *Stock) TableName() string {
	return "stocks"
}

func (s *Stock) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

func (s *Stock) Validate() error {
	if len(s.Ticker) < 4 || len(s.Ticker) > 6 || !tickerRegex.MatchString(s.Ticker) {
		return ErrInvalidTicker
	}

	if len(s.Name) < 2 || len(s.Name) > 100 {
		return ErrInvalidName
	}

	if len(s.Sector) < 2 || len(s.Sector) > 50 {
		return ErrInvalidSector
	}

	if s.Rank < 0 || s.Rank > 10 {
		return ErrInvalidRank
	}

	return nil
}
