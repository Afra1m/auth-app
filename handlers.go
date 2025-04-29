package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	users       = map[string]string{} // логин -> хэш пароля
	sessions    = map[string]string{} // sessionID -> логин
	sessionsMu  sync.Mutex
	cacheFile   = "cache.json"
	cacheTTL    = time.Minute
	lastCacheAt time.Time
	cacheData   []byte
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, nil)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	if _, exists := users[login]; exists {
		http.Error(w, "Пользователь уже существует", http.StatusBadRequest)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	users[login] = string(hash)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	login := r.FormValue("login")
	password := r.FormValue("password")

	hashedPwd, ok := users[login]
	if !ok || bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(password)) != nil {
		http.Error(w, "Неверный логин или пароль", http.StatusUnauthorized)
		return
	}

	sessionID := generateSessionID()
	sessionsMu.Lock()
	sessions[sessionID] = login
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("templates/profile.html")
	cookie, _ := r.Cookie("session_id")

	sessionsMu.Lock()
	login := sessions[cookie.Value]
	sessionsMu.Unlock()

	tmpl.Execute(w, map[string]string{"Login": login})
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if time.Since(lastCacheAt) < cacheTTL {
		w.Header().Set("Content-Type", "application/json")
		w.Write(cacheData)
		return
	}

	// Сгенерировать новые данные
	data := map[string]interface{}{
		"timestamp": now.Format(time.RFC3339),
		"message":   "Актуальные данные",
	}
	bytes, _ := json.MarshalIndent(data, "", "  ")

	// Кэшировать в файл
	_ = ioutil.WriteFile(cacheFile, bytes, 0644)
	cacheData = bytes
	lastCacheAt = now

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		sessionsMu.Lock()
		_, ok := sessions[cookie.Value]
		sessionsMu.Unlock()

		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func generateSessionID() string {
	return time.Now().Format("20060102150405.000000")
}
