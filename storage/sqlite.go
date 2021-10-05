package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"No3371.github.com/song_librarian.bot/logger"
	_ "github.com/mattn/go-sqlite3"
)

func (s *sqlite) SaveChannelMapping(cId uint64, bIDs map[int]struct{}) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	defer func () {
		if err != nil {
			tx.Rollback()
			logger.Logger.Errorf("[DB] Rollback: %s", err)
		} else {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				logger.Logger.Errorf("[DB] Rollback: %s", err)
			}
		}
	} ()
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

func (s *sqlite) SaveBinding(bId int, b interface {}) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer func () {
		if err != nil {
			tx.Rollback()
			logger.Logger.Errorf("[DB] Rollback: %s", err)
		} else {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				logger.Logger.Errorf("[DB] Rollback: %s", err)
			}
		}
	} ()

	var j []byte
	j, err = json.Marshal(b)
	if err != nil {
		return err
	}
	logger.Logger.Infof("[DB] Saving: %s", j)

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

func (s *sqlite) LoadChannelMapping (cId uint64) (bIDs map[int]struct{}, err error)  {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func () {
		if err != nil {
			tx.Rollback()
			logger.Logger.Errorf("[DB] Rollback: %s", err)
		} else {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				logger.Logger.Errorf("[DB] Rollback: %s", err)
			}
		}
	} ()
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

func (s *sqlite) LoadBinding (bId int, b interface{}) (err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func () {
		if err != nil {
			tx.Rollback()
			logger.Logger.Errorf("[DB] Rollback: %s", err)
		} else {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				logger.Logger.Errorf("[DB] Rollback: %s", err)
			}
		}
	} ()

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
	logger.Logger.Infof("[DB] Loaded: %s", j)

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
	defer func () {
		if err != nil {
			tx.Rollback()
			logger.Logger.Errorf("[DB] Rollback: %s", err)
		} else {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				logger.Logger.Errorf("[DB] Rollback: %s", err)
			}
		}
	} ()

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

func (s *sqlite) GetBindingCount () (count int, err error) {
	var tx *sql.Tx
	tx, err = s.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  true,
	})
	defer func () {
		if err != nil {
			tx.Rollback()
			logger.Logger.Errorf("[DB] Rollback: %s", err)
		} else {
			err = tx.Commit()
			if err != nil {
				tx.Rollback()
				logger.Logger.Errorf("[DB] Rollback: %s", err)
			}
		}
	} ()

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

type sqlite struct {
	*sql.DB
}

func Sqlite () (sv *sqlite, err error) {
	var s *sql.DB
	s, err = sql.Open("sqlite3", "./db")
	if err != nil {
		return nil, err
	}

	var r *sql.Rows
	r, err = s.Query(`
	SELECT Count(*) FROM sqlite_master WHERE type='table' AND (name='Bindings' OR name='Mappings');
	`)
	if err != nil {
		logger.Logger.Fatalf("%s", err)
	}

	var count int
	func () {
		defer r.Close()
		if r.Next() {
			err = r.Scan(&count)
			if err != nil {
				logger.Logger.Fatalf("%s", err)
			}
		} else {
			logger.Logger.Fatalf("sqlite: Failed to get count")
		}
	} ()

	sv = &sqlite{
		s,
	}

	if count == 0 {
		err = sv.tx(`
		CREATE TABLE Bindings (
			B_ID int,
			Json string
		)
		`)
		if err != nil {
			logger.Logger.Fatalf("[Storage] Failed to create table of Bindings: %s", err)
		}
		err = sv.tx(`
		CREATE TABLE Mappings (
			C_ID string,
			B_IDs string
		)
		`)
		if err != nil {
			logger.Logger.Fatalf("[Storage] Failed to create table of Mapping: %s", err)
		}
	} else if count != 2 {
		logger.Logger.Fatalf("[Storage] Did not found exactly 2 tables (Mappings & Bindings)")
	}

	return sv, nil
}

func (s *sqlite) tx (stmt string) error {
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