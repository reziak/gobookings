package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/reziak/gobookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
		SELECT id, room_name, created_at, updated_at 
		FROM rooms
		ORDER BY room_name ASC
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var rm models.Room
		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreateAt,
			&rm.UpdatedAt,
		)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, rm)
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

	query := `SELECT id, room_name, created_at, updated_at FROM rooms WHERE id=$1`

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

func (m *postgresDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, first_name, last_name, email, password, access_level, created_at, updated_at
		FROM users WHERE id=$1
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User

	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}
	return u, nil
}

func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE users SET first_name=$1, last_name=$2, email=$3, access_level=$4, updated_at=$5
	`

	_, err := m.DB.ExecContext(
		ctx,
		query,
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

func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	query := `SELECT id, password FROM users WHERE email=$1`

	row := m.DB.QueryRowContext(ctx, query, email)
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

func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed, rm.id, 
		rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id=rm.id)
		ORDER BY r.start_date ASC
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
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
			&i.Processed,
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

func (m *postgresDBRepo) AllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
		r.end_date, r.room_id, r.created_at, r.updated_at, rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id=rm.id)
		WHERE r.processed=0
		ORDER BY r.start_date ASC
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation
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

func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservation models.Reservation

	query := `
		SELECT r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date,
		r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		FROM reservations r
		LEFT JOIN rooms rm ON (r.room_id = rm.id)
		WHERE r.id=$1
	`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&reservation.ID,
		&reservation.FirstName,
		&reservation.LastName,
		&reservation.Email,
		&reservation.Phone,
		&reservation.StartDate,
		&reservation.EndDate,
		&reservation.RoomID,
		&reservation.CreatedAt,
		&reservation.UpdatedAt,
		&reservation.Processed,
		&reservation.Room.ID,
		&reservation.Room.RoomName,
	)

	if err != nil {
		return reservation, err
	}

	return reservation, nil
}

func (m *postgresDBRepo) UpdateReservation(u models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		UPDATE reservations SET first_name=$1, last_name=$2, email=$3, phone=$4, updated_at=$5
		WHERE id=$6
	`

	_, err := m.DB.ExecContext(
		ctx,
		query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Phone,
		time.Now(),
		u.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `DELETE FROM reservations WHERE id=$1`

	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `UPDATE reservations SET processed=$1 WHERE id=$2`

	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	query := `
		SELECT id, COALESCE(reservation_id, 0), restriction_id, room_id, start_date, end_date
		FROM room_restrictions WHERE $1 < end_date AND $2 >= 
		AND room_id=$3;
	`

	rows, err := m.DB.QueryContext(ctx, query, start, end, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err := rows.Scan(
			&r.ID,
			&r.ReservationID,
			&r.RestrictionID,
			&r.RoomID,
			&r.StartDate,
			&r.EndDate,
		)
		if err != nil {
			return nil, err
		}
		restrictions = append(restrictions, r)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return restrictions, nil
}

func (m *postgresDBRepo) InsertBlockForRoom(id int, start_date time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO room_restrictions (start_date, end_date, room_id, restriction_id,
		created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := m.DB.ExecContext(ctx, query, start_date, start_date.AddDate(0, 0, 1), id, 2, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (m *postgresDBRepo) DeleteBlockByID(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		DELETE FROM room_restrictions WHERE id=$1
	`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
