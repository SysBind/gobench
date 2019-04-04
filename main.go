package main

import (
	"time"
	"fmt"
	"log"
	"strconv"
	"database/sql"
	"github.com/remeh/sizedwaitgroup"	
	_ "github.com/go-sql-driver/mysql"
)

type Plan struct {
	Repeat int
	MaxConnections int
}

type Benchmark struct {
	ID int
	Plan
}

type Connection struct {
	Host string
	Port int
	User string
	Pass string
	Database string
	Driver string	
}

type Stats struct {
	Duration time.Duration
}

type Run struct {
	ID int
	Benchmark
	Stats
	Conn Connection
}

func (conn Connection) String() string {
	if len(conn.Database) == 0 {
		conn.Database = "mysql"
	}	
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conn.User, conn.Pass, conn.Host, conn.Port, conn.Database)
}

func (run *Run) Prepare() {
	run.ID = 3

	db, err := sql.Open(run.Conn.Driver, run.Conn.String())
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	
	_, err = db.Exec("CREATE DATABASE bench_"+strconv.Itoa(run.ID) + ";");
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("USE bench_" + strconv.Itoa(run.ID) + ";");
	if err != nil {
		panic(err)
	}

	run.Conn.Database = "bench_" + strconv.Itoa(run.ID)
}

func (run *Run) Exec() {
	start := time.Now()
	db, err := sql.Open(run.Conn.Driver, run.Conn.String())
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("CREATE TABLE x(id int, s1 char(200), s2 char(200), s3 char(200));")
	if err != nil {
		panic(err)
	}

	log.Printf("Spawning %d Inserts (Maximum %d concurrent connections)", run.Repeat, run.MaxConnections)
	// Throttled wait group
	swg := sizedwaitgroup.New(run.MaxConnections)
	for i := 0; i < run.Repeat; i++ {
		swg.Add()
		go func(idx int) {
			defer swg.Done()
			db, err := sql.Open(run.Conn.Driver, run.Conn.String())
			defer db.Close()
			if err != nil {
				log.Fatal(err)
			}
			_, err = db.Exec("INSERT INTO x VALUES(" + strconv.Itoa(idx) + ", 'x', 'y', 'z');")
			if err != nil {
				panic(err)
			}			
		}(i)
	}
	swg.Wait()
	t := time.Now()
	run.Duration = t.Sub(start)
}

func (run *Run) Destroy() {
	db, err := sql.Open(run.Conn.Driver, run.Conn.String())
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DROP DATABASE bench_"+strconv.Itoa(run.ID)+";")
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("GoBench Starting")

	conn := Connection{ Host: "127.0.0.1", Port: 3306, User: "root", Pass: "q1w2e3r4", Driver: "mysql" }
	
	run := Run{ ID: 0, Benchmark: Benchmark{ ID: 0, Plan: Plan { Repeat: 300, MaxConnections: 50 } }, Conn: conn }
	
	run.Prepare()
	defer run.Destroy()

	run.Exec()
	
	fmt.Printf("GoBench Done in %d nanoseconds (%f seconds) \n", run.Duration, float64(run.Duration) / 1000000000)
}
