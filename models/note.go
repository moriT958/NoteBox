package models

import (
	"notebox/config"
	"path/filepath"
	"time"
)

type Note struct {
	ID       int
	Title    string
	CreateAt time.Time
}

func (n *Note) CreatedAtStr() string {
	return n.CreateAt.Format(`2006-01-02`)
}

func (n *Note) GetFilePath() string {
	return filepath.Join(config.Volume(), n.Title+"-"+n.CreatedAtStr()+".md")
}

type Repository interface {
	Save(Note) (int, error)
	FindByID(int) (*Note, error)
	FindAll() ([]*Note, error)
	DeleteByID(int) error
}
