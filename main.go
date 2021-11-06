package main

import (
	cethacea "github.com/elek/cethacea/pkg"
	"github.com/spf13/viper"
	"log"
)

func main() {

	viper.SetConfigName(".ceth")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("CETH")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("%++v", err)
		}
	}

	err = cethacea.RootCmd.Execute()
	if err != nil {
		log.Fatalf("%++v", err)
	}
}
