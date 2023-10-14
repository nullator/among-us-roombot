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
	SaveRoom(*models.Room) error
	SaveDraftRoom(*models.Room) error
	GetRoom(string) (*models.Room, error)
	GetDraftRoom(string) (*models.Room, error)
	DeleteRoom(string) error
	DeleteDraftRoom(string) error
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

func (r *Repository) SaveRoom(room *models.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(room.Code, data, "rooms")
}

func (r *Repository) SaveDraftRoom(room *models.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(room.Code, data, "draft_rooms")
}

func (r *Repository) GetRoom(code string) (*models.Room, error) {
	data, err := r.db.GetBytes(code, "rooms")
	if err != nil {
		return nil, err
	}

	var room models.Room
	err = json.Unmarshal(data, &room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *Repository) GetDraftRoom(code string) (*models.Room, error) {
	data, err := r.db.GetBytes(code, "draft_rooms")
	if err != nil {
		return nil, err
	}

	var room models.Room
	err = json.Unmarshal(data, &room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func (r *Repository) DeleteRoom(room string) error {
	err := r.db.Delete(room, "rooms")
	return err
}

func (r *Repository) DeleteDraftRoom(room string) error {
	err := r.db.Delete(room, "draft_rooms")
	return err
}
