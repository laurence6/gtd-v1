package main

import (
	"errors"
	"github.com/laurence6/gtd.go/core"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04"
)

var location, _ = time.LoadLocation("Local")

var errInvalid = errors.New("Invalid!")

var t *template.Template

func web() {
	var err error
	t, err = template.ParseFiles(
		"templates/add.html",
		"templates/default.html",
		"templates/edit.html",
		"templates/form.html",
		"templates/index.html",
	)
	if err != nil {
		log.Fatalln(err.Error())
	}
	http.HandleFunc("/", index)
	http.HandleFunc("/add", add)
	http.HandleFunc("/edit", edit)
	http.HandleFunc("/update", updateTask)
	http.HandleFunc("/delete", deleteTask)
	err = http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	taskList := []*gtd.Task{}
	for _, v := range tp {
		taskList = append(taskList, v)
	}
	gtd.SortByDefault(taskList)
	htmlTaskList := []*htmlTask{}
	for _, v := range taskList {
		htmlTaskList = append(htmlTaskList, taskToHTMLTask(v))
	}
	err := t.ExecuteTemplate(w, "index", htmlTaskList)
	if err != nil {
		log.Println(err.Error())
		oops(w, r, err)
		return
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	err := t.ExecuteTemplate(w, "add", "")
	if err != nil {
		log.Println(err.Error())
		oops(w, r, err)
		return
	}
}

func edit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		oops(w, r, err)
		return
	}
	if id != 0 {
		if task, ok := tp[id]; ok {
			_ = t.ExecuteTemplate(w, "add", "")
			err = t.ExecuteTemplate(w, "edit", taskToHTMLTask(task))
			if err != nil {
				log.Println(err.Error())
				oops(w, r, err)
				return
			}
			return
		}
	}
	http.Redirect(w, r, "add", 302)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		id, err := stoI64(r.PostForm.Get("ID"))
		if err != nil {
			log.Println(err.Error())
			oops(w, r, err)
			return
		}
		if id == 0 {
			task, err := tp.NewTask()
			if err != nil {
				log.Println(err.Error())
				oops(w, r, err)
				return
			}
			err = updateTaskFromForm(task, r.PostForm)
			if err != nil {
				log.Println(err.Error())
				oops(w, r, err)
				tp.Delete(task)
				return
			}
			http.Redirect(w, r, "", 302)
			return
		} else if task, ok := tp[id]; ok {
			err := updateTaskFromForm(task, r.PostForm)
			if err != nil {
				log.Println(err.Error())
				oops(w, r, err)
			}
			http.Redirect(w, r, "", 302)
			return
		}
	}
	oops(w, r, errInvalid)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		log.Println(err.Error())
		oops(w, r, err)
		return
	}
	if id != 0 {
		if task, ok := tp[id]; ok {
			err := tp.Delete(task)
			if err != nil {
				log.Println(err.Error())
				oops(w, r, err)
				return
			}
			http.Redirect(w, r, "", 302)
			return
		}
	}
	oops(w, r, errInvalid)
}

func oops(w http.ResponseWriter, r *http.Request, err error) {
	w.Write([]byte("Oops: " + err.Error()))
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

type htmlTask struct {
	gtd.Task
	NoDue            bool
	DueDate          string
	DueTime          string
	Next             bool
	NextDate         string
	NextTime         string
	NoNotification   bool
	NotificationDate string
	NotificationTime string
}

func taskToHTMLTask(task *gtd.Task) *htmlTask {
	ht := &htmlTask{}
	ht.Task = *task
	if ht.Task.Due != 0 {
		due := time.Unix(ht.Task.Due, 0)
		ht.DueDate = due.Format(dateLayout)
		ht.DueTime = due.Format(timeLayout)
	} else {
		ht.NoDue = true
	}
	if ht.Task.Next != 0 {
		ht.Next = true
		next := time.Unix(ht.Task.Next, 0)
		ht.NextDate = next.Format(dateLayout)
		ht.NextTime = next.Format(timeLayout)
	}
	if ht.Task.Notification != 0 {
		notification := time.Unix(ht.Task.Notification, 0)
		ht.NotificationDate = notification.Format(dateLayout)
		ht.NotificationTime = notification.Format(timeLayout)
	} else {
		ht.NoNotification = true
	}
	return ht
}

// parseDateTimeInLocation parses date & time string into unix timestamp
func parseDateTimeInLocation(dateStr, timeStr string, loc *time.Location) (int64, error) {
	var datetime time.Time
	var err error
	if dateStr != "" {
		if timeStr != "" {
			datetime, err = time.ParseInLocation(dateLayout+timeLayout, dateStr+timeStr, location)
		} else {
			datetime, err = time.ParseInLocation(dateLayout, dateStr, location)
		}
		if err != nil {
			return 0, err
		}
		return datetime.Unix(), nil
	}
	return 0, nil
}

// updateTaskFromForm updates Subject, Due, Priority, Next, Notification, Note fields
func updateTaskFromForm(task *gtd.Task, form url.Values) error {
	var err error
	task.Subject = form.Get("Subject")

	noDue := form.Get("NoDue")
	if noDue == "on" {
		task.Due = 0
	} else {
		datetime, err := parseDateTimeInLocation(form.Get("DueDate"), form.Get("DueTime"), location)
		if err != nil {
			return err
		}
		task.Due = datetime
	}

	task.Priority, err = strconv.Atoi(form.Get("Priority"))
	if err != nil {
		return err
	}

	next := form.Get("Next")
	if next == "on" {
		datetime, err := parseDateTimeInLocation(form.Get("NextDate"), form.Get("NextTime"), location)
		if err != nil {
			return err
		}
		task.Next = datetime
	} else {
		task.Next = 0
	}

	noNotification := form.Get("NoNotification")
	if noNotification == "on" {
		task.Notification = 0
	} else {
		datetime, err := parseDateTimeInLocation(form.Get("NotificationDate"), form.Get("NotificationTime"), location)
		if err != nil {
			return err
		}
		task.Notification = datetime
	}

	task.Note = form.Get("Note")
	return nil
}
