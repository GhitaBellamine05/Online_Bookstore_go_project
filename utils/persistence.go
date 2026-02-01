package utils

import (
    "encoding/json"
    "os"
    "online-bookstore/models"
)

type DB struct {
    Books     map[int]models.Book     `json:"books"`
    Authors   map[int]models.Author   `json:"authors"`
    Users     map[int]models.User     `json:"users"` // Replaces Customers
    Orders    map[int]models.Order    `json:"orders"`
    NextIDs   map[string]int          `json:"next_ids"`
}

func LoadDB(filename string) (*DB, error) {
    file, err := os.Open(filename)
    if err != nil {
        if os.IsNotExist(err) {
            return &DB{
                Books:   make(map[int]models.Book),
                Authors: make(map[int]models.Author),
                Users:   make(map[int]models.User), 
                Orders:  make(map[int]models.Order),
                NextIDs: map[string]int{
                    "book":   1,
                    "author": 1,
                    "user":   1, 
                    "order":  1,
                },
            }, nil
        }
        return nil, err
    }
    defer file.Close()

    var db DB
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&db)
    return &db, err
}

func SaveDB(filename string, db *DB) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()
    enc := json.NewEncoder(file)
    enc.SetIndent("", "  ")
    return enc.Encode(db)
}