// @title           Feature Flag API
// @version         1.0
// @description     A lightweight feature flag service with rule-based evaluation.
// @host            localhost:3000
// @BasePath        /

package main

import (
	"fmt"
	"net/http"

	db "github.com/agogte/feature-flag-service/api/database"
	_ "github.com/agogte/feature-flag-service/api/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	//seeding one flag to test immediately
	db.Init("./data/flags.db")
	defer db.DB.Close()

	http.HandleFunc("/", router)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	fmt.Println("Flag API running on http://localhost:3000")
	fmt.Println("Swagger UI at http:localhost:3000/swagger/index.html")

	http.ListenAndServe(":3000", nil)
}
