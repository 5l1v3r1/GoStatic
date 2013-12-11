package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const lenPath = len("/view/")

var templates = template.Must(template.ParseFiles("templates/index.html", "templates/edit.html",
	"templates/view.html", "templates/notfound.html",
	"templates/header.html", "templates/footer.html"))
var titleValidator = regexp.MustCompile("^[a-zA-Z0-9]+$")

type Page struct {
	Title string
	Perma string
	Body  []byte
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	tmpl = tmpl + ".html"
	err := templates.ExecuteTemplate(w, tmpl, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func loadPage(title string) (*Page, error) {
	filename := "pages/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func getPageList() ([]string, error) {
	files, err := ioutil.ReadDir("pages")
	if err != nil {
		return nil, err
	}
	pages := make([]string, len(files))
	for i, p := range files {
		pages[i] = strings.Replace(p.Name(), ".txt", "", 1)
	}
	return pages, nil
}

func log(title string, action string, r *http.Request) {
	if r.Header["X-Real-Ip"] != nil {
		fmt.Printf("%s: %s by %s\n", title, action, r.Header["X-Real-Ip"][0])
	} else {
		fmt.Printf("%s: %s by %s\n", title, action, r.RemoteAddr)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log("Index", "viewed", r)
	if len(r.URL.Path) == 1 {
		// asked for root dir.
		pages, err := getPageList()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = templates.ExecuteTemplate(w, "index.html", pages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else { // Didn't ask for root, but we'll give it to them anyways.
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	log("View", title, r)
	p, err := loadPage(title)
	if err != nil {
		fmt.Println(err.Error())
		// Page not found.
		renderTemplate(w, "notfound", &Page{Title: title})
	} else {
		// Page found. Attempt to display.
		renderTemplate(w, "view", p)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[lenPath:]
		if !titleValidator.MatchString(title) {
			http.NotFound(w, r)
			return
		}
		fn(w, r, title)
	}
}

func includeHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Path[1:]
	http.ServeFile(w, r, filename)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/js/", includeHandler)
	http.HandleFunc("/css/", includeHandler)
	http.ListenAndServe(":54545", nil)
}
