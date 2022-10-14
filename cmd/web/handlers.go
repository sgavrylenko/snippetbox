package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/sgavrylenko/snippetbox/internal/models"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, err)
	}

	data.Snippets = snippets
	app.render(w, http.StatusOK, "home.tmpl", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	data := app.newTemplateData(r)

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	data.Snippet = snippet
	app.render(w, http.StatusOK, "view.tmpl", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, http.StatusOK, "create.tmpl", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	title := r.PostForm.Get("title")
	content := r.PostForm.Get("content")

	expires, err := strconv.Atoi(r.PostForm.Get("expires"))
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	fieldErrors := make(map[string]string)

	if strings.TrimSpace(title) == "" {
		fieldErrors["title"] = "This field can't be blank"
	} else if utf8.RuneCountInString(title) > 100 {
		fieldErrors["title"] = "This field can't be more than 100 characters long"
	}

	if strings.TrimSpace(content) == "" {
		fieldErrors["content"] = "This field can't be blank"
	}

	if expires != 1 && expires != 7 && expires != 365 {
		fieldErrors["expires"] = "This field must be equal 1,7 or 365"
	}

	if len(fieldErrors) > 0 {
		_, _ = fmt.Fprint(w, fieldErrors)
		return
	}

	id, err := app.snippets.Insert(title, content, expires)
	if err != nil {
		app.serverError(w, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
