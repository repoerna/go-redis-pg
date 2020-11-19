package api

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

// Server struct
type Server struct {
	Router *mux.Router
	Logger http.Handler
	DB     *sql.DB
	Cache  *redis.Client
}

// var err error

// var ps services.ProductService

// InitializeRoutes services routes
func (s *Server) InitializeRoutes() {
	s.Router.HandleFunc("/api/", func(h1 http.ResponseWriter, h2 *http.Request) {
		fmt.Fprint(h1, "Hello")
	}).Methods("GET")
	s.Router.HandleFunc("/api/products/list", s.GetProducts).Methods("GET")
	s.Router.HandleFunc("/api/products/create", s.CreateProduct).Methods("POST")
	s.Router.HandleFunc("/api/products/{id:[0-9]+}", s.GetProduct).Methods("GET")
	s.Router.HandleFunc("/api/products/{id:[0-9]+}", s.UpdateProduct).Methods("PUT")
	s.Router.HandleFunc("/api/products/{id:[0-9]+}", s.DeleteProduct).Methods("DELETE")
}

// Initialize DB and Redis Client
func (s *Server) Initialize(username, password, host, port, dbName, cacheAddr, cachePass string) {
	dataSource := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbName)

	database, err := sql.Open("postgres", dataSource)
	if err != nil {
		// log.Fatal(err)
		panic(err)
	}
	// defer database.Close()
	s.DB = database
	log.Printf("db connected ...")

	cache := redis.NewClient(&redis.Options{
		Addr:     cacheAddr,
		Password: cachePass,
		DB:       0,
	})

	_, err = cache.Ping().Result()
	if err != nil {
		panic(err)
	}
	s.Cache = cache
	log.Printf("cache db connected ...")

	s.Router = mux.NewRouter()

	s.Logger = handlers.CombinedLoggingHandler(os.Stdout, s.Router)

	s.Router.Use(
		// s.authMiddleware,
		s.cacheMiddleware)

	s.InitializeRoutes()

}

func (s *Server) Run(addr string) {
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	log.Fatal(http.ListenAndServe(":"+viper.GetString("Server.port"),
		handlers.CORS(headersOk, originsOk, methodsOk)(s.Logger)))

	log.Printf("server is ready ...")
}
