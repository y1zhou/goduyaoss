package db

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
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

var insertSQL = `
INSERT INTO duyaoss (
	net_provider, provider, timestamp, provider_group, remarks,
	loss, ping, google_ping, avg_speed, max_speed, udp_nat_type
)
VALUES (
	:net_provider, :provider, :timestamp, :provider_group, :remarks,
	:loss, :ping, :google_ping, :avg_speed, :max_speed, :udp_nat_type
);
`

// Row is the struct for a row to be inserted in the database
type Row struct {
	NetProvider string    `db:"net_provider"`
	Provider    string    `db:"provider"`
	Timestamp   time.Time `db:"timestamp"`
	Group       string    `db:"provider_group"`
	Remarks     string    `db:"remarks"`
	Loss        float64   `db:"loss"`
	Ping        float64   `db:"ping"`
	GooglePing  float64   `db:"google_ping"`
	AvgSpeed    float64   `db:"avg_speed"`
	MaxSpeed    float64   `db:"max_speed"`
	UDPNATType  string    `db:"udp_nat_type"`
}

// ConnectDb connects to a database, verifies with a ping, and creates the table.
func ConnectDb(dbFilename string) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", dbFilename)
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(schema)

	return db
}

// InsertRows adds rows to db in the correct format
func InsertRows(db *sqlx.DB, netProvider string, provider string, timestamp time.Time, tbl [][]string) {
	numRows, numCols := len(tbl[0]), len(tbl)

	tx := db.MustBegin()
	for i := 0; i < numRows; i++ {
		rowData := Row{
			NetProvider: netProvider,
			Provider:    provider,
			Timestamp:   timestamp,
			Group:       tbl[0][i],
			Remarks:     tbl[1][i],
			Loss:        fixPercent(tbl[2][i]),
			Ping:        fixNumber(tbl[3][i]),
			GooglePing:  fixNumber(tbl[4][i]),
			AvgSpeed:    fixSpeed(tbl[5][i]),
		}

		if numCols == 7 {
			rowData.UDPNATType = tbl[6][i]
		}
		if numCols == 8 {
			rowData.MaxSpeed = fixSpeed(tbl[6][i])
			rowData.UDPNATType = tbl[7][i]
		}

		tx.NamedExec(insertSQL, &rowData)
	}
	tx.Commit()
}

// func QueryDb(db *sqlx.DB, netProvider string, provider string) time.Time {
// 	row := db.QueryRow("SELECT timestamp FROM duyaoss WHERE net_provider=? AND provider=?", netProvider, provider)
// 	var timestamp time.Time
// 	err = row.Scan(&timestamp)
// }

func fixPercent(s string) float64 {
	// remove the percent sign at the end
	rgx := regexp.MustCompile(`%$`)
	res := rgx.ReplaceAllString(s, "")
	sNum, err := strconv.ParseFloat(res, 64)
	if err != nil {
		log.Fatal(err)
	}

	return sNum
}

// all numbers present in the table should have precision 2
func fixNumber(s string) float64 {
	if len(s) < 3 {
		return 0
	}
	correctRgx := regexp.MustCompile(`\.\d\d$`)
	var res string
	if !correctRgx.MatchString(s) {
		res = strings.ReplaceAll(s, ".", "")
		idx := len(res) - 2
		res = res[:idx] + "." + res[idx:]
	} else {
		res = s
	}

	sNum, err := strconv.ParseFloat(res, 64)
	if err != nil {
		log.Fatal(err)
	}

	return sNum
}

// AvgSpeed and MaxSpeed have units at the end (KB, MB or GB)
func fixSpeed(s string) float64 {
	if len(s) < 3 || s == "NA" {
		return 0
	}
	idx := len(s) - 2
	unit := s[idx:]
	scalar := 1.0
	switch unit {
	case "KB":
		scalar = float64(1e3)
	case "MB":
		scalar = float64(1e6)
	case "GB":
		scalar = float64(1e9)
	}
	speed := fixNumber(s[:idx])

	return speed * scalar
}
