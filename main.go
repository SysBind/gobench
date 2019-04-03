package main

import (
	"fmt"
	"log"
	"strconv"
	"database/sql"	
	_ "github.com/go-sql-driver/mysql"
)

type Plan struct {
	Repeat int
}

type Benchmark struct {
	ID int
	Plan
}

type Run struct {
	ID int
	Benchmark
}

func (run *Run) Prepare(db *sql.DB) {
	run.ID = 3

	_, err := db.Exec("CREATE DATABASE bench_"+strconv.Itoa(run.ID)+";")
	if err != nil {
		panic(err)
	}
}

func (run *Run) Destroy(db *sql.DB) {
	_, err := db.Exec("DROP DATABASE bench_"+strconv.Itoa(run.ID)+";")
	if err != nil {
		panic(err)
	}	
}

func main() {
	fmt.Println("GoBench Starting")

	run := Run {Repeat: 1000}

	db, err := sql.Open("mysql", "root:q1w2e3r4@tcp(127.0.0.1:3306)/mysql")
	defer db.Close()
	
	if err != nil {
		log.Fatal(err)
	}

	run.Prepare(db)
	defer run.Destroy(db)	
	fmt.Println("GoBench Done")
}
