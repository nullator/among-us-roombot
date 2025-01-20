package repository

import (
	"among-us-roombot/internals/models"
	"among-us-roombot/pkg/base"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/boltdb/bolt"
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
	GetAndUpdateUserRequestTimestamp(id int64) (time.Time, error)
}

var _ RepositoryInterface = (*Repository)(nil)

func NewRepository(db base.Base) *Repository {
	return &Repository{db: db}
}

// SaveUserStatus сохраняет статус пользователя в БД
// id - telegram id пользователя
// key - ключ по которому сохраняется статус (используется для отслеживания состояния пользователя)
// value - значение статуса
func (r *Repository) SaveUserStatus(id int64, key string, value string) error {
	tg_id := fmt.Sprintf("%d", id)
	err := r.db.Save(key, value, tg_id)

	slog.Debug("Успешно сохранены данные в БД")
	return err

}

// GetUserStatus возвращает статус пользователя из БД
// id - telegram id пользователя
func (r *Repository) GetUserStatus(id int64, key string) (string, error) {
	tg_id := fmt.Sprintf("%d", id)

	status, err := r.db.Get(key, tg_id)
	if err != nil {
		return "", err
	}

	return status, nil
}

// GetRoomList возвращает список комнат из БД
// комнаты сохраняются в бакет "rooms"
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

// SaveRoom сохраняет комнату в БД
// модель комнаты содержит код room.Code, по которому она сохраняется в БД
func (r *Repository) SaveRoom(room *models.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(room.Code, data, "rooms")
}

// SaveDraftRoom сохраняет черновик комнаты в БД
// черновик комнаты сохраняется в отдельный бакет "draft_rooms"
// модель комнаты содержит код room.Code, по которому она сохраняется в БД
func (r *Repository) SaveDraftRoom(room *models.Room) error {
	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(room.Code, data, "draft_rooms")
}

// GetRoom возвращает комнату из БД по ее коду
// код - код из игры амонг
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

// GetDraftRoom возвращает черновик комнаты из БД по ее коду
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

// DeleteRoom удаляет комнату из БД по ее коду
// room - код из игры амонг
func (r *Repository) DeleteRoom(room string) error {
	err := r.db.Delete(room, "rooms")
	return err
}

// DeleteDraftRoom удаляет черновик комнаты из БД по ее коду
// room - код из игры амонг
func (r *Repository) DeleteDraftRoom(room string) error {
	err := r.db.Delete(room, "draft_rooms")
	return err
}

// GetHostList возвращает список хостеров из БД
// хостеры храняться в бакете "hosters"
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

// SaveHoster сохраняет хостера в БД
// модель хостера содержит id hoster.ID, по которому он сохраняется в БД
func (r *Repository) SaveHoster(hoster *models.Hoster) error {
	id := fmt.Sprintf("%d", hoster.ID)

	data, err := json.Marshal(hoster)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(id, data, "hosters")
}

// GetHoster возвращает хостера из БД по его telegram id
func (r *Repository) GetHoster(id int64) (*models.Hoster, error) {
	tg_id := fmt.Sprintf("%d", id)

	data, err := r.db.GetBytes(tg_id, "hosters")
	if err != nil {
		return nil, err
	}

	var hoster models.Hoster
	// если хостер не найден в БД, то возвращается nil (не ошибка)
	// это нужно для того, чтобы в случае отсутствия хостера в БД, создать нового
	// значение nil обрабатывается в месте вызова функции
	if data == nil {
		return nil, nil
	} else {
		err = json.Unmarshal(data, &hoster)
		if err != nil {
			return nil, err
		}
		return &hoster, nil
	}

}

// GetUser возвращает пользователя из БД по его telegram id
func (r *Repository) GetUser(id int64) (*models.Follower, error) {
	tg_id := fmt.Sprintf("%d", id)

	data, err := r.db.GetBytes(tg_id, "users")
	if err != nil {
		return nil, err
	}

	var user models.Follower
	// если пользователь не найден в БД, то возвращается nil (не ошибка)
	// это нужно для того, чтобы в случае отсутствия пользователя в БД, создать нового
	// значение nil обрабатывается в месте вызова функции
	if data == nil {
		return nil, nil
	} else {
		err = json.Unmarshal(data, &user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

}

// SaveUser сохраняет пользователя в БД
// модель пользователя содержит id user.ID, по которому он сохраняется в БД
func (r *Repository) SaveUser(user *models.Follower) error {
	id := fmt.Sprintf("%d", user.ID)

	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return r.db.SaveBytes(id, data, "users")
}

// UserRequestTimestamps хранит временные отметки запросов пользователя, возвращает время 3-го по счету запроса
func (r *Repository) GetAndUpdateUserRequestTimestamp(id int64) (time.Time, error) {
	tg_id := fmt.Sprintf("%d", id)
	var timestamps models.UserRequestTimestamps

	// Получаем текущие временные отметки из базы данных
	data, err := r.db.GetBytes("request_timestamps", tg_id)
	if err != nil && err != bolt.ErrBucketNotFound {
		return time.Time{}, err
	}

	if data != nil {
		err = json.Unmarshal(data, &timestamps)
		if err != nil {
			return time.Time{}, err
		}
	}

	// Если временных отметок нет, возвращаем временную отметку 24 часа назад
	var lastTimestamp time.Time
	if len(timestamps.Timestamps) == 0 {
		lastTimestamp = time.Now().Add(-24 * time.Hour)
	} else if len(timestamps.Timestamps) < 3 {
		lastTimestamp = time.Now().Add(-24 * time.Hour)
	} else {
		lastTimestamp = timestamps.Timestamps[len(timestamps.Timestamps)-1]
	}

	// Добавляем текущую временную отметку
	timestamps.Timestamps = append(timestamps.Timestamps, time.Now())

	// Оставляем только последние три временные отметки
	if len(timestamps.Timestamps) > 3 {
		timestamps.Timestamps = timestamps.Timestamps[len(timestamps.Timestamps)-3:]
	}

	// Сохраняем обновленные временные отметки в базу данных
	data, err = json.Marshal(timestamps)
	if err != nil {
		return time.Time{}, err
	}

	err = r.db.SaveBytes("request_timestamps", data, tg_id)
	if err != nil {
		return time.Time{}, err
	}

	return lastTimestamp, nil
}
