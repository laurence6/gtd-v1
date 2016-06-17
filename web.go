package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
		"templates/login.html",
	)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func web() {
	http.HandleFunc("/", landing)

	http.HandleFunc("/auth", jsonHandlerWrapper(index))

	http.HandleFunc("/index", jsonHandlerWrapper(index))
	http.HandleFunc("/add", jsonHandlerWrapper(addTask))
	http.HandleFunc("/addSub", jsonHandlerWrapper(addSubTask))
	http.HandleFunc("/edit", jsonHandlerWrapper(editTask))
	http.HandleFunc("/done", jsonHandlerWrapper(doneTask))
	http.HandleFunc("/delete", jsonHandlerWrapper(deleteTask))
	http.HandleFunc("/update", jsonHandlerWrapper(updateTask))

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("static/css"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("static/js"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("static/fonts"))))

	go func() {
		addr, ok := conf["web_listen_addr"].(string)
		if !ok {
			log.Panic("Cannot get web server listen addr 'web_listen_addr'")
		}
		log.Panic(http.ListenAndServe(addr, nil))
	}()
}

func landing(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		httpError(w, 404, "")
		return
	}
	t.ExecuteTemplate(w, "default", "")
}

const (
	jsonStatusOK            = "OK"
	jsonStatusAuthenticated = "Authenticated"
	jsonStatusRedirect      = "Redirect"
	jsonStatusError         = "Error"
)

type responseJSON struct {
	Status  string
	Content string
}

func newResponseJSON() *responseJSON {
	return &responseJSON{jsonStatusOK, ""}
}

func jsonHandlerWrapper(f func(http.ResponseWriter, *http.Request) *responseJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			httpError(w, 405, "POST")
			return
		}

		var response *responseJSON

		response = auth(w, r)
		if response == nil {
			response = f(w, r)
		}

		var rJSON []byte
		if response != nil {
			var err error
			rJSON, err = json.Marshal(response)
			if err != nil {
				response = newResponseJSON()
				jsonError(response, err.Error())
				rJSON, _ = json.Marshal(response)
			}
		} else {
			response = newResponseJSON()
			jsonError(response, "Empty response")
			rJSON, _ = json.Marshal(response)
		}
		w.Write(rJSON)
	}
}

func auth(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()

	if r.RequestURI == "/auth" {
		r.ParseForm()

		password := r.PostFormValue("Password")
		ok, err := CheckPassword(password)
		if err != nil {
			log.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
		if ok {
			token := NewToken()
			expires := 60 * 60 * 24 * 30
			err = SetToken(token, expires)
			if err != nil {
				log.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, MaxAge: expires})
			response.Status = jsonStatusAuthenticated
			return response
		}

		jsonError(response, "Password incorrect")
		return response
	}

	if cookie, err := r.Cookie("token"); err == nil {
		if ok, err := CheckToken(cookie.Value); err == nil && ok {
			return nil
		} else if err != nil {
			log.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
	}

	response = login(w, r)
	return response
}

func login(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	_ = t.ExecuteTemplate(b, "login", "")

	response.Content = b.String()
	return response
}

func index(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	err := t.ExecuteTemplate(b, "index", defaultIndex)
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	response.Content = b.String()
	return response
}

func addTask(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	err := t.ExecuteTemplate(b, "form", "")
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	response.Content = b.String()
	return response
}

func addSubTask(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()

	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id != 0 {
		tpRW.RLock()
		task := tp.Get(id)
		tpRW.RUnlock()
		if task != nil {
			tpRW.Lock()
			subTask, err := tp.NewSubTask(task)
			tpRW.Unlock()
			tp.Changed()
			if err != nil {
				log.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			jsonRedirect(response, "/edit?ID="+strconv.FormatInt(subTask.ID, 10))
			return response
		}
	}

	jsonRedirect(response, "/add")
	return response
}

func editTask(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	r.ParseForm()

	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id != 0 {
		tpRW.RLock()
		task := tp.Get(id)
		tpRW.RUnlock()
		if task != nil {
			_ = t.ExecuteTemplate(b, "form", "")
			err = t.ExecuteTemplate(b, "parentsubtask", task)
			if err != nil {
				log.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			err = t.ExecuteTemplate(b, "edit", task)
			if err != nil {
				log.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			response.Content = b.String()
			return response
		}
	}

	jsonRedirect(response, "/add")
	return response
}

func doneTask(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()

	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id != 0 {
		tpRW.RLock()
		task := tp.Get(id)
		tpRW.RUnlock()
		if task != nil {
			tpRW.Lock()
			err := tp.Done(task)
			tpRW.Unlock()
			if err != nil {
				log.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			tp.Changed()
			jsonRedirect(response, "/index")
			return response
		}
	}

	jsonError(response, errInvalid.Error())
	return response
}

func deleteTask(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()

	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id != 0 {
		tpRW.RLock()
		task := tp.Get(id)
		tpRW.RUnlock()
		if task != nil {
			tpRW.Lock()
			err := tp.Delete(task)
			tpRW.Unlock()
			if err != nil {
				log.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			tp.Changed()
			jsonRedirect(response, "/index")
			return response
		}
	}

	jsonError(response, errInvalid.Error())
	return response
}

func updateTask(w http.ResponseWriter, r *http.Request) *responseJSON {
	response := newResponseJSON()
	var task *gtd.Task

	r.ParseForm()

	id, err := stoI64(r.PostFormValue("ID"))
	if err != nil {
		log.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id == 0 {
		tpRW.Lock()
		task, err = tp.NewTask()
		tpRW.Unlock()
		if err != nil {
			tp.Delete(task)
			log.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
		tpRW.Lock()
		err = updateTaskFromForm(task, r.PostForm)
		tpRW.Unlock()
		if err != nil {
			tp.Delete(task)
			log.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
		tp.Changed()
		jsonRedirect(response, "/index")
		return response
	}

	tpRW.RLock()
	task = tp.Get(id)
	tpRW.RUnlock()
	if task != nil {
		tpRW.Lock()
		err := updateTaskFromForm(task, r.PostForm)
		tpRW.Unlock()
		if err != nil {
			log.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
		tp.Changed()
		jsonRedirect(response, "/index")
		return response
	}

	jsonError(response, errInvalid.Error())
	return response
}

var httpErrorMessage = map[int]string{
	404: "404 NotFound",
	405: "405 MethodNotAllowed",
	500: "500 InternalServerError",
}

func httpError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	codeMessage, ok := httpErrorMessage[code]
	if !ok {
		codeMessage = strconv.Itoa(code)
	}
	fmt.Fprintf(w, "<html><body><center><h1>%s</h1></center><p>%s</p></body></html>", codeMessage, message)
}

func jsonRedirect(r *responseJSON, content string) {
	r.Status = jsonStatusRedirect
	r.Content = content
}

func jsonError(r *responseJSON, content string) {
	r.Status = jsonStatusError
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

// updateTaskFromForm updates Subject, Due, Priority, Notification, Next, Note fields.
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

	noNotification := form.Get("NoNotification")
	if noNotification == "on" {
		task.Notification.Set(0)
	} else {
		err := task.Notification.ParseDateTimeInLocation(form.Get("NotificationDate"), form.Get("NotificationTime"), location)
		if err != nil {
			return err
		}
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

	task.Note = form.Get("Note")

	return nil
}
