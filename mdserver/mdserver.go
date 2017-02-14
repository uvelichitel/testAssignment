package mdserver

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"unicode/utf8"
)

func Md5Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Wrong method", http.StatusBadRequest)
		return
	}
	var data struct {
		Id   int    `json:"id"`
		Text string `json:"text"`
	}
	if json.NewDecoder(r.Body).Decode(&data) != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	switch {
	case data.Text == "":
		http.Error(w, "No payload supplied", http.StatusBadRequest)
		return
	case utf8.RuneCountInString(data.Text) > 100:
		http.Error(w, "Too much data", http.StatusBadRequest)
		return
	case data.Id < 1:
		http.Error(w, "Negative Id", http.StatusBadRequest)
		return
	default:
	}
	if json, err := json.Marshal(struct {
		Md5hash string `json:"md5хеш"`
	}{fmt.Sprintf("%x", md5.Sum([]byte(strconv.Itoa(data.Id)+data.Text+strconv.Itoa(data.Id%2))))}); err != nil {
		http.Error(w, "Who know...", http.StatusInternalServerError)
		return
	} else if _, err = w.Write(json); err != nil {
		log.Println(err)
	}
}