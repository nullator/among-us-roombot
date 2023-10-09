package repository

import (
	"among-us-roombot/pkg/base"
	"fmt"
	"log/slog"
)

type Repository struct {
	db base.Base
}

type RepositoryInterface interface {
	SaveUserData(int64, string, string) error
	GetUserData(int64, string) (string, error)
	GetRoomList() ([]string, error)
	AddRoom(string) error
	DeleteRoom(string) error
}

var _ RepositoryInterface = (*Repository)(nil)

func NewRepository(db base.Base) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveUserData(id int64, data string, info string) error {
	tg_id := fmt.Sprintf("%d", id)
	err := r.db.Save(data, info, tg_id)

	slog.Debug("Успешно сохранены данные в БД")
	return err

}

func (r *Repository) GetUserData(id int64, data string) (string, error) {
	tg_id := fmt.Sprintf("%d", id)

	status, err := r.db.Get(data, tg_id)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (r *Repository) GetRoomList() ([]string, error) {
	var roomList []string

	return roomList, nil
}

func (r *Repository) AddRoom(room string) error {
	return nil
}

func (r *Repository) DeleteRoom(room string) error {
	return nil
}
