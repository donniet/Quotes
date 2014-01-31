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
	"fmt"
	"strconv"
	"strings"
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
	c := appengine.NewContext(r)
	
	// first get the last quote
	lq := get_last_quoteid(r)
	qid := r.FormValue("qid")
	
	if lq < 0 {
		return "Hello!", nil;
	} else {
		var qidn int32 = 0
		
		if qid != "" && user.IsAdmin(c) {
			var err error = nil 
			var pn int64 = 0
			
			pn, err = strconv.ParseInt(qid, 10, 32)
			qidn = int32(pn)
			
			if err != nil || qidn < 0 || qidn > lq {
				qidn = rand.Int31n(lq+1);				
			}
		} else if lq > 0 {
			qidn = rand.Int31n(lq+1);
		}
		
		c := appengine.NewContext(r)
		
		q := datastore.NewQuery("Quote").Ancestor(quote_master_key(c)).Filter("QuoteId =", qidn)
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
	c := appengine.NewContext(r)
	
	u := user.Current(c);
	url, _ := user.LoginURL(c, "/")
	
	
	if u == nil {
		w.Header().Set("Location", url);
		w.WriteHeader(http.StatusFound);
		return;
	} else if !user.IsAdmin(c) && strings.ToLower(u.Email) != "laurenek@gmail.com" {
		t, _ := template.ParseFiles("templates/error.html");
		
		t.Execute(w, template.HTML(fmt.Sprintf("Not authorized, <a href=\"%s\">click here</a> to login", url)));
		w.WriteHeader(http.StatusUnauthorized);
		return;
	}
	
	t, err := template.ParseFiles("templates/home.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		q, _ := get_quote(r)
	
		data := struct{
			Quote string
			IsAdmin bool
			Forced bool
		}{
			Quote: q,
			IsAdmin: user.IsAdmin(c),
			Forced: r.FormValue("qid") != "",
		}
	
		t.Execute(w, data)
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
	
	http.Redirect(w, r, fmt.Sprintf("/?qid=%d", q.QuoteId), http.StatusFound)
}
