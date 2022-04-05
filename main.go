package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

// Golang <----DRIVER(pq)-----> DB PostgreSQL

type Book struct {
	isbn   string
	title  string
	author string
	price  float32
}

// database/sql
func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	books := make([]Book, 0)
	for rows.Next() {
		bk := Book{}
		err := rows.Scan(&bk.isbn, &bk.title, &bk.author, &bk.price)
		if err != nil {
			log.Fatal(err)
		}
		books = append(books, bk)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		fmt.Println(book.price, book.title, book.isbn, book.author)
	}

}
