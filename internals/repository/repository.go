package repository

import (
	"among-us-roombot/internals/models"
	"among-us-roombot/pkg/base"
	"fmt"
	"log/slog"
)

type Repository struct {
	db base.Base
}

type RepositoryInterface interface {
	SaveUserStatus(int64, string) error
	GetUserStatus(int64) (string, error)
	GetRoomList() ([]string, error)
	AddRoom(*models.Room) error
	DeleteRoom(string) error
}

var _ RepositoryInterface = (*Repository)(nil)

func NewRepository(db base.Base) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveUserStatus(id int64, status string) error {
	tg_id := fmt.Sprintf("%d", id)
	err := r.db.Save("status", status, tg_id)

	slog.Debug("Успешно сохранены данные в БД")
	return err

}

func (r *Repository) GetUserStatus(id int64) (string, error) {
	tg_id := fmt.Sprintf("%d", id)

	status, err := r.db.Get("status", tg_id)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (r *Repository) GetRoomList() ([]string, error) {
	var roomList []string

	return roomList, nil
}

func (r *Repository) AddRoom(room *models.Room) error {
	return nil
}

func (r *Repository) DeleteRoom(room string) error {
	return nil
}
