// @title           Feature Flag API
// @version         1.0
// @description     A lightweight feature flag service with rule-based evaluation.
// @host            localhost:3000
// @BasePath        /

package main

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/agogte/feature-flag-service/api/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	//seeding one flag to test immediately
	store["dark-mode"] = Flag{
		Key:         "dark-mode",
		Description: "Switch UI to dark mode",
		IsEnabled:   true,
		Rules: []Rule{
			{Type: "percentage", Rollout: 50},
		},
		CreatedAt: time.Now(),
	}

	http.HandleFunc("/", router)
	http.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	fmt.Println("Flag API running on http://localhost:3000")
	fmt.Println("Swagger UI at http:localhost:3000/swagger/index.html")

	http.ListenAndServe(":3000", nil)
}
