package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"

	"github.com/laurence6/gtd.go/model"
)

const (
	dateLayout  = "2006-01-02"
	timeLayout  = "15:04"
	tokenExpire = 60 * 60 * 24 * 30 // 30 Days
)

var location, _ = time.LoadLocation("Local")

var t *template.Template

// Task columns
const (
	taskUserID  = "user_id"
	taskUserIDL = "task.user_id"

	taskID     = "id"
	taskIDL    = "task.id"
	taskStart  = "start"
	taskStartL = "task.start"

	taskSubject       = "subject"
	taskSubjectL      = "task.subject"
	taskDue           = "due"
	taskDueL          = "task.due"
	taskPriority      = "priority"
	taskPriorityL     = "task.priority"
	taskNotification  = "notification"
	taskNotificationL = "task.notification"
	taskNext          = "next"
	taskNextL         = "task.next"
	taskNote          = "note"
	taskNoteL         = "task.note"

	taskTags = "Tags"

	taskParentTaskID  = "parent_task_id"
	taskParentTaskIDL = "task.parent_task_id"
	taskParentTask    = "ParentTask"
	taskSubTasks      = "SubTasks"
)

func init() {
	var err error
	t, err = template.ParseFiles(
		"templates/default.html",
		"templates/edit.html",
		"templates/form.html",
		"templates/home.html",
		"templates/login.html",
	)
	if err != nil {
		logger.Fatalln(err.Error())
	}
}

func web() {
	http.HandleFunc("/", landing)

	http.HandleFunc("/auth", jsonHandlerWrapper(home))

	http.HandleFunc("/home", jsonHandlerWrapper(home))
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
		logger.Panic(http.ListenAndServe(conf.WebListenAddr, nil))
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

func jsonHandlerWrapper(f func(http.ResponseWriter, *http.Request, Flash) *responseJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			httpError(w, 405, "POST")
			return
		}

		var response *responseJSON

		flash := Flash{}
		response = auth(w, r, flash)
		if response == nil {
			response = f(w, r, flash)
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

func auth(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	if r.RequestURI == "/auth" {
		r.ParseForm()
		userID := r.PostFormValue("UserID")
		password := r.PostFormValue("Password")

		var err error
		userID, err = CheckPassword(userID, EncPassword(password))
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}

		if userID != "" {
			token := NewToken()
			err = SetToken(userID, token, tokenExpire)
			if err != nil {
				logger.Println(err.Error())
				jsonError(response, err.Error())
				return response
			}
			http.SetCookie(w, &http.Cookie{Name: "token", Value: token, MaxAge: tokenExpire})
			response.Status = jsonStatusAuthenticated
			return response
		}

		jsonError(response, "Password incorrect")
		return response
	}

	if cookie, err := r.Cookie("token"); err == nil {
		if userID, err := CheckToken(cookie.Value); err == nil && userID != "" {
			flash["UserID"] = userID
			return nil
		} else if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
	}

	b := &bytes.Buffer{}

	_ = t.ExecuteTemplate(b, "login", "")

	response.Content = b.String()
	return response
}

func home(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	tasks, err := model.GetTasksByUserID(flash["UserID"], taskIDL, taskSubjectL, taskDueL, taskPriorityL, taskNoteL, taskSubTasks)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	SortByDefault(tasks)

	err = t.ExecuteTemplate(b, "home", tasks)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	response.Content = b.String()
	return response
}

func addTask(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	_ = t.ExecuteTemplate(b, "form", "")

	response.Content = b.String()
	return response
}

func addSubTask(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	_ = t.ExecuteTemplate(b, "form", "")
	err = t.ExecuteTemplate(b, "edit", model.Task{ParentTaskID: id})
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	response.Content = b.String()
	return response
}

func editTask(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()
	b := &bytes.Buffer{}

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	task, err := model.GetTask(flash["UserID"], id, taskIDL, taskSubjectL, taskDueL, taskPriorityL, taskNotificationL, taskNextL, taskNoteL, taskParentTaskIDL, taskParentTask, taskSubTasks)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	_ = t.ExecuteTemplate(b, "form", "")
	err = t.ExecuteTemplate(b, "parentsubtask", task)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	err = t.ExecuteTemplate(b, "edit", task)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	response.Content = b.String()
	return response
}

func doneTask(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	err = model.DoneTask(model.Task{
		UserID: flash["UserID"],
		ID:     id,
	})
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	jsonRedirect(response, "/home")
	return response
}

func deleteTask(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	err = model.DeleteTask(model.Task{
		UserID: flash["UserID"],
		ID:     id,
	})
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	jsonRedirect(response, "/home")
	return response
}

func updateTask(w http.ResponseWriter, r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.PostFormValue("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	task := model.Task{
		UserID: flash["UserID"],
	}

	err = updateTaskFromForm(&task, r.PostForm)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id == 0 {
		task.ID = newID()
		task.Start = model.NewTime(time.Now().Unix())
		fmt.Println(task)
		err = model.CreateTask(task)
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
	} else {
		task.ID = id
		err = model.UpdateTask(task, taskSubject, taskDue, taskPriority, taskNotification, taskNext, taskNote)
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
	}

	jsonRedirect(response, "/home")
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
	r.Content = content
}

// updateTaskFromForm updates Subject, Due, Priority, Notification, Next, Note fields.
func updateTaskFromForm(task *model.Task, form url.Values) error {
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

	task.ParentTaskID, err = stoI64(form.Get("ParentTaskID"))
	if err != nil {
		return err
	}

	return nil
}
