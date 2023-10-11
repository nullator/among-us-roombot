package repository

import (
	"among-us-roombot/internals/models"
	"among-us-roombot/pkg/base"
	"encoding/json"
	"fmt"
	"log/slog"
)

type Repository struct {
	db base.Base
}

type RepositoryInterface interface {
	SaveUserStatus(int64, string, string) error
	GetUserStatus(int64, string) (string, error)
	GetRoomList() ([]models.Room, error)
	AddRoom(*models.Room) error
	DeleteRoom(string) error
}

var _ RepositoryInterface = (*Repository)(nil)

func NewRepository(db base.Base) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveUserStatus(id int64, key string, value string) error {
	tg_id := fmt.Sprintf("%d", id)
	err := r.db.Save(key, value, tg_id)

	slog.Debug("Успешно сохранены данные в БД")
	return err

}

func (r *Repository) GetUserStatus(id int64, key string) (string, error) {
	tg_id := fmt.Sprintf("%d", id)

	status, err := r.db.Get(key, tg_id)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (r *Repository) GetRoomList() ([]models.Room, error) {
	data, err := r.db.GetAll("rooms")
	if err != nil {
		return nil, err
	}

	var roomList []models.Room
	for _, roomByte := range data {
		var room models.Room
		err := json.Unmarshal(roomByte, &room)
		if err != nil {
			return nil, err
		}
		roomList = append(roomList, room)
	}

	return roomList, nil
}

func (r *Repository) AddRoom(room *models.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(room.Code, data, "rooms")
}

func (r *Repository) DeleteRoom(room string) error {
	err := r.db.Delete(room, "rooms")
	return err
}
