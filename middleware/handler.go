package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"stocks-api/models"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // postgres golang driver
)

type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

func createConnection() *sql.DB {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}

	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to Postgres")
	return db
}

func GetStock(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		fmt.Printf("Unable to convert the string into int %v", err)
		return
	}

	stock, err := getStock(int64(id))

	if err != nil {
		fmt.Printf("Unable to get stock %v", err)
		return
	}

	json.NewEncoder(w).Encode(stock)
}

func GetAllStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := getAllStock()

	if err != nil {
		fmt.Printf("Unable to get all stock %v", err)
		return
	}

	json.NewEncoder(w).Encode(stocks)
}

func CreateStock(w http.ResponseWriter, r *http.Request) {

	// Set response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	var stock models.Stock

	// Decode the request body into the stock struct
	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		fmt.Printf("Unable to decode the request body: %v", err)
		// Log the error and send a bad request response
		http.Error(w, fmt.Sprintf("Unable to decode the request body: %v", err), http.StatusBadRequest)
		return
	}

	insertID, err := insertStock(stock)
	if err != nil {
		fmt.Printf("Unable to create stock: %v", err)

		// Log the error and send an internal server error response
		http.Error(w, fmt.Sprintf("Unable to create stock: %v", err), http.StatusInternalServerError)
		return

		// fmt.Printf("Unable to create stock %v", err)
		// return
	}

	res := response{
		ID:      insertID,
		Message: "stock created successfully",
	}

	// json.NewEncoder(w).Encode(res)
	// Encode the response and send it back
	if err := json.NewEncoder(w).Encode(res); err != nil {
		fmt.Printf("Unable to encode the response: %v", err)

		http.Error(w, fmt.Sprintf("Unable to encode the response: %v", err), http.StatusInternalServerError)
		return
	}
}

func UpdateStock(w http.ResponseWriter, r *http.Request) {

	// Set response content type to JSON
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		fmt.Printf("Unable to convert the string into int %v", err)
		http.Error(w, fmt.Sprintf("Unable to convert the string into int %v", err), http.StatusBadRequest)
		return
	}

	var stock models.Stock

	err = json.NewDecoder(r.Body).Decode(&stock)

	if err != nil {
		fmt.Printf("Unable to decode the request body. %v", err)
		http.Error(w, fmt.Sprintf("Unable to decode the request body %v", err), http.StatusBadRequest)
		return

	}

	updatedRows, err := updateStock(int64(id), stock)
	if err != nil {
		fmt.Printf("Unable to update stock %v", err)
		http.Error(w, fmt.Sprintf("Unable to update stock %v", err), http.StatusBadRequest)
		return
	}

	msg := fmt.Sprintf("successfully stock updated, Total rows/stocks effected %v", updatedRows)
	res := response{
		ID:      int64(id),
		Message: msg,
	}
	json.NewEncoder(w).Encode(res)

}

func DeleteStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])

	if err != nil {
		fmt.Printf("Unable to convert the string into int %v", err)
		return
	}

	deletedRows, err := deleteStock(int64(id))
	if err != nil {
		fmt.Printf("Unable to delete stock %v", err)
		return
	}

	msg := fmt.Sprintf(" stock deleted successfully, Total rows/stocks effected %v", deletedRows)
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	json.NewEncoder(w).Encode(res)
}

func insertStock(stock models.Stock) (int64, error) {
	db := createConnection()
	defer db.Close()

	sqlStatement := `INSERT INTO stocks(name, price, company) VALUES ($1, $2, $3) RETURNING stockid`

	var id int64

	err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)

	if err != nil {
		fmt.Printf("Unable to execute the query. %v", err)
		return id, err
	}

	fmt.Printf("Inserted single stock successfully %v", id)

	return id, err
}

func getStock(id int64) (models.Stock, error) {
	db := createConnection()
	defer db.Close()

	var stock models.Stock

	sqlStatement := `SELECT * FROM stocks WHERE stockid = $1`

	row := db.QueryRow(sqlStatement, id)

	err := row.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return stock, nil

	case nil:
		return stock, nil

	default:
		fmt.Printf("Unable to scan the row %v", err)
	}

	fmt.Printf("fetch single stock successfully %v", id)

	return stock, err
}

func getAllStock() ([]models.Stock, error) {
	db := createConnection()
	defer db.Close()

	var stocks []models.Stock
	sqlStatement := `SELECT * FROM stocks`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		fmt.Printf("Unable to execute sql statement %v", err)
	}

	for rows.Next() {
		var stock models.Stock

		err = rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

		if err != nil {
			fmt.Printf("Unable to scan the row %v", err)
		}

		stocks = append(stocks, stock)
	}

	return stocks, err
}

func updateStock(id int64, stock models.Stock) (int64, error) {
	db := createConnection()

	defer db.Close()

	sqlStatement := `UPDATE stocks SET name=$2, price=$3, company=$4 WHERE stockid=$1`

	res, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)

	if err != nil {
		fmt.Printf("Error while executing db query %v", err)
		return 0, err
	}

	rowsEffected, err := res.RowsAffected()

	if err != nil {
		fmt.Printf("Error while checking the affected rows %v", err)
		return rowsEffected, err

	}

	fmt.Printf("Total rows/records affected %v", rowsEffected)

	return rowsEffected, err
}

func deleteStock(id int64) (int64, error) {
	db := createConnection()

	defer db.Close()

	sqlStatement := `DELETE FROM stocks WHERE stockid=$1`

	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		fmt.Printf("Error while executing db query %v", err)
		return 0, err

	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		fmt.Printf("Error while checking the affected rows %v", err)
		return rowsAffected, err

	}

	fmt.Printf("Total rows/records affected %v", rowsAffected)

	return rowsAffected, err

}
