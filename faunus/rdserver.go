package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Parameters struct {
	P1 string   `json:"p1" bson:"p1"`
	P2 []string `json:"p2" bson:"p2"`
	P3 []string `json:"p3" bson:"p3"`
	P4 []string `json:"p4" bson:"p4"`
	P5 []string `json:"p5" bson:"p5"`
	P6 []string `json:"p6" bson:"p6"`
}

type Redirection struct {
	Point      `bson:",inline"`
	Parameters `bson:",inline"`
}

//из коробки парсит в map, а я хотел struct, struct - comparable может быть ключом
func ParseQuery(query string, ordered bool) (Parameters, error) {
	var p Parameters
	var err error
	var key, value string
	for query != "" {
		key = query
		if i := strings.IndexAny(key, "&;"); i >= 0 {
			key, query = key[:i], key[i+1:]
		} else {
			query = ""
		}
		if key == "" {
			continue
		}
		value = ""
		if key[2] != byte(61) {
			continue
		}
		key, value = key[:2], key[3:]

		value, err = url.QueryUnescape(value)
		if err != nil {
			continue
		}
		switch key {
		case "p1":
			p.P1 = value
		case "p2":
			p.P2 = append(p.P2, value)
		case "p3":
			p.P3 = append(p.P3, value)
		case "p4":
			p.P4 = append(p.P4, value)
		case "p5":
			p.P5 = append(p.P5, value)
		case "p6":
			p.P6 = append(p.P6, value)
		default:
		}
	}
	if !ordered {
		sort.Strings(p.P2)
		sort.Strings(p.P3)
		sort.Strings(p.P4)
		sort.Strings(p.P5)
		sort.Strings(p.P6)
	}
	return p, err
}

type FoldedParameters struct {
	P1 string
	P2 string
	P3 string
	P4 string
	P5 string
	P6 string
}

// слайсы сливаются в ключи, уникальность сущностей снижается, но допустимо по моему
func FoldParameters(p *Parameters) FoldedParameters {
	var fp FoldedParameters
	fp.P2 = strings.Join(p.P2, " ")
	fp.P3 = strings.Join(p.P3, " ")
	fp.P4 = strings.Join(p.P4, " ")
	fp.P5 = strings.Join(p.P5, " ")
	fp.P6 = strings.Join(p.P6, " ")
	return fp
}

type Point struct {
	Path string `json:"url" bson:"url"`
	Id   int    `json:"id" bson:"id"`
}

//намвный кеш, защищен для конкурентной записи
type Cache struct {
	M map[FoldedParameters]Point
	L sync.RWMutex
}

type Config struct {
	mgo.Credential
	DBUrl               string //[mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	DBName              string
	CollectionName      string
	CacheSize           int    //в записях
	CacheClearingPeriod int64  //в миллисекундах
	CacheUpdatePeriod   int64  //в миллисекундах
	ServerUrl           string //host:port
	Ordered             bool   //if parameters order in array matter
}

//центральная сущность, мониторит соединения и состояние
type Redirector struct {
	Cache
	session  *mgo.Session
	listener net.Listener
	Config
}

func (rd *Redirector) GetRedirect(p *Parameters) (string, bool) {
	r, ok := rd.GetCache(p)
	if !ok {
		r, ok = rd.GetDB(p)
	}
	return r, ok
}

func (rd *Redirector) GetCache(p *Parameters) (string, bool) {
	fp := FoldParameters(p)
	rd.L.RLock()
	res, ok := rd.M[fp]
	rd.L.RUnlock()
	return res.Path, ok
}

func (rd *Redirector) GetDB(ps *Parameters) (string, bool) {
	var p Point
	if err := rd.session.DB(rd.DBName).C(rd.CollectionName).Find(ps).Sort("-id").Select(bson.M{"id": 1, "url": 1}).One(&p); err == nil {
		fp := FoldParameters(ps)
		rd.L.Lock()
		rd.M[fp] = p
		rd.L.Unlock()
		return p.Path, true
	}
	return "", false
}

// если в ℂonfig выставлен Ordered false элементы массивов будут отсортированы в mongo силами mongo
func (rd *Redirector) UpdateCache() {
	temp := make(map[FoldedParameters]Point)
	if !rd.Ordered {
		rd.session.DB(rd.DBName).C(rd.CollectionName).UpdateAll(nil, bson.M{"$push": bson.M{"p2": bson.M{"$each": []string{}, "$sort": 1}}})
		rd.session.DB(rd.DBName).C(rd.CollectionName).UpdateAll(nil, bson.M{"$push": bson.M{"p3": bson.M{"$each": []string{}, "$sort": 1}}})
		rd.session.DB(rd.DBName).C(rd.CollectionName).UpdateAll(nil, bson.M{"$push": bson.M{"p4": bson.M{"$each": []string{}, "$sort": 1}}})
		rd.session.DB(rd.DBName).C(rd.CollectionName).UpdateAll(nil, bson.M{"$push": bson.M{"p5": bson.M{"$each": []string{}, "$sort": 1}}})
		rd.session.DB(rd.DBName).C(rd.CollectionName).UpdateAll(nil, bson.M{"$push": bson.M{"p6": bson.M{"$each": []string{}, "$sort": 1}}})
	}
	iter := rd.session.DB(rd.DBName).C(rd.CollectionName).Find(nil).Iter()
	for counter, doc := 0, new(Redirection); iter.Next(doc) && (counter < rd.Config.CacheSize); {
		fp := FoldParameters(&doc.Parameters)
		if v, ok := temp[fp]; !ok || v.Id < doc.Id {
			temp[fp] = Point{doc.Path, doc.Id}
			counter++
		}
	}
	rd.L.Lock()
	rd.M = temp
	rd.L.Unlock()
}
func (rd *Redirector) CompactCache() {
	if spare := (len(rd.M) - rd.CacheSize); spare > 0 {
		rd.L.Lock()
		for k := range rd.M {
			if spare == 0 {
				break
			}
			delete(rd.M, k)
			spare--
		}
		rd.L.Unlock()
	}
}

func (rd *Redirector) MonitorDB() {
	compact := time.NewTicker(time.Duration(rd.CacheClearingPeriod))
	update := time.NewTicker(time.Duration(rd.CacheUpdatePeriod))
	for {
		select {
		case <-compact.C:
			rd.CompactCache()
		case <-update.C:
			rd.UpdateCache()
		}
	}
}

func (rd *Redirector) HandleHTTP(w http.ResponseWriter, r *http.Request) {
	if redirection, err := ParseQuery(r.URL.RawQuery, rd.Ordered); err == nil {
		if urlStr, ok := rd.GetRedirect(&redirection); ok {
			http.Redirect(w, r, urlStr, http.StatusSeeOther)
		} else {
			http.Error(w, "Нет редиректа", http.StatusNotFound)
		}
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

func (rd *Redirector) Setup() error {
	var err error
	if rd.session, err = mgo.Dial(rd.DBUrl); err != nil {
		return err
	}
	if err = rd.session.Login(&rd.Credential); err != nil {
		return err
	}
	if rd.listener, err = net.Listen("tcp", rd.ServerUrl); err != nil {
		return err
	}

	rd.Cache = Cache{M: make(map[FoldedParameters]Point)}
	return nil

}

func (rd *Redirector) Serve() error {
	rd.UpdateCache()
	go rd.MonitorDB()
	return http.Serve(rd.listener, http.HandlerFunc(rd.HandleHTTP))
}

//для зачистки, а можно для экстренного торможения
func (rd *Redirector) Teardown() {
	rd.session.LogoutAll()
	rd.session.Close()
	rd.listener.Close()
}

func main() {
	config := Config{
		Credential: mgo.Credential{
			Username: "uvelichitel",
			Password: "**gigiii**",
		},
		DBUrl:               "localhost/somedb",
		DBName:              "somedb",
		CollectionName:      "redirects",
		CacheSize:           50,
		CacheClearingPeriod: 100,  //in millisecond
		CacheUpdatePeriod:   1000, //in millisecond
		ServerUrl:           "localhost:6060",
		Ordered:             false, //if parameters order in array matter
	}
	rd := Redirector{Config: config}
	if err := rd.Setup(); err != nil {
		fmt.Println(err)
	}
	defer rd.Teardown()
	log.Fatal(rd.Serve())
}