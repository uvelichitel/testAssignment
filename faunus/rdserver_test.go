package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sort"
	"testing"

	mgo "gopkg.in/mgo.v2"
)

var rd Redirector

var good Parameters
var wrong Parameters

func TestMain(m *testing.M) {
	config := Config{
		Credential: mgo.Credential{
			Username: "uvelichitel",
			Password: "***hlhhh**",
		},
		DBUrl:               "localhost",
		DBName:              "dbr",
		CollectionName:      "redi",
		CacheSize:           10,
		CacheClearingPeriod: 100,  //in millisecond
		CacheUpdatePeriod:   1000, //in millisecond
		ServerUrl:           "localhost:6060",
		Ordered:             false, //if parameters order in array matter
	}
	rd = Redirector{Config: config}
	if err := rd.Setup(); err != nil {
		fmt.Println(err)
	}
	defer rd.Teardown()
	go rd.Serve()
	os.Exit(m.Run())
}

func TestGood(t *testing.T) {
	file, err := os.Open("redirects_fixtures.json")
	if err != nil {
		fmt.Println(err)
	}
	dec := json.NewDecoder(file)
	var q []Parameters
	err = dec.Decode(&q)
	if err != nil {
		fmt.Println(err)
	}
	good = q[5]
	sort.Strings(good.P2)
	sort.Strings(good.P3)
	sort.Strings(good.P4)
	sort.Strings(good.P5)
	sort.Strings(good.P6)
	for _, v := range q {
		va := make(url.Values)
		va["p1"] = append(va["p1"], v.P1)
		va["p2"] = v.P2
		va["p3"] = v.P3
		va["p4"] = v.P4
		va["p5"] = v.P5
		va["p6"] = v.P6
		resp, err := http.Get("http://" + rd.ServerUrl + "/?" + va.Encode())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Status ", resp.Status)
		respBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s\n", respBody)
		if resp.Status != "200 OK" {
			t.Fail()
		}
	}
}
func TestWrong(t *testing.T) {
	file, err := os.Open("redirects_fixtures.json")
	if err != nil {
		fmt.Println(err)
	}
	dec := json.NewDecoder(file)
	var q []Parameters
	err = dec.Decode(&q)
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range q {
		va := make(url.Values)
		va["p1"] = append(va["p1"], v.P1)
		va["p2"] = v.P2
		va["p3"] = v.P3
		va["p4"] = v.P4
		va["p5"] = v.P4 //<<here enject misstake
		va["p6"] = v.P6
		resp, err := http.Get("http://" + rd.ServerUrl + "/?" + va.Encode())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Status ", resp.Status)
		respBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s\n", respBody)
		if resp.Status == "200 OK" {
			t.Fail()
		}
	}
}
func BenchmarkCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rd.GetCache(&good)
	}
}
func BenchmarkDB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rd.GetDB(&good)
	}
}