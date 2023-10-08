package global

import (
	"fmt"
	"strconv"
	"strings"
)

// Javascript doesn't handle well uint64, so we convert it to string
type FrontUint64 uint64

func (u FrontUint64) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%d"`, u)), nil
}

func (u *FrontUint64) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*u = FrontUint64(v)
	return nil
}

// Entry struct for database
type Entry struct {
	// Basic:
	Time    uint64 `db:"time" json:"time" ch:"time"`
	Level   string `db:"level" json:"level" ch:"level"`
	Message string `db:"message" json:"message" ch:"message"`

	// Search columns:
	EnvID     string `db:"env_id" json:"env_id" ch:"env_id"`
	Element   string `db:"element" json:"element" ch:"element"`
	Pod       string `db:"pod" json:"pod" ch:"pod"`
	Version   string `db:"version" json:"version" ch:"version"`
	Project   string `db:"git_repo" json:"git_repo" ch:"git_repo"`
	UserEmail string `db:"user_email" json:"user_email" ch:"user_email"`

	// Tags:
	TagsString map[string]string  `db:"tags_string" json:"tags_string" ch:"tags_string"`
	TagsNumber map[string]float64 `db:"tags_number" json:"tags_number" ch:"tags_number"`

	// Other:
	Event bool `db:"-" ch:"-"`
}

type Message struct {
	// Basic info
	EnvID   string
	Element string
	Pod     string
	Version string
	Project string

	// Log data
	Data []byte

	// Parser format
	Parser string
}

func (msg Message) ToEntry() *Entry {
	return &Entry{
		EnvID:   msg.EnvID,
		Element: msg.Element,
		Pod:     msg.Pod,
		Version: msg.Version,
		Project: msg.Project,
	}
}
