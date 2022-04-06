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

func main() {
	connectionPool, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}

	books, err := GetAllBooksFromDB(connectionPool)
	if err != nil {
		log.Fatal(err)
	}

	for _, book := range books {
		fmt.Println(book.price, book.title, book.isbn, book.author)
	}
}

func connectDB() (*sql.DB, error) {
	connectionPool, err := sql.Open("postgres", "postgres://postgres:postgres@localhost?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return connectionPool, nil
}

func GetAllBooksFromDB(connectionPool *sql.DB) ([]Book, error) {
	rows, err := connectionPool.Query("SELECT * FROM books")
	if err != nil {
		return []Book{}, err
	}
	defer rows.Close()

	books := make([]Book, 0)
	for rows.Next() {
		bk := Book{}
		err := rows.Scan(&bk.isbn, &bk.title, &bk.author, &bk.price)
		if err != nil {
			return []Book{}, err
		}
		books = append(books, bk)
	}
	err = rows.Err()
	if err != nil {
		return []Book{}, err
	}

	return books, nil
}
