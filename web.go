package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"

	"github.com/laurence6/gtd.go/core"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

var errInvalid = errors.New("Invalid!")

var location, _ = time.LoadLocation("Local")

var t *template.Template

func init() {
	var err error
	t, err = template.ParseFiles(
		"templates/default.html",
		"templates/edit.html",
		"templates/form.html",
		"templates/index.html",
	)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func web() {
	http.HandleFunc("/", landing)
	http.HandleFunc("/index", jsonWrapper(index))
	http.HandleFunc("/add", jsonWrapper(add))
	http.HandleFunc("/addSub", jsonWrapper(addSub))
	http.HandleFunc("/edit", jsonWrapper(edit))
	http.HandleFunc("/update", jsonWrapper(updateTask))
	http.HandleFunc("/delete", jsonWrapper(deleteTask))

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))

	go func() {
		log.Fatalln(http.ListenAndServe(":8000", nil))
	}()
}

func landing(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", 404)
		return
	}
	t.ExecuteTemplate(w, "default", "")
}

type responseJSON struct {
	StatusCode int
	Content    string
}

func newResponseJSON() *responseJSON {
	return &responseJSON{}
}

func jsonWrapper(f func(r *http.Request) *responseJSON) http.HandlerFunc {
	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		response := f(r)
		jsonObj, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(jsonObj)
	}
	return handlerFunc
}

func index(r *http.Request) *responseJSON {
	response := &responseJSON{200, ""}
	b := &bytes.Buffer{}
	err := t.ExecuteTemplate(b, "index", defaultIndex)
	if err != nil {
		log.Println(err.Error())
		oops(response, err.Error())
		return response
	}
	response.Content = b.String()
	return response
}

func add(r *http.Request) *responseJSON {
	response := &responseJSON{200, ""}
	b := &bytes.Buffer{}
	err := t.ExecuteTemplate(b, "form", "")
	if err != nil {
		log.Println(err.Error())
		oops(response, err.Error())
		return response
	}
	response.Content = b.String()
	return response
}

func addSub(r *http.Request) *responseJSON {
	response := &responseJSON{200, ""}
	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		oops(response, err.Error())
		return response
	}
	if id != 0 {
		if task := tp.Get(id); task != nil {
			tp.Lock()
			subTask, err := tp.NewSubTask(task)
			tp.Unlock()
			tp.Changed()
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				return response
			}
			redirect(response, "edit?ID="+strconv.FormatInt(subTask.ID, 10))
			return response
		}
	}
	redirect(response, "add")
	return response
}

func edit(r *http.Request) *responseJSON {
	response := &responseJSON{200, ""}
	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		oops(response, err.Error())
		return response
	}
	b := &bytes.Buffer{}
	if id != 0 {
		if task := tp.Get(id); task != nil {
			_ = t.ExecuteTemplate(b, "form", "")
			err = t.ExecuteTemplate(b, "parentsubtask", task)
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				return response
			}
			err = t.ExecuteTemplate(b, "edit", task)
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				return response
			}
			response.Content = b.String()
			return response
		}
	}
	redirect(response, "add")
	return response
}

func updateTask(r *http.Request) *responseJSON {
	response := &responseJSON{200, ""}
	if r.Method == "POST" {
		r.ParseForm()
		id, err := stoI64(r.PostForm.Get("ID"))
		if err != nil {
			log.Println(err.Error())
			oops(response, err.Error())
			return response
		}
		if id == 0 {
			tp.Lock()
			task, err := tp.NewTask()
			tp.Unlock()
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				tp.Delete(task)
				return response
			}
			tp.Lock()
			err = updateTaskFromForm(task, r.PostForm)
			tp.Unlock()
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				tp.Delete(task)
				return response
			}
			tp.Changed()
			redirect(response, "index")
			return response
		} else if task := tp.Get(id); task != nil {
			tp.Lock()
			err := updateTaskFromForm(task, r.PostForm)
			tp.Unlock()
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				return response
			}
			tp.Changed()
			redirect(response, "index")
			return response
		}
	}
	oops(response, errInvalid.Error())
	return response
}

func deleteTask(r *http.Request) *responseJSON {
	response := &responseJSON{200, ""}
	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		oops(response, err.Error())
		return response
	}
	if id != 0 {
		if task := tp.Get(id); task != nil {
			tp.Lock()
			err := tp.Delete(task)
			tp.Unlock()
			if err != nil {
				log.Println(err.Error())
				oops(response, err.Error())
				return response
			}
			tp.Changed()
			redirect(response, "index")
			return response
		}
	}
	oops(response, errInvalid.Error())
	return response
}

func redirect(r *responseJSON, content string) {
	r.StatusCode = 302
	r.Content = content
}

func oops(r *responseJSON, content string) {
	r.StatusCode = 500
	r.Content = "Oops: " + content
}

func stoI64(str string) (int64, error) {
	if str == "" {
		return 0, nil
	}
	i64, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i64, nil
}

// updateTaskFromForm updates Subject, Due, Priority, Next, Notification, Note fields
func updateTaskFromForm(task *gtd.Task, form url.Values) error {
	var err error
	task.Subject = form.Get("Subject")

	noDue := form.Get("NoDue")
	if noDue == "on" {
		task.Due.Set(0)
	} else {
		err := task.Due.ParseDateTimeInLocation(form.Get("DueDate"), form.Get("DueTime"), location)
		if err != nil {
			return err
		}
	}

	task.Priority, err = strconv.Atoi(form.Get("Priority"))
	if err != nil {
		return err
	}

	next := form.Get("Next")
	if next == "on" {
		err := task.Next.ParseDateTimeInLocation(form.Get("NextDate"), form.Get("NextTime"), location)
		if err != nil {
			return err
		}
	} else {
		task.Next.Set(0)
	}

	noNotification := form.Get("NoNotification")
	if noNotification == "on" {
		task.Notification.Set(0)
	} else {
		err := task.Notification.ParseDateTimeInLocation(form.Get("NotificationDate"), form.Get("NotificationTime"), location)
		if err != nil {
			return err
		}
	}

	task.Note = form.Get("Note")

	return nil
}
