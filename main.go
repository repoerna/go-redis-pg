package main

import (
	"fmt"
	"log"

	"go-redis-pg/api"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./config/")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	log.Printf("starting api ...")
	s := api.Server{}
	log.Printf("reading config file ...")
	s.Initialize(viper.GetString("DB.username"),
		viper.GetString("DB.password"),
		viper.GetString("DB.host"),
		viper.GetString("DB.port"),
		viper.GetString("DB.db_name"),
		viper.GetString("Cache.addr"),
		viper.GetString("Cache.password"))

	s.Run(viper.GetString("Server.port"))
	log.Printf("api is running ...")
}
