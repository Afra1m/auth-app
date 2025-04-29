package main

import (
	"log"
	"net/http"
)

func main() {

	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Serving static file: %s", r.URL.Path)
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP(w, r)
	})

	// Роуты
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/profile", authMiddleware(profileHandler))
	http.HandleFunc("/data", dataHandler)

	log.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Ошибка сервера: ", err)
	}
}
