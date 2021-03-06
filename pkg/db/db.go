package db

import (
	"database/sql"
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

var querySQL = `
SELECT timestamp FROM duyaoss
WHERE
	net_provider =? AND provider =?
ORDER BY
	timestamp DESC LIMIT 1;
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

// connectDb connects to a database, verifies with a ping, and creates the table.
func connectDb(dbFilename string) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", dbFilename)
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(schema)

	return db
}

// InsertRows adds rows to db in the correct format
func InsertRows(dbName string, netProvider string, provider string, timestamp time.Time, tbl [][]string) {
	DB := connectDb(dbName)
	defer DB.Close()
	numRows, numCols := len(tbl[0]), len(tbl)

	tx := DB.MustBegin()
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

		_, err := tx.NamedExec(insertSQL, &rowData)
		if err != nil {
			log.Fatalf("Error in transaction for %s -> %s, row %d\n",
				netProvider, provider, i)
		}
	}
	tx.Commit()
}

// QueryTime gets the latest timestamp for a specific provider
func QueryTime(dbName string, netProvider string, provider string) time.Time {
	DB := connectDb(dbName)
	defer DB.Close()
	var p Row
	err := DB.QueryRowx(querySQL, netProvider, provider).StructScan(&p)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}
		}
		log.Fatalf("Error finding the timestamp for %q -> %q\n", netProvider, provider)
	}
	return p.Timestamp
}

func fixPercent(s string) float64 {
	// remove the percent sign at the end
	res := strings.ReplaceAll(s, "%", "")

	return fixNumber(res)
}

// all numbers present in the table should have precision 2
func fixNumber(s string) float64 {
	rgx := regexp.MustCompile(`[^\d]`)
	res := rgx.ReplaceAllString(s, "")
	if len(res) < 3 {
		return 0
	}

	idx := len(res) - 2
	res = res[:idx] + "." + res[idx:]

	sNum, err := strconv.ParseFloat(res, 64)
	if err != nil {
		log.Fatalf("Error parsing %q -> %q: %s", s, res, err.Error())
	}

	return sNum
}

// AvgSpeed and MaxSpeed have units at the end (KB, MB or GB)
func fixSpeed(s string) float64 {
	if len(s) < 3 {
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
