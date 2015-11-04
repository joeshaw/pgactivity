package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/lib/pq"
)

type activity struct {
	datname         string
	pid             int
	usename         string
	applicationName string
	clientAddr      sql.NullString
	clientHostname  sql.NullString
	clientPort      sql.NullInt64
	backendStart    pq.NullTime
	xactStart       pq.NullTime
	queryStart      pq.NullTime
	stateChange     pq.NullTime
	waiting         sql.NullBool
	state           sql.NullString
	query           string
}

func (a activity) String() string {
	f := `datname          | %s
pid              | %d
usename          | %s
application_name | %s
client_addr      | %s
client_hostname  | %s
client_port      | %s
backend_start    | %s
xact_start       | %s
query_start      | %s
state_change     | %s
waiting          | %s
state            | %s
query            | %s`

	return fmt.Sprintf(
		f,
		a.datname,
		a.pid,
		a.usename,
		a.applicationName,
		fmtNullString(a.clientAddr),
		fmtNullString(a.clientHostname),
		fmtNullInt64(a.clientPort),
		fmtNullTime(a.backendStart),
		fmtNullTime(a.xactStart),
		fmtNullTime(a.queryStart),
		fmtNullTime(a.stateChange),
		fmtNullBool(a.waiting),
		fmtNullString(a.state),
		stripExtraSpace(a.query),
	)
}

func fmtNullString(n sql.NullString) string {
	if !n.Valid {
		return "[null]"
	}
	return n.String
}

func fmtNullInt64(n sql.NullInt64) string {
	if !n.Valid {
		return "[null]"
	}
	return fmt.Sprint(n.Int64)
}

func fmtNullBool(n sql.NullBool) string {
	if !n.Valid {
		return "[null]"
	}
	return fmt.Sprint(n.Bool)
}

func fmtNullTime(n pq.NullTime) string {
	if !n.Valid {
		return "[null]"
	}
	return n.Time.String()
}

func stripExtraSpace(s string) string {
	prev := rune(' ')
	f := func(r rune) rune {
		defer func() { prev = r }()

		if unicode.IsSpace(r) {
			if unicode.IsSpace(prev) {
				return -1
			}

			return ' '
		}

		return r
	}

	return strings.Map(f, s)
}

func main() {
	var interval int
	flag.IntVar(&interval, "i", 1, "poll interval")
	flag.Parse()

	db, err := sql.Open("postgres", "")
	if err != nil {
		fmt.Println("Unable to connect to DB:", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		fmt.Println("Unable to talk to DB:", err)
		os.Exit(1)
	}

	q := `SELECT
		datname,
		pid,
		usename,
		application_name,
		host(client_addr),
		client_hostname,
		client_port,
		backend_start,
		xact_start,
		query_start,
		state_change,
		waiting,
		state,
		query
	       FROM pg_stat_activity`

	t := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		<-t.C

		rows, err := db.Query(q)
		if err != nil {
			fmt.Printf("Unable to query: %s", err)
			continue
		}

		fmt.Printf("\nQuerying at %s\n", time.Now())

		count := 0
		for rows.Next() {
			var a activity
			err := rows.Scan(
				&a.datname, &a.pid, &a.usename, &a.applicationName,
				&a.clientAddr, &a.clientHostname, &a.clientPort,
				&a.backendStart, &a.xactStart, &a.queryStart, &a.stateChange,
				&a.waiting, &a.state, &a.query,
			)
			if err != nil {
				fmt.Printf("Unable to scan: %s", err)
				rows.Close()
				continue
			}

			count++
			fmt.Printf("-[ RECORD %d ]----\n", count)
			fmt.Printf("%s\n", a)
		}

		rows.Close()

		if err := rows.Err(); err != nil {
			fmt.Printf("Error processing rows: %s", err)
		}
	}
}
