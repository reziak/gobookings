package dbrepo

import (
	"context"
	"time"

	"github.com/reziak/gobookings/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	stmt := `INSERT INTO reservations (first_name, last_name, email, phone,
					start_date, end_date, room_id, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;`

	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `INSERT INTO room_restrictions (start_date, end_date, room_id,
					reservation_id, created_at, updated_at, restriction_id)
					VALUES ($1, $2, $3, $4, $5, $6, $7);`

	_, err := m.DB.ExecContext(
		ctx,
		stmt,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		res.ReservationID,
		time.Now(),
		time.Now(),
		res.RestrictionID,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `SELECT count(id) FROM room_restrictions 
						WHERE $1 < end_date AND $2 > start_date
						AND room_id = $3;`

	row := m.DB.QueryRowContext(ctx, query, start, end, roomID)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `SELECT r.id, r.room_name
						FROM room r
						WHERE r.id NOT IN (
							SELECT room_id 
							FROM room_restrictions rr 
							WHERE $1 < rr.end_date AND $2 > rr.start_date
						);`

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

func (m *postgresDBRepo) GetRoomByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `SELECT id, room_name, created_at, updated_at FROM rooms WHERE id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreateAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room, err
	}

	return room, nil
}