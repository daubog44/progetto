package main

import (
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
)

func main() {
	host := os.Getenv("CASSANDRA_HOST")
	if host == "" {
		host = "cassandra"
	}

	log.Println("Waiting for Cassandra to be ready...")
	// Simple retry loop
	var session *gocql.Session
	var err error

	// Initial connection to create keyspace (system keyspace)
	for i := 0; i < 30; i++ {
		cluster := gocql.NewCluster(host)
		cluster.Consistency = gocql.Quorum
		cluster.ProtoVersion = 4
		cluster.ConnectTimeout = 10 * time.Second
		session, err = cluster.CreateSession()
		if err == nil {
			break
		}
		log.Printf("Cassandra unavailable: %v. Retrying...", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to Cassandra after retries: %v", err)
	}
	defer session.Close()

	log.Println("Connected. executing migration scripts...")

	// 1. Create Keyspace
	err = session.Query(`
		CREATE KEYSPACE IF NOT EXISTS messaging 
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
	`).Exec()
	if err != nil {
		log.Fatalf("Failed to create keyspace: %v", err)
	}

	// 2. Create Table
	err = session.Query(`
		CREATE TABLE IF NOT EXISTS messaging.users (
			user_id text PRIMARY KEY,
			email text,
			username text,
			created_at timestamp
		);
	`).Exec()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Migration completed successfully.")
}
