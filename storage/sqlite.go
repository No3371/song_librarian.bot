package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"No3371.github.com/song_librarian.bot/logger"
	_ "github.com/mattn/go-sqlite3"
)

func (s *sqlite) SaveChannelMapping(cId uint64, bIDs map[int]struct{}) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	if err != nil {
		return err
	}

	var j []byte
	j, err = json.Marshal(bIDs)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		`
	SELECT C_ID
	FROM Mappings
	WHERE C_ID = %d;
	`, cId)

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return err
	}

	var exists = true
	if !rows.Next() {
		// Need to create
		exists = false
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	rows.Close()

	if exists {
		stmt := fmt.Sprintf(
			`
		UPDATE Mappings
		SET C_ID = %d, B_IDs = '%s'
		WHERE C_ID = %d;
		`, cId, j, cId)

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}
	} else {
		stmt := fmt.Sprintf(
			`
		INSERT INTO Mappings (C_ID, B_IDs)
		VALUES (%d, '%s');
		`, cId, string(j))

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *sqlite) SaveBinding(bId int, b interface{}) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()

	var j []byte
	j, err = json.Marshal(b)
	if err != nil {
		return err
	}
	// logger.Logger.Infof("[DB] Saving: %s", j)

	query := fmt.Sprintf(
		`
	SELECT Json
	FROM Bindings
	WHERE B_ID = %d;
	`, bId)

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return err
	}

	var exists = true
	if !rows.Next() {
		// Need to create
		exists = false
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	rows.Close()

	if exists {
		stmt := fmt.Sprintf(
			`
		UPDATE Bindings
		SET Json = '%s'
		WHERE B_ID = %d;
		`, string(j), bId)

		var r sql.Result
		r, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

		var ra int64
		ra, err = r.RowsAffected()
		if err != nil {
			return err
		}
		logger.Logger.Infof("[DB] %d rows affected.", ra)
	} else {
		stmt := fmt.Sprintf(
			`
		INSERT INTO Bindings (B_ID, Json)
		VALUES (%d, '%s');
		`, bId, string(j))

		var r sql.Result
		r, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

		var ra int64
		ra, err = r.RowsAffected()
		if err != nil {
			return err
		}
		logger.Logger.Infof("[DB] %d rows affected.", ra)

	}

	return nil
}

func (s *sqlite) LoadChannelMapping(cId uint64) (bIDs map[int]struct{}, err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	query := fmt.Sprintf(
		`
	SELECT B_IDs
	FROM Mappings
	WHERE C_ID = %d;
	`, cId)

	var rows *sql.Rows
	rows, err = tx.Query(query)
	if err != nil {
		return nil, err
	}

	var j string

	for rows.Next() {
		err = rows.Scan(&j)
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	rows.Close()

	err = json.Unmarshal([]byte(j), &bIDs)

	return bIDs, nil
}

func (s *sqlite) LoadBinding(bId int, b interface{}) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()

	query := fmt.Sprintf(
		`
	SELECT Json
	FROM Bindings
	WHERE B_ID = %d;
	`, bId)

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return err
	}

	var j string
	for rows.Next() {
		err = rows.Scan(&j)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	rows.Close()
	// logger.Logger.Infof("[DB] Loaded: %s", j)

	if len(j) == 0 {
		return errors.New("scanned nothing")
	}

	err = json.Unmarshal([]byte(j), b)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlite) RemoveBinding(bId int) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()

	stmt := fmt.Sprintf(
		`
	DELETE FROM Bindings
	WHERE B_ID = %d;
	`, bId)

	_, err = tx.Exec(stmt)
	if err != nil {
		return err
	}

	return nil
}

func (s *sqlite) SaveAll() (err error) {
	return nil
}

func (s *sqlite) Close() (err error) {
	return s.DB.Close()
}

func (s *sqlite) GetBindingCount() (count int, err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()

	query :=
		`
	SELECT COUNT(*)
	FROM Bindings;
	`

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		return 0, err
	}
	rows.Close()

	return count, nil

}

func (s *sqlite) SaveCommandId(defId int, cmdId uint64, version uint32) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		`
	SELECT CD
	FROM Commands
	WHERE CD = %d;
	`, defId)

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return err
	}

	var exists = true
	if !rows.Next() {
		// Need to create
		exists = false
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	rows.Close()

	if exists {
		stmt := fmt.Sprintf(
			`
		UPDATE Commands
		SET CD = %d, CMD_ID = '%s', V = %d
		WHERE CD = %d;
		`, defId, strconv.FormatUint(cmdId, 10), version, defId)

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}
	} else {
		stmt := fmt.Sprintf(
			`
		INSERT INTO Commands (CD, CMD_ID, V)
		VALUES (%d, '%s', %d);
		`, defId, strconv.FormatUint(cmdId, 10), version)

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *sqlite) LoadCommandId(defId int) (cmdId uint64, version uint32, err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	query := fmt.Sprintf(
		`
	SELECT CMD_ID, V
	FROM Commands
	WHERE CD = %d;
	`, defId)

	var rows *sql.Rows
	rows, err = tx.Query(query)
	if err != nil {
		return 0, 0, err
	}

	var j string
	var v uint64

	for rows.Next() {
		err = rows.Scan(&j, &v)
		if err != nil {
			return 0, 0, err
		}
	}
	err = rows.Err()
	if err != nil {
		return 0, 0, err
	}
	rows.Close()

	if j == "" {
		return 0, 0, err
	}

	cmdId, err = strconv.ParseUint(j, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	var _version uint64
	_version, err = strconv.ParseUint(j, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	version = uint32(_version)

	return cmdId, version, nil
}

func (s *sqlite) SaveSubState(uId uint64, state bool) (err error) {
	var stateNum int
	if state {
		stateNum = 1
	} else {
		stateNum = 0
	}
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	if err != nil {
		return err
	}

	query := fmt.Sprintf(
		`
	SELECT SUB
	FROM USERS
	WHERE U_ID = %d;
	`, uId)

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return err
	}

	var exists = true
	if !rows.Next() {
		// Need to create
		exists = false
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	rows.Close()

	if exists {
		stmt := fmt.Sprintf(
			`
		UPDATE Users
		SET SUB = %d
		WHERE U_ID = %d;
		`, stateNum, uId)

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}
	} else {
		stmt := fmt.Sprintf(
			`
		INSERT INTO Users (U_ID, SUB)
		VALUES (%d, %d);
		`, uId, stateNum)

		_, err = tx.Exec(stmt)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *sqlite) LoadSubState(uId uint64) (state bool, err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	query := fmt.Sprintf(
		`
	SELECT SUB
	FROM Users
	WHERE U_ID = %d;
	`, uId)

	var rows *sql.Rows
	rows, err = tx.Query(query)
	if err != nil {
		return false, err
	}

	var v int32 = -1

	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			return false, err
		}
	}
	err = rows.Err()
	if err != nil {
		return false, err
	}
	rows.Close()
	if v == 0 {
		return false, nil
	} else if v == 1 {
		return true, nil
	} else if v == -1 { // Not saved
		return true, nil // Defaukt to true
	} else {
		return false, errors.New("unexpected sub state stored")
	}
}

func (s *sqlite) SaveMem(tId int, slot int, data string) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()
	if err != nil {
		return err
	}

	stmt := fmt.Sprintf(
		`
	INSERT INTO Memories
		(TRACK, SLOT, TS, JSON)
	VALUES
		(%d, %d, %d, '%s')
	ON CONFLICT(TRACK,SLOT) DO UPDATE SET
		TS = excluded.TS,
		JSON = excluded.JSON;
	`, tId, slot, time.Now().UnixMilli(), data)

	if s.PRINTSTMT {
		logger.Logger.Info(stmt)
	}

	_, err = tx.Exec(stmt)
	return err
}

func (s *sqlite) GetLatestMemIndex(tId int) (slot int, err error) {

	query := fmt.Sprintf(
		`
	SELECT SLOT, MAX(TS)
	FROM Memories
	WHERE TRACK = %d;
	`, tId)

	var rows *sql.Rows
	var nullableSlot sql.NullInt32
	var nullableTs sql.NullInt64
	rows, err = s.Query(query)
	defer rows.Close()
	if err != nil {
		return 0, err
	}

	if rows.Next() {
		err = rows.Scan(&nullableSlot, &nullableTs)
		if err != nil {
			return 0, err
		}
	}

	if nullableSlot.Valid {
		slot = int(nullableSlot.Int32)
	} else {
		logger.Logger.Errorf("Invalid value")
		return 0, fmt.Errorf("INVALID")
	}

	err = rows.Err()
	if err != nil {
		return 0, err
	}

	return
}

func (s *sqlite) LoadMems(tId int, from int, to int, deserializer func(slot int, data string) error) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func() {
		if err != nil {
			logger.Logger.Errorf("[DB] Error found before tx commit: %v", err)
			err = tx.Rollback()
			if err != nil {
				logger.Logger.Errorf("[DB] Rollback error: %s", err)
			}
		} else {
			err = tx.Commit()
			if err != nil {
				logger.Logger.Errorf("[DB] Tx commit error: %s", err)
			}
		}
	}()

	query := fmt.Sprintf(
		`
	SELECT SLOT, JSON
	FROM Memories
	WHERE TRACK = %d
	AND SLOT BETWEEN %d AND %d;
	`, tId, from, to-1)

	var rows *sql.Rows
	rows, err = s.Query(query)
	if err != nil {
		return err
	}

	var slot int
	var json string
	for rows.Next() {
		err = rows.Scan(&slot, &json)
		if err != nil {
			return err
		}
		err = deserializer(slot, json)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	rows.Close()

	return
}

// func (s *sqlite) SaveMemTrack(tId int, data string) (err error) {
// 	var tx *sql.Tx
// 	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
// 		Isolation: 0,
// 		ReadOnly:  false,
// 	})
// 	defer func () {
// 		if err != nil {
// 			tx.Rollback()
// 			logger.Logger.Errorf("[DB] Rollback: %s", err)
// 		} else {
// 			err = tx.Commit()
// 			if err != nil {
// 				tx.Rollback()
// 				logger.Logger.Errorf("[DB] Rollback: %s", err)
// 			}
// 		}
// 	} ()
// 	if err != nil {
// 		return err
// 	}

// 	stmt := fmt.Sprintf(
// 	`
// 	INSERT INTO MemTrackCaches
// 		(TRACK, DATA)
// 	VALUES
// 		(%d, '%s')
// 	ON DUPLICATE KEY UPDATE
// 		TRACK = VALUES(TRACK),
// 		DATA = VALUES(DATA);
// 	`, tId, data)

// 	if s.PRINTSTMT {
// 		logger.Logger.Info(stmt)
// 	}
// 	_, err = tx.Exec(stmt)
// 	return err
// }

// func (s *sqlite) LoadMemTrack (tId int) (trackJson string, err error) {
// 	var tx *sql.Tx
// 	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
// 		Isolation: 0,
// 		ReadOnly:  true,
// 	})
// 	defer func () {
// 		if err != nil {
// 			tx.Rollback()
// 			logger.Logger.Errorf("[DB] Rollback: %s", err)
// 		} else {
// 			err = tx.Commit()
// 			if err != nil {
// 				tx.Rollback()
// 				logger.Logger.Errorf("[DB] Rollback: %s", err)
// 			}
// 		}
// 	} ()

// 	query := fmt.Sprintf(
// 	`
// 	SELECT DATA
// 	FROM MemTrackCaches
// 	WHERE TRACK = %d;
// 	`, tId)

// 	var rows *sql.Rows
// 	rows, err = s.Query(query)
// 	if err != nil {
// 		return "", err
// 	}

// 	var j string
// 	for rows.Next() {
// 		err = rows.Scan(&j)
// 		if err != nil {
// 			return "", err
// 		}
// 	}
// 	err = rows.Err()
// 	if err != nil {
// 		return "", err
// 	}
// 	rows.Close()
// 	logger.Logger.Infof("[DB] Loaded: %s", j)

// 	if len(j) == 0 {
// 		return "", errors.New("scanned nothing")
// 	}

// 	err = json.Unmarshal([]byte(j), &trackJson)
// 	if err != nil {
// 		return "", err
// 	}

// 	return
// }

type sqlite struct {
	*sql.DB
	PRINTSTMT bool
}

func Sqlite(debug bool) (sv *sqlite, err error) {
	var s *sql.DB
	s, err = sql.Open("sqlite3", "./db")
	if err != nil {
		return nil, err
	}

	// var r *sql.Rows
	// r, err = s.Query(`
	// SELECT name FROM sqlite_master WHERE type='table';
	// `)
	// if err != nil {
	// 	logger.Logger.Fatalf("%s", err)
	// }

	// var tName string
	// func () {
	// 	defer r.Close()
	// 	for r.Next() {
	// 		err = r.Scan(&tName)
	// 		if err != nil {
	// 			logger.Logger.Fatalf("%s", err)
	// 		}
	// 		// logger.Logger.Infof("[Storage] %s", tName)
	// 		switch tName {
	// 		case "Mappings":
	// 			tableMappingsFound = true
	// 			break
	// 		case "Bindings":
	// 			tableBindingsFound = true
	// 			break
	// 		case "Commands":
	// 			tableCommandsFound = true
	// 			break
	// 		}
	// 	}
	// } ()

	sv = &sqlite{
		s,
		debug,
	}

	err = sv.tx(`
	CREATE TABLE IF NOT EXISTS Mappings (
		C_ID string,
		B_IDs string
	)
	`)
	if err != nil {
		logger.Logger.Fatalf("[Storage] Failed to create table \"Mappings\": %s", err)
	} else {
		logger.Logger.Infof("[Storage] Ensured table \"Mappings\".")
	}

	err = sv.tx(`
	CREATE TABLE IF NOT EXISTS Bindings (
		B_ID int,
		Json string
	)
	`)
	if err != nil {
		logger.Logger.Fatalf("[Storage] Failed to create table \"Bindings\": %s", err)
	} else {
		logger.Logger.Infof("[Storage] Ensured table \"Bindings\".")
	}

	err = sv.tx(`
	CREATE TABLE IF NOT EXISTS Commands (
		CD int,
		CMD_ID string,
		V int
	)
	`)
	if err != nil {
		logger.Logger.Fatalf("[Storage] Failed to create table \"Commands\": %s", err)
	} else {
		logger.Logger.Infof("[Storage] Ensured table \"Commands\".")
	}

	err = sv.tx(`
	CREATE TABLE IF NOT EXISTS Users (
		U_ID int NOT NULL,
		SUB bool NOT NULL,
		PRIMARY KEY (U_ID)
	)
	`)
	if err != nil {
		logger.Logger.Fatalf("[Storage] Failed to create table \"Users\": %s", err)
	} else {
		logger.Logger.Infof("[Storage] Ensured table \"Users\".")
	}

	err = sv.tx(`
	CREATE TABLE IF NOT EXISTS Memories (
		TRACK int NOT NULL,
		SLOT int NOT NULL,
		JSON string NOT NULL,
		TS int NOT NULL,
		PRIMARY KEY (TRACK, SLOT)
	)
	`)
	if err != nil {
		logger.Logger.Fatalf("[Storage] Failed to create table \"Memories\": %s", err)
	} else {
		logger.Logger.Infof("[Storage] Ensured table \"Memories\".")
	}
	return sv, err
}

func (s *sqlite) tx(stmt string) error {
	var tx *sql.Tx
	var err error
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}

	_, err = tx.Exec(stmt)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
