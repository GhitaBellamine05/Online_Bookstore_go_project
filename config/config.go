package config

import "os"

type StoreType string
const (
	JSONStore StoreType = "json"
	SQLStore  StoreType = "sql"
)
func GetStoreType() StoreType {
	storeType := os.Getenv("STORE_TYPE")
	if storeType == "" {
		return JSONStore
	}
	switch storeType {
	case "sql":
		return SQLStore
	default:
		return JSONStore
	}
}