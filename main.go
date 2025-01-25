/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"
	"notebox/cmd"
	"os"
)

func init() {
	if _, err := os.Stat("storage"); os.IsNotExist(err) {
		if err := os.Mkdir("storage", os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	cmd.Execute()
}
