package services

import (
	"database/sql"
	"encoding/json"
	"go-redis-pg/api/models"
	"go-redis-pg/api/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

var db *sql.DB

type ProductService struct {
	// handlers map[string]ProductInterface
}

// curl --user user1:pass1 127.0.0.1:8000/api/products/list
func (s *ProductService) GetProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM products")
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	products := make([]*models.Product, 0)
	for rows.Next() {
		p := models.Product{}
		err := rows.Scan(&p.Id, &p.Name, &p.Manufacturer)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		}

		// var data []byte
		// data, err = json.Marshal(p)
		// log.Printf(string(data))
		// products = append(products, &p)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	var data []byte
	data, err = json.Marshal(products)
	if err != nil {
		utils.RespondWithError(w, http.StatusNotAcceptable, err.Error())
	}
	log.Printf(string(data))
	utils.RespondWithJSON(w, http.StatusOK, products)
}

// curl --header "Content-Type: application/json" --request POST --data '{"name": "ABC", "manufacturer": "ACME"}' \
// 		--user user1:pass1 127.0.0.1:8000/api/products/new
func (s *ProductService) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	defer r.Body.Close()

	_, err := db.Query("INSERT INTO products (name, manufacturer) VALUES (?, ?)", p.Name, p.Manufacturer)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithMessage(w, http.StatusCreated, "New row added.")
}

// curl --user user1:pass1 127.0.0.1:8000/api/products/10
func (s *ProductService) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := models.Product{Id: id}
	row := db.QueryRow("SELECT name, manufacturer FROM products WHERE id=?", p.Id)
	if err := row.Scan(&p.Name, &p.Manufacturer); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, p)
}

// curl --request PUT --data '{"name": "ABC", "manufacturer": "ACME"}' --user user1:pass1 127.0.0.1:8000/api/products/11
func (s *ProductService) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p models.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	defer r.Body.Close()
	p.Id = id

	_, err = db.Query("UPDATE products SET name=?, manufacturer=? WHERE id=?", p.Name, p.Manufacturer, p.Id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, p)
}

// curl --request DELETE --user user1:pass1 127.0.0.1:8000/api/products/10
func (s *ProductService) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	_, err = db.Query("DELETE FROM products WHERE id=?", id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithMessage(w, http.StatusOK, "Deleted.")
}
