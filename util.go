package main

import (
	"io/ioutil"
	"encoding/json"
	"log"
	"os"
)

func LoadConfigs(fileName string) (Configs,error) {
	file, e := ioutil.ReadFile(fileName)
	if e != nil {
		log.Printf("Load config file error: %v\n", e)
		os.Exit(1)
	}
	
	var config Configs
	err := json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Config load error:%v \n",err)
		return config,err
	}
	return config,nil
}

