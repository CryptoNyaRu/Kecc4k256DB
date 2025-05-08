package kecc4k256db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"sync"
	"time"
)

type Kecc4k256DB struct {
	db         *sql.DB
	ctx        context.Context
	updateLock sync.Mutex
}

func Open(path string) (kecc4k256DB *Kecc4k256DB, err error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?cache=shared", path))
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	const queryMaintenance = `CREATE TABLE IF NOT EXISTS maintenance (
	    methodsPage               INTEGER NOT NULL DEFAULT 0,
        methodsID                 INTEGER NOT NULL DEFAULT 0,
        methodsMaintenanceTime    INTEGER NOT NULL DEFAULT 0,
        eventsPage                INTEGER NOT NULL DEFAULT 0,
        eventsID                  INTEGER NOT NULL DEFAULT 0,
        eventsMaintenanceTime     INTEGER NOT NULL DEFAULT 0);
        INSERT INTO maintenance (methodsPage, methodsID, methodsMaintenanceTime, eventsPage, eventsID, eventsMaintenanceTime)
        SELECT 0, 0, 0, 0, 0, 0
        WHERE NOT EXISTS (SELECT 1 FROM maintenance);`
	if _, err = db.ExecContext(ctx, queryMaintenance); err != nil {
		_ = db.Close()
		return nil, err
	}

	const queryMethods = `CREATE TABLE IF NOT EXISTS methods (
		selector TEXT,
		method   TEXT,
		PRIMARY KEY (selector, method)
	) WITHOUT ROWID;`
	if _, err = db.ExecContext(ctx, queryMethods); err != nil {
		_ = db.Close()
		return nil, err
	}

	const queryEvents = `CREATE TABLE IF NOT EXISTS events (
		signature TEXT,
		event   TEXT,
		PRIMARY KEY (signature, event)
	) WITHOUT ROWID;`
	if _, err = db.ExecContext(ctx, queryEvents); err != nil {
		_ = db.Close()
		return nil, err
	}

	kecc4k256DB = &Kecc4k256DB{
		db:  db,
		ctx: ctx,
	}

	return kecc4k256DB, nil
}

func (s *Kecc4k256DB) JournalMode() (journalMode string, err error) {
	if err = s.db.QueryRowContext(s.ctx, `PRAGMA journal_mode;`).Scan(&journalMode); err != nil {
		return "", err
	}

	return journalMode, nil
}

type Maintenance struct {
	MethodsPage            int64
	MethodsID              int64
	MethodsMaintenanceTime int64
	EventsPage             int64
	EventsID               int64
	EventsMaintenanceTime  int64
}

func (s *Kecc4k256DB) Maintenance() (maintenance *Maintenance, err error) {
	const query = `SELECT methodsPage, methodsID, methodsMaintenanceTime, eventsPage, eventsID, eventsMaintenanceTime FROM maintenance LIMIT 1`

	maintenance = &Maintenance{}
	if err = s.db.QueryRowContext(s.ctx, query).Scan(&maintenance.MethodsPage, &maintenance.MethodsID, &maintenance.MethodsMaintenanceTime, &maintenance.EventsPage, &maintenance.EventsID, &maintenance.EventsMaintenanceTime); err != nil {
		return nil, err
	}

	return maintenance, nil
}
func (s *Kecc4k256DB) UpdateMethodsMaintenance(methodsPage int64, methodsID int64) (err error) {
	const query = `UPDATE maintenance SET methodsPage = ?, methodsID = ?, methodsMaintenanceTime = ? WHERE ROWID = 1`
	_, err = s.db.ExecContext(s.ctx, query, methodsPage, methodsID, time.Now().Unix())

	return err
}
func (s *Kecc4k256DB) UpdateEventsMaintenance(eventsPage int64, eventsID int64) (err error) {
	const query = `UPDATE maintenance SET eventsPage = ?, eventsID = ?, eventsMaintenanceTime = ? WHERE ROWID = 1`
	_, err = s.db.ExecContext(s.ctx, query, eventsPage, eventsID, time.Now().Unix())

	return err
}

type MethodRecord struct {
	Selector string
	Method   string
}

func (s *Kecc4k256DB) GetMethodsBySelector(selector string) (methods []string, err error) {
	const query = `SELECT method FROM methods WHERE selector = ?`
	rows, err := s.db.QueryContext(s.ctx, query, selector)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var method string
		if err = rows.Scan(&method); err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	return methods, rows.Err()
}
func (s *Kecc4k256DB) GetSelectorByMethod(method string) (selector string, err error) {
	const query = `SELECT selector FROM methods WHERE method = ?;`

	if err = s.db.QueryRowContext(s.ctx, query, method).Scan(&selector); errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return selector, err
}
func (s *Kecc4k256DB) MethodRecords() (methodRecords int64, err error) {
	const query = `SELECT COUNT(*) FROM methods`
	if err = s.db.QueryRowContext(s.ctx, query).Scan(&methodRecords); err != nil {
		return 0, err
	}

	return methodRecords, nil
}
func (s *Kecc4k256DB) UpsertMethodRecords(methodRecords []*MethodRecord) (err error) {
	tx, err := s.db.BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	const query = `INSERT INTO methods(selector, method) VALUES (?, ?) ON CONFLICT(selector, method) DO NOTHING`
	stmt, err := tx.PrepareContext(s.ctx, query)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	for _, methodRecord := range methodRecords {
		if _, err = stmt.ExecContext(s.ctx, methodRecord.Selector, methodRecord.Method); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type EventRecord struct {
	Signature string
	Event     string
}

func (s *Kecc4k256DB) GetEventBySignature(signature string) (event string, err error) {
	const query = `SELECT event FROM events WHERE signature = ?;`

	if err = s.db.QueryRowContext(s.ctx, query, signature).Scan(&event); errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return event, err
}
func (s *Kecc4k256DB) GetSignatureByEvent(event string) (signature string, err error) {
	const query = `SELECT signature FROM events WHERE event = ?;`

	if err = s.db.QueryRowContext(s.ctx, query, event).Scan(&signature); errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return signature, err
}
func (s *Kecc4k256DB) EventRecords() (eventRecords int64, err error) {
	const query = `SELECT COUNT(*) FROM events`

	if err = s.db.QueryRowContext(s.ctx, query).Scan(&eventRecords); err != nil {
		return 0, err
	}

	return eventRecords, nil
}
func (s *Kecc4k256DB) UpsertEventRecords(eventRecords []*EventRecord) (err error) {
	tx, err := s.db.BeginTx(s.ctx, nil)
	if err != nil {
		return err
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	const query = `INSERT INTO events(signature, event) VALUES (?, ?) ON CONFLICT(signature, event) DO NOTHING`
	stmt, err := tx.PrepareContext(s.ctx, query)
	if err != nil {
		return err
	}
	defer func(stmt *sql.Stmt) {
		_ = stmt.Close()
	}(stmt)

	for _, eventRecord := range eventRecords {
		if _, err = stmt.ExecContext(s.ctx, eventRecord.Signature, eventRecord.Event); err != nil {
			return err
		}
	}
	return tx.Commit()
}
