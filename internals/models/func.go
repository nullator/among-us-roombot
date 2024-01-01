package models

func (u UserList) FindUserIndexByID(users []User, id int64) int {
	for i, user := range users {
		if user.ID == id {
			return i
		}
	}
	return -1
}

// RoomList implements sort.Interface for []Room based on
func (r RoomList) Len() int {
	return len(r)
}

func (r RoomList) Less(i, j int) bool {
	return r[i].Time.Before(r[j].Time)
}

func (r RoomList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
