package mdserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var ts *httptest.Server

func TestMain(m *testing.M) {
	ts = httptest.NewServer(http.HandlerFunc(Md5Handler))
	defer ts.Close()
	os.Exit(m.Run())
}

func TestGet(t *testing.T) {
	resp, err := http.Get(ts.URL)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.Status)
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s", respBody)
}

func TestLongPost(t *testing.T) {
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{
"id": 7, 
 "text": "Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку Шла Иаша по шоссе и сосала сушку  "
}`))
	if err != nil {
		fmt.Println(err)
	}
	if resp.Status != "400 Bad Request" {
		t.Errorf("Не возвращает ошибку на длинный запрос")
	}
}

func TestPost(t *testing.T) {
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{
"id": 7, 
 "text": "Шла Иаша по шоссе и сосала сушку"
}`))
	if err != nil {
		fmt.Println(err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	if string(respBody) != `{"md5хеш":"f8936ddf7dd3de1a360d7297a299056c"}` {
		t.Errorf("Неправильный хаш %s , правильный f8936ddf7dd3de1a360d7297a299056c ", respBody)
	}
}