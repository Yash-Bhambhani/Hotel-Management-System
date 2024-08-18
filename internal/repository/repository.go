package repository

import (
	"hotel_management_system/internal/models"
	"time"
)

type DatabaseRepo interface {
	InsertReservation(res models.ReservationData) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) error
	ParticularRoomAvailabilityByDate(start, end time.Time, roomId int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)

	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	AuthUser(email, testPassword string) (int, string, error)
	AllReservations() ([]models.ReservationData, error)
}
