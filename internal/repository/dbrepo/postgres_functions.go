package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"hotel_management_system/internal/models"
	"time"
)

func (m *PostgresDBRepo) InsertReservation(res models.ReservationData) (int, error) {
	// ctx == context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	var newID int
	query := `insert into reservations(first_name,last_name,email,phone,start_date,
                          end_date,room_id,"CreatedAt","UpdatedAt")
                          values ($1, $2, $3, $4, $5, $6, $7, $8,$9) returning id`
	err := m.DB.QueryRowContext(ctx, query,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now()).Scan(&newID)
	if err != nil {
		return 0, err
	}
	return newID, nil
}
func (m *PostgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	query := `insert into room_restrictions(start_date,end_date,room_id,reservation_id,restriction_id,"CreatedAt","UpdatedAt") 
				values ($1, $2, $3, $4, $5, $6, $7)`
	_, err := m.DB.ExecContext(ctx, query,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		res.ReservationID,
		res.RestrictionID,
		time.Now(),
		time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (m *PostgresDBRepo) ParticularRoomAvailabilityByDate(start, end time.Time, roomId int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `select count(id)
			from reservations 
			where 
			start_date >= $1 and end_date <= $2 and
			room_id = $3`
	var count int
	err := m.DB.QueryRowContext(ctx, query, start, end, roomId).Scan(&count)
	if err != nil {
		return false, err
	}
	if count == 0 {
		return true, nil
	}
	return false, nil
}
func (m *PostgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		select
			r.id, r.room_name
		from
			rooms r
		where r.id not in 
		(select room_id from room_restrictions rr where $1 <= rr.end_date and $2 >= rr.start_date);
		`

	rows, err := m.DB.QueryContext(ctx, query, start, end)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room
		err := rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *PostgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `
		select id, room_name, "CreatedAt", "UpdatedAt" from rooms where id = $1
`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)

	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, access_level, "CreatedAt", "UpdatedAt, password" 
			from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.Password,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}
func (m *PostgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email = $3, access_level = $4, "UpdatedAt" = $5
`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return err
	}

	return nil
}
func (m *PostgresDBRepo) AuthUser(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&id, &hashedPassword)
	if err != nil {
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword, nil
}
func (m *PostgresDBRepo) AllReservations() ([]models.ReservationData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.ReservationData

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, 
       r.end_date, r.room_id, r."CreatedAt", r."UpdatedAt",
       rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		ORDER BY r.start_date ASC;

`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("Error in closing Database")
		}
	}(rows)

	for rows.Next() {
		var i models.ReservationData
		err := rows.Scan(
			&i.ID,
			&i.FirstName,
			&i.LastName,
			&i.Email,
			&i.Phone,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.CreatedAt,
			&i.UpdatedAt,
			//&i.Processed,
			&i.Room.ID,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, i)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}
