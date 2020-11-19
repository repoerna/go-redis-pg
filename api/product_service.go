package api

import (
	"encoding/json"
	"go-redis-pg/api/models"
	"go-redis-pg/api/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// curl --user user1:pass1 127.0.0.1:8000/api/products/list
func (s *Server) GetProducts(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT * FROM products")
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

		products = append(products, &p)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	// data = json.NewEncoder(w).Encode(products)
	utils.RespondWithJSON(w, http.StatusOK, products)
}

// curl --header "Content-Type: application/json" --request POST --data '{"name": "ABC", "manufacturer": "ACME"}' \
// 		--user user1:pass1 127.0.0.1:8000/api/products/new
func (s *Server) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p *models.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	defer r.Body.Close()

	_, err := s.DB.Query("INSERT INTO products (name, manufacturer) VALUES ($1, $2)", p.Name, p.Manufacturer)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithMessage(w, http.StatusCreated, "New row added.")
}

// curl --user user1:pass1 127.0.0.1:8000/api/products/10
func (s *Server) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	p := &models.Product{Id: id}
	row := s.DB.QueryRow("SELECT name, manufacturer FROM products WHERE id=$1", p.Id)
	if err := row.Scan(&p.Name, &p.Manufacturer); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, p)
}

// curl --request PUT --data '{"name": "ABC", "manufacturer": "ACME"}' --user user1:pass1 127.0.0.1:8000/api/products/11
func (s *Server) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var p *models.Product
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&p); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid payload")
		return
	}
	defer r.Body.Close()
	p.Id = id

	_, err = s.DB.Query("UPDATE products SET name=$1, manufacturer=$2 WHERE id=$3", p.Name, p.Manufacturer, p.Id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	content, _ := json.Marshal(p)
	err = s.Cache.Set(r.RequestURI, content, 10*time.Minute).Err()
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	utils.RespondWithJSON(w, http.StatusOK, p)
}

// curl --request DELETE --user user1:pass1 127.0.0.1:8000/api/products/10
func (s *Server) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	_, err = s.DB.Query("DELETE FROM products WHERE id=$1", id)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondWithMessage(w, http.StatusOK, "Deleted.")
}
