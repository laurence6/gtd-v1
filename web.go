package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/laurence6/gtd.go/model"
)

const (
	dateLayout  = "2006-01-02"
	timeLayout  = "15:04"
	tokenExpire = 60 * 60 * 24 * 30 // 30 Days
)

var landingHTML []byte

func init() {
	f, err := os.Open("static/default.html")
	if err != nil {
		logger.Fatalln(err.Error())
	}

	landingHTML, err = ioutil.ReadAll(f)
	if err != nil {
		logger.Fatalln(err.Error())
	}
}

func web() {
	http.HandleFunc("/auth", jsonHandlerWrapper(home))

	http.HandleFunc("/home", jsonHandlerWrapper(home))
	http.HandleFunc("/edit", jsonHandlerWrapper(editTask))
	http.HandleFunc("/done", jsonHandlerWrapper(doneTask))
	http.HandleFunc("/delete", jsonHandlerWrapper(deleteTask))
	http.HandleFunc("/update", jsonHandlerWrapper(updateTask))
	http.HandleFunc("/tags", jsonHandlerWrapper(tags))
	http.HandleFunc("/tag", jsonHandlerWrapper(tag))

	http.HandleFunc("/", landing)
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
	w.Write(landingHTML)
}

const (
	jsonStatusOK                   = "OK"
	jsonStatusAuthenticated        = "Authenticated"
	jsonStatusAuthenticationFailed = "AuthenticationFailed"
	jsonStatusRedirect             = "Redirect"
	jsonStatusError                = "Error"
)

type responseJSON struct {
	Status string
	Data   map[string]interface{}
}

func newResponseJSON() *responseJSON {
	return &responseJSON{
		Status: jsonStatusOK,
		Data:   map[string]interface{}{},
	}
}

func jsonHandlerWrapper(f func(*http.Request, Flash) *responseJSON) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			httpError(w, 405, "POST")
			return
		}

		var response *responseJSON

		flash := Flash{}
		response = auth(r, flash)
		if response == nil {
			response = f(r, flash)
		}

		var rJSON []byte
		if response != nil {
			var err error
			rJSON, err = json.Marshal(response)
			if err != nil {
				logger.Print(err)
				response = newResponseJSON()
				jsonError(response, err.Error())
				rJSON, _ = json.Marshal(response)
			}
			w.Write(rJSON)
		}
	}
}

func auth(r *http.Request, flash Flash) *responseJSON {
	r.ParseForm()

	response := newResponseJSON()

	if r.RequestURI == "/auth" {
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
			response.Status = jsonStatusAuthenticated
			response.Data["token"] = token
			return response
		}

		response.Status = jsonStatusAuthenticationFailed
		return response
	}

	userID, err := CheckToken(r.PostFormValue("token"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	if userID != "" {
		flash["UserID"] = userID
		return nil
	}

	response.Status = jsonStatusAuthenticationFailed
	return response
}

func home(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	tasks, err := model.GetTasksByUserID(flash["UserID"],
		model.CTaskID|
			model.CTaskSubject|
			model.CTaskDue|
			model.CTaskPriority|
			model.CTaskNote|
			model.CTaskSubTaskIDs)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	SortByDefault(tasks)

	list := getTaskList(tasks)

	response.Data["tasks"] = tasks
	response.Data["list"] = list
	return response
}

func editTask(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	task, err := model.GetTask(flash["UserID"], id,
		model.CTaskID|
			model.CTaskSubject|
			model.CTaskDue|
			model.CTaskPriority|
			model.CTaskReminder|
			model.CTaskNext|
			model.CTaskNote|
			model.CTaskParentTaskID|
			model.CTaskSubTaskIDs|
			model.CTaskTags)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	response.Data["task"] = task

	if task.ParentTaskID != 0 {
		parentTask, err := model.GetTask(flash["UserID"], task.ParentTaskID,
			model.CTaskID|
				model.CTaskSubject|
				model.CTaskDue|
				model.CTaskPriority)
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
		response.Data["parentTask"] = parentTask
	}

	if len(task.SubTaskIDs) != 0 {
		subTasks, err := model.GetTasksByID(flash["UserID"], task.SubTaskIDs,
			model.CTaskID|
				model.CTaskSubject|
				model.CTaskDue|
				model.CTaskPriority)
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
		response.Data["subTasks"] = subTasks
	}

	return response
}

func doneTask(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	task := model.Task{}
	task.UserID = flash["UserID"]
	task.ID = id

	err = model.DoneTask(task)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	jsonRedirect(response, "/home")
	return response
}

func deleteTask(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.Form.Get("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	task := model.Task{}
	task.UserID = flash["UserID"]
	task.ID = id

	err = model.DeleteTask(task)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	jsonRedirect(response, "/home")
	return response
}

func updateTask(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	id, err := stoI64(r.PostFormValue("ID"))
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	task := model.Task{}
	task.UserID = flash["UserID"]

	err = updateTaskFromForm(&task, r.PostForm)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	if id == 0 {
		task.ID = newID()
		err = model.CreateTask(task)
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
	} else {
		task.ID = id
		err = model.UpdateTask(task,
			model.CTaskSubject|
				model.CTaskDue|
				model.CTaskPriority|
				model.CTaskReminder|
				model.CTaskNext|
				model.CTaskNote|
				model.CTaskTags)
		if err != nil {
			logger.Println(err.Error())
			jsonError(response, err.Error())
			return response
		}
	}

	jsonRedirect(response, "/home")
	return response
}

func tags(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	tags, err := model.GetTagsByUserID(flash["UserID"])
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	response.Data["tags"] = tags
	return response
}

func tag(r *http.Request, flash Flash) *responseJSON {
	response := newResponseJSON()

	r.ParseForm()
	name := r.FormValue("Name")

	tasks, err := model.GetTasksByTag(flash["UserID"], name,
		model.CTaskID|
			model.CTaskSubject|
			model.CTaskDue|
			model.CTaskPriority|
			model.CTaskNote|
			model.CTaskSubTaskIDs)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}

	SortByDefault(tasks)

	list := getTaskList(tasks)

	psTasks, err := model.GetTasksByID(flash["UserID"], getMissParentSubTaskIDs(tasks),
		model.CTaskID|
			model.CTaskSubject|
			model.CTaskDue|
			model.CTaskPriority|
			model.CTaskNote)
	if err != nil {
		logger.Println(err.Error())
		jsonError(response, err.Error())
		return response
	}
	tasks = append(tasks, psTasks...)

	response.Data["tasks"] = tasks
	response.Data["list"] = list
	return response
}

func getTaskList(tasks model.Tasks) []string {
	list := make([]string, len(tasks))
	for n, task := range tasks {
		list[n] = strconv.FormatInt(task.ID, 10)
	}
	return list
}

func getMissParentSubTaskIDs(tasks model.Tasks) []int64 {
	tasksMap := make(map[int64]int8, len(tasks))
	for _, task := range tasks {
		tasksMap[task.ID] = 1

		if task.ParentTaskID != 0 {
			if _, ok := tasksMap[task.ParentTaskID]; !ok {
				tasksMap[task.ParentTaskID] = 0
			}
		}

		for _, subTaskID := range task.SubTaskIDs {
			if _, ok := tasksMap[subTaskID]; !ok {
				tasksMap[subTaskID] = 0
			}
		}
	}

	ids := []int64{}
	for taskID, bit := range tasksMap {
		if bit == 0 {
			ids = append(ids, taskID)
		}
	}
	return ids
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
	fmt.Fprintf(w, "%s %s", codeMessage, message)
}

func jsonRedirect(r *responseJSON, uri string) {
	r.Status = jsonStatusRedirect
	r.Data["URI"] = uri
}

func jsonError(r *responseJSON, message string) {
	r.Status = jsonStatusError
	r.Data["ErrorMessage"] = message
}

// updateTaskFromForm updates Subject, Due, Priority, Reminder, Next, Note fields.
func updateTaskFromForm(task *model.Task, form url.Values) error {
	var err error
	task.Subject = form.Get("Subject")

	noDue := form.Get("NoDue")
	if noDue == "true" {
		task.Due.Set(0)
	} else {
		err := task.Due.ParseDateTime(form.Get("DueDate"), form.Get("DueTime"))
		if err != nil {
			return err
		}
	}

	task.Priority, err = strconv.Atoi(form.Get("Priority"))
	if err != nil {
		return err
	}

	noReminder := form.Get("NoReminder")
	if noReminder == "true" {
		task.Reminder.Set(0)
	} else {
		err := task.Reminder.ParseDateTime(form.Get("ReminderDate"), form.Get("ReminderTime"))
		if err != nil {
			return err
		}
	}

	next := form.Get("Next")
	if next == "true" {
		err := task.Next.ParseDateTime(form.Get("NextDate"), form.Get("NextTime"))
		if err != nil {
			return err
		}
	} else {
		task.Next.Set(0)
	}

	task.Note = form.Get("Note")

	if tagsStr := strings.Trim(form.Get("Tags"), ","); tagsStr != "" {
		task.Tags = strings.Split(tagsStr, ",")
	} else {
		task.Tags = []string{}
	}

	task.ParentTaskID, err = stoI64(form.Get("ParentTaskID"))
	if err != nil {
		return err
	}

	return nil
}
