package storage

import (
	"database/sql"
	"github.com/fl64/gist-updater-bot/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s *Storage) Init() error {
	req := `
	CREATE TABLE IF NOT EXISTS bot_settings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	uid INTEGER,
	token TEXT,
	gistid TEXT,
	gistfile TEXT,
	UNIQUE (uid)
	);
	`
	statement, err := s.db.Prepare(req)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
}

// Check if settings exists
func (s *Storage) SettingsExists(uid int64) (bool, error) {
	sqlCheck := `SELECT count(*) FROM bot_settings WHERE uid = $1`
	row := s.db.QueryRow(sqlCheck, uid)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	if count == 1 {
		return true, nil
	}
	return false, nil
}

// Check if settings exists
func (s *Storage) GetSettings(uid int64) (*models.Settings, error) {
	sqlCheck := `SELECT uid, token, gistid, gistfile FROM bot_settings WHERE uid = $1`
	row := s.db.QueryRow(sqlCheck, uid)
	settings := &models.Settings{}
	err := row.Scan(&settings.Uid, &settings.Token, &settings.GistID, &settings.GistFile)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Storage) SetUserSettings(uid int64, token, gistID, gistFile string) error {
	check, err := s.SettingsExists(uid)
	if err != nil {
		return err
	}
	var sqlReq string
	if check {
		sqlReq = `UPDATE bot_settings SET token = $1, gistid = $2, gistfile = $3 WHERE uid = $4`
		_, err = s.db.Exec(sqlReq, token, gistID, gistFile, uid)
	} else {
		sqlReq = `INSERT INTO bot_settings (uid, token, gistid, gistfile) VALUES ($1, $2, $3, $4)`
		_, err = s.db.Exec(sqlReq, uid, token, gistID, gistFile)
	}
	if err != nil {
		return err
	}
	return nil
}
