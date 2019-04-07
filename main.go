package main

import (
	"time"
	"os"
	"fmt"
	"log"
	"strconv"
	"database/sql"
	"github.com/remeh/sizedwaitgroup"	
	_ "github.com/go-sql-driver/mysql"
)

type Plan struct {
	Repeat int64
	Delay time.Duration
	MaxConnections int64
}

type Benchmark struct {
	ID int
	Plan
}

type Connection struct {
	Host string
	Port int64
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


	log.Printf("Creating Database bench_"+strconv.Itoa(run.ID))
	_, err = db.Exec("CREATE DATABASE bench_"+strconv.Itoa(run.ID) + ";");
	if err != nil {
		log.Printf("Dropping already existed Database bench_"+strconv.Itoa(run.ID))
		_, err = db.Exec("DROP DATABASE bench_"+strconv.Itoa(run.ID) + ";");
		log.Printf("Creating Database bench_"+strconv.Itoa(run.ID))
		_, err = db.Exec("CREATE DATABASE bench_"+strconv.Itoa(run.ID) + ";");
		if err != nil {		
			panic(err)
		}
	}
	_, err = db.Exec("USE bench_" + strconv.Itoa(run.ID) + ";");
	if err != nil {
		panic(err)
	}

	run.Conn.Database = "bench_" + strconv.Itoa(run.ID)
}

func (run *Run) Exec() {
	db, err := sql.Open(run.Conn.Driver, run.Conn.String())
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}

	for {
		_, err = db.Exec("CREATE TABLE x(id int, s1 char(200), s2 char(200), s3 char(200), CONSTRAINT PRIMARY KEY (id));")
		if err != nil {
			panic(err)
		}
		
		start := time.Now()
		log.Printf("Spawning %d Inserts (Maximum %d concurrent connections)", run.Repeat, run.MaxConnections)
		// Throttled wait group
		swg := sizedwaitgroup.New(int(run.MaxConnections))
		for i := 0; i < int(run.Repeat); i++ {
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
		log.Printf("Run done in %d nanoseconds (%f seconds) \n", run.Duration, float64(run.Duration) / float64(time.Second))
		log.Printf("Sleeping for %d seconds before next run \n", run.Delay / time.Second)
		time.Sleep(run.Delay)
		_, err = db.Exec("DROP TABLE x;")
	}
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

func EnvOrDefault(key string, defaultStr string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultStr
	}
	return val
}

func EnvOrDefaultInt(key string, defaultInt int64) int64 {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultInt		
	}
	retval, err := strconv.ParseInt(val, 10, 0)
	if err != nil {
		log.Panicf("Could not parse value '%s' as int for environment variable %s", val, key)
	}			
	return retval	
}

func main() {
	fmt.Println("GoBench Starting")


	host := EnvOrDefault("DB_HOST", "127.0.0.1")
	port := EnvOrDefaultInt("DB_PORT", 3306)
	user := EnvOrDefault("DB_USER", "root")
	pass := EnvOrDefault("DB_PASS", "q1w2e3r4")
	repeat := EnvOrDefaultInt("REPEAT", 1000)
	delay := time.Duration(EnvOrDefaultInt("DELAY", 300)) * time.Second
	maxConn := EnvOrDefaultInt("MAX_CONNECTIONS", 70)

	conn := Connection{ Host: host, Port: port, User: user, Pass: pass, Driver: "mysql" }

	log.Printf("Connection String is %s", conn.String())
	
	run := Run{ ID: 0, Benchmark: Benchmark{ ID: 0, Plan: Plan { Repeat: repeat, Delay: delay, MaxConnections: maxConn } }, Conn: conn }
	
	run.Prepare()
	defer run.Destroy()

	run.Exec()
	
	log.Printf("GoBench Done in %d nanoseconds (%f seconds) \n", run.Duration, float64(run.Duration) / 1000000000)
}
