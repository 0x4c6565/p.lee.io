package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0x4c6565/p.lee.io/pkg/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Storage interface {
	Get(ctx context.Context, id string) (*model.Paste, error)
	Add(ctx context.Context, paste model.Paste) (string, error)
	Delete(ctx context.Context, id string) error
	GetExpired(ctx context.Context) ([]*model.Paste, error)
}

type InMemoryStorage struct {
	pastes map[string]*model.Paste
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		pastes: make(map[string]*model.Paste),
	}
}

func (s *InMemoryStorage) Get(ctx context.Context, id string) (*model.Paste, error) {
	if val, ok := s.pastes[id]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("Paste not found")
}
func (s *InMemoryStorage) Add(ctx context.Context, paste model.Paste) (string, error) {
	id := uuid.NewString()
	paste.ID = id
	s.pastes[id] = &paste
	return id, nil
}
func (s *InMemoryStorage) Delete(ctx context.Context, id string) error {
	if _, ok := s.pastes[id]; ok {
		delete(s.pastes, id)
		return nil
	}

	return fmt.Errorf("Paste not found")
}

func (s *InMemoryStorage) GetExpired(ctx context.Context) ([]*model.Paste, error) {
	return nil, nil
}

type SQLStorage struct {
	conn *sqlx.DB
}

func NewSQLStorage(host string, port int, user string, password string, db string) (*SQLStorage, error) {
	conn, err := sqlx.Connect("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, host, port, db))
	if err != nil {
		return nil, err
	}

	return &SQLStorage{
		conn: conn,
	}, nil
}

func (s *SQLStorage) Get(ctx context.Context, id string) (*model.Paste, error) {
	p := model.Paste{}
	err := s.conn.GetContext(ctx, &p, "SELECT * FROM paste WHERE id=?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewNotFoundError(fmt.Sprintf("Paste not found with ID %s", id))
		}
		return nil, err
	}

	return &p, nil
}

func (s *SQLStorage) Add(ctx context.Context, paste model.Paste) (string, error) {
	id := uuid.NewString()

	_, err := s.conn.ExecContext(ctx, "INSERT INTO paste (id, timestamp, expires, content, syntax) VALUES (?,?,?,?,?)", id, paste.Timestamp, paste.Expires, paste.Content, paste.Syntax)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *SQLStorage) Delete(ctx context.Context, id string) error {
	_, err := s.conn.ExecContext(ctx, "DELETE FROM paste WHERE id=?", id)
	return err
}

func (s *SQLStorage) GetExpired(ctx context.Context) ([]*model.Paste, error) {
	var expiredPastes []*model.Paste
	err := s.conn.SelectContext(ctx, &expiredPastes, "SELECT *, (expires+timestamp) AS expire_timestamp FROM paste WHERE expires > 0 HAVING expire_timestamp <= ? ", time.Now().Unix())
	if err != nil {
		return nil, err
	}

	return expiredPastes, nil
}
