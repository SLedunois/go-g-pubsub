package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	pq "github.com/lib/pq"
)

func waitForNotification(l *pq.Listener) {
	select {
	case n := <-l.Notify:
		fmt.Println("received notification")
		fmt.Printf("pid=%d channel=%s content=%s", n.BePid, n.Channel, n.Extra)
	case <-time.After(90 * time.Second):
		go l.Ping()
		// Check if there's more work available, just in case it takes
		// a while for the Listener to notice connection loss and
		// reconnect.
		fmt.Println("received no work for 90 seconds, checking for new work")
	}
}

func main() {
	connStr := "user=postgres password=password dbname=example sslmode=disable"
	_, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	minReconn := 10 * time.Second
	maxReconn := time.Minute
	listener := pq.NewListener(connStr, minReconn, maxReconn, reportProblem)
	err = listener.Listen("pubsubalert")
	if err != nil {
		panic(err)
	}

	fmt.Println("entering main loop")
	for {
		// process all available work before waiting for notifications
		waitForNotification(listener)
	}
}

//SELECT pg_notify('pubsubalert', 'alert content');
