package main

import (
	"encoding/json"

	"fmt"

	"github.com/labstack/gommon/log"
)

type greeting struct {
	ID      string `json:"id"`
	Message string `json:"msg"`
}

type baseResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func main() {

	jsonText := `
	{
		  "type":"greeting"
		, "data":
		{
		     "id":"1234"
		   , "msg":"Hello World"
		   , "ignore":true
		}
	}
	`

	r := &baseResponse{}
	if err := json.Unmarshal([]byte(jsonText), &r); err != nil {
		log.Fatal("Could not parse json: ", err)
	}

	fmt.Println(r)
	switch r.Type {
	case "greeting":
		g := &greeting{}

		json.Unmarshal(r.Data, &g)
		fmt.Println(g)
	}

}
