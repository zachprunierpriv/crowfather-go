package database

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSaveThread(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	svc := &DatabaseService{db: db}
	mock.ExpectExec("INSERT INTO threads").WithArgs("ctx", "tid").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := svc.SaveThread("ctx", "tid"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetThread(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	svc := &DatabaseService{db: db}
	rows := sqlmock.NewRows([]string{"thread_id"}).AddRow("tid")
	mock.ExpectQuery("SELECT thread_id FROM threads").WithArgs("ctx").WillReturnRows(rows)

	id, err := svc.GetThread("ctx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "tid" {
		t.Errorf("expected tid got %s", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	svc := &DatabaseService{db: db}
	mock.ExpectExec("INSERT INTO messages").WithArgs("tid", "user", "mid", "hello").WillReturnResult(sqlmock.NewResult(1, 1))

	if err := svc.SaveMessage("tid", "user", "mid", "hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
