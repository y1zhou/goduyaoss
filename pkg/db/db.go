package db

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

var schema = `
CREATE TABLE IF NOT EXISTS duyaoss (
	net_provider    TEXT,
	provider        TEXT,
	timestamp       DATE,
	provider_group  TEXT,
	remarks         TEXT,
	loss            REAL,
	ping            REAL,
	google_ping     REAL,
	avg_speed       REAL,
	max_speed       REAL  DEFAULT 0,
	udp_nat_type    TEXT  DEFAULT ''
);
`

// Row is the struct for a row to be inserted in the database
type Row struct {
	NetProvider string    `db:"net_provider"`
	Provider    string    `db:"provider"`
	Timestamp   time.Time `db:"timestamp"`
	Group       string    `db:"provider_group"`
	Remarks     string    `db:"remarks"`
	Loss        float32   `db:"loss"`
	Ping        float32   `db:"ping"`
	GooglePing  float32   `db:"google_ping"`
	AvgSpeed    float32   `db:"avg_speed"`
	MaxSpeed    float32   `db:"max_speed"`
	UDPNATType  string    `db:"udp_nat_type"`
}

// ConnectDb connects to a database, verifies with a ping, and creates the table.
func ConnectDb(db *sqlx.DB, dbName string) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", dbName)
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(schema)

	return db
}

