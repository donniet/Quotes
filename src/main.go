package quotes 

import (
	"net/http"
	"html/template"
	"math/rand"
	
	"appengine"
	"appengine/datastore" 
	"appengine/user"
	
	"errors"
	"time"
)

type Quote struct {
	QuoteId int32
	Quote string
	Date time.Time
}

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/add", add_quote_handler)
	http.HandleFunc("/add/post", add_quote_post_handler)
}

func quote_master_key(c appengine.Context) *datastore.Key {
	return datastore.NewKey(c, "QuoteBook", "default_quotebook", 0, nil)
}

func get_last_quoteid(r *http.Request) int32 {
	c := appengine.NewContext(r)

	q := datastore.NewQuery("Quote").Ancestor(quote_master_key(c)).Order("-QuoteId").Limit(1)
	quotes := make([]Quote, 0, 1)
	
	if _, err := q.GetAll(c, &quotes); err != nil || len(quotes) == 0 {
		return -1
	} else {
		return quotes[0].QuoteId
	}
}

func get_quote(r *http.Request) (string, error) {
	// first get the last quote
	lq := get_last_quoteid(r)
	
	if lq < 0 {
		return "Hello!", nil;
	} else {
		var qid int32 = 0
		
		if lq > 0 {
			qid = rand.Int31n(lq+1);
		}
		
		c := appengine.NewContext(r)
		
		q := datastore.NewQuery("Quote").Ancestor(quote_master_key(c)).Filter("QuoteId =", qid)
		quotes := make([]Quote, 0, 1)
		
		if _, err := q.GetAll(c, &quotes); err != nil {
			return "Error", errors.New("Query Error")
		} else if len(quotes) == 0 {
			return "Hello!", nil;
		} else {
			return quotes[0].Quote, nil
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/home.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		quote, _ := get_quote(r)
	
		t.Execute(w, quote)
	}
}

func add_quote_handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	if !user.IsAdmin(c) {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}

	t, err := template.ParseFiles("templates/add_quote.html")
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		quote, _ := get_quote(r)
	
		t.Execute(w, quote)
	}
}

func add_quote_post_handler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	
	if !user.IsAdmin(c) {
		http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
		return
	}

	quote := r.FormValue("content")
	
	if quote == "" || r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	
	lq := get_last_quoteid(r);
	
	q := Quote{
		QuoteId: lq + 1, 
		Quote: quote, 
		Date: time.Now(),
	}
	
	key := datastore.NewIncompleteKey(c, "Quote", quote_master_key(c))
	_, err := datastore.Put(c, key, &q)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	http.Redirect(w, r, "/", http.StatusFound)
}
