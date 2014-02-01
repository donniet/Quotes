package quotes

import (
	"regexp"
	"net/http"
)


type RegexpRouteHandler interface {
	ServeHTTPRegex(w http.ResponseWriter, r *http.Request, matches []string)
}

type simpleHandlerFuncWrapper struct {
	handler func(http.ResponseWriter, *http.Request)
}
func (h *simpleHandlerFuncWrapper) ServeHTTPRegex(w http.ResponseWriter, r *http.Request, matches []string) {
	h.handler(w, r)
}

type simpleHandlerWrapper struct {
	handler http.Handler
}
func (h *simpleHandlerWrapper) ServeHTTPRegex(w http.ResponseWriter, r *http.Request, matches []string) {
	h.handler.ServeHTTP(w, r)
}

type regexpHandlerFuncWrapper struct {
	handler func(http.ResponseWriter,*http.Request,[]string)
}
func (h *regexpHandlerFuncWrapper) ServeHTTPRegex(w http.ResponseWriter, r *http.Request, matches []string) {
	h.handler(w, r, matches)
}

type regexproute struct {
	pattern *regexp.Regexp
	handler RegexpRouteHandler
}

type RegexpRequestRouter struct {
	routes []regexproute
}

func (r *RegexpRequestRouter) Handler(rexp string, handler http.Handler) error {
	re, err := regexp.Compile(rexp)
	
	if err != nil {
		return err
	}
	
	r.routes = append(r.routes, regexproute{
		re,
		&simpleHandlerWrapper{handler},
	})
	
	return nil
}

func (r *RegexpRequestRouter) HandleFunc(rexp string, handler func(http.ResponseWriter,*http.Request)) error {
	re, err := regexp.Compile(rexp)
	
	if err != nil {
		return err
	}
	
	r.routes = append(r.routes, regexproute{
		re, 
		&simpleHandlerFuncWrapper{handler},
	})
	
	return nil
}
func (r *RegexpRequestRouter) HandleRegexFunc(rexp string, handler func(http.ResponseWriter,*http.Request,[]string)) error {
	re, err := regexp.Compile(rexp);
	
	if err != nil {
		return err
	}
	
	r.routes = append(r.routes, regexproute{
		re,
		&regexpHandlerFuncWrapper{handler},
	})
	
	return nil
}
func (h *RegexpRequestRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range h.routes {
		if matches := route.pattern.FindStringSubmatch(r.URL.Path); matches != nil {
			route.handler.ServeHTTPRegex(w, r, matches)
			return
		}
	}
	
	http.NotFound(w, r)
}