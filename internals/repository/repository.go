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
	GetHostList() ([]models.Hoster, error)
	SaveHoster(*models.Hoster) error
	GetUser(int64) (*models.Follower, error)
	SaveUser(*models.Follower) error
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

func (r *Repository) GetHostList() ([]models.Hoster, error) {
	data, err := r.db.GetAll("hosters")
	if err != nil {
		return nil, err
	}

	var hosters []models.Hoster
	for _, hostByte := range data {
		var hoster models.Hoster
		err := json.Unmarshal(hostByte, &hoster)
		if err != nil {
			return nil, err
		}
		hosters = append(hosters, hoster)
	}

	return hosters, nil
}

func (r *Repository) SaveHoster(hoster *models.Hoster) error {
	id := fmt.Sprintf("%d", hoster.ID)

	data, err := json.Marshal(hoster)
	if err != nil {
		return err
	}

	slog.Debug("Получена модель хостера для сохранения в БД", slog.Any("hoster", hoster))

	return r.db.SaveBytes(id, data, "hosters")
}

func (r *Repository) GetHoster(id int64) (*models.Hoster, error) {
	tg_id := fmt.Sprintf("%d", id)

	data, err := r.db.GetBytes(tg_id, "hosters")
	if err != nil {
		return nil, err
	}

	var hoster models.Hoster
	if data == nil {
		slog.Debug("Хостер не найден в БД")
		return nil, nil
	} else {
		err = json.Unmarshal(data, &hoster)
		if err != nil {
			return nil, err
		}
		return &hoster, nil
	}

}

func (r *Repository) GetUser(id int64) (*models.Follower, error) {
	tg_id := fmt.Sprintf("%d", id)

	data, err := r.db.GetBytes(tg_id, "users")
	if err != nil {
		return nil, err
	}
	slog.Debug("Пользователь успешно загружен из БД", slog.Any("user", data))

	var user models.Follower
	if data == nil {
		slog.Debug("Пользователь не найден в БД")
		return nil, nil
	} else {
		err = json.Unmarshal(data, &user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

}

func (r *Repository) SaveUser(user *models.Follower) error {
	id := fmt.Sprintf("%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(id, data, "users")
}
