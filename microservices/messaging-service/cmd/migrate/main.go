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

	keyspace := os.Getenv("APP_CASSANDRA_KEYSPACE")
	if keyspace == "" {
		log.Fatal("APP_CASSANDRA_KEYSPACE environment variable is not set")
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
	log.Printf("Creating keyspace '%s' if not exists...", keyspace)
	// Properly escape or validate input in a real app, but for now trusting env var
	err = session.Query("CREATE KEYSPACE IF NOT EXISTS " + keyspace + " WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};").Exec()
	if err != nil {
		log.Fatalf("Failed to create keyspace: %v", err)
	}

	// 2. Create Table
	log.Printf("Creating table '%s.users' if not exists...", keyspace)
	err = session.Query("CREATE TABLE IF NOT EXISTS " + keyspace + ".users (" +
		"user_id text PRIMARY KEY, " +
		"email text, " +
		"username text, " +
		"created_at timestamp" +
		");").Exec()
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Migration completed successfully.")
}
