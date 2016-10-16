function getDateTime(ts) {
  if (ts == 0) {
    return ["", ""]
  }

  var addZero = function(s) {
    return (s.length == 2) ? s : "0" + s
  }
  var dt = new Date(ts * 1000)
  var year = String(dt.getUTCFullYear())
  var month = addZero(String(dt.getUTCMonth() + 1))
  var date = addZero(String(dt.getUTCDate()))
  var hours = addZero(String(dt.getUTCHours()))
  var minutes = addZero(String(dt.getUTCMinutes()))
  return [
    year + "-" + month + "-" + date,
    hours + ":" + minutes,
  ]
}

function getDateTimeString(ts, defaultString) {
  if (ts == 0) {
    return defaultString
  }

  var dt = getDateTime(ts)
  return dt[0] + " " + dt[1]
}

function getURIFromHash() {
  return location.hash.substring(1) || "/home"
}

function getParamFromURI(uri) {
  var params = {}
  var query = uri.slice(uri.indexOf("?") + 1)
  var vars = query.split("&")
  for (var i = 0; i < vars.length; i++) {
    var pair = vars[i].split("=")
    if (typeof params[pair[0]] === "undefined") {
      params[pair[0]] = decodeURIComponent(pair[1])
    } else if (typeof params[pair[0]] === "string") {
      params[pair[0]] = [params[pair[0]], decodeURIComponent(pair[1])]
    } else {
      params[pair[0]].push(decodeURIComponent(pair[1]))
    }
  }
  return params
}

var Login = React.createClass({
  getInitialState: function() {
    return {UserID: "", Password: ""}
  },
  handleChange: function(event) {
    var state = {}
    state[event.target.name] = event.target.value
    this.setState(state)
  },
  handleSubmit: function(event) {
    event.preventDefault()
    ajax("/auth", this.state, success, error)
  },
  render: function() {
    return (
      <form className="col-xs-12 col-sm-4" onSubmit={this.handleSubmit}>
        <div className="form-group">
          <label>UserID</label>
          <input name="UserID" type="text" className="form-control" onChange={this.handleChange} />
        </div>

        <div className="form-group">
          <label>Password</label>
          <input name="Password" type="password" className="form-control" onChange={this.handleChange} />
        </div>

        <div className="form-group">
          <button type="submit" className="btn btn-default form-control">Login</button>
        </div>
      </form>
    )
  }
})

var TaskListNode = React.createClass({
  render: function() {
    var task = this.props.task

    var note
    if (task.Note) {
      note = <pre className="task-body-item">{task.Note}</pre>
    }

    var subTasks
    if (this.props.subTasks) {
      subTasks = this.props.subTasks.map(function(task) {
        return (<TaskListNode key={task.ID} task={task} />)
      })
    }

    return (
      <li className={(task.Priority == 1) ? "danger" : "default"}>
        <div className="task-header">
          <b><span className={"label " + ((task.Priority == 1) ? "label-danger" : "label-default")}>{task.Priority}</span> {task.Subject}</b>
          <span className="text-nowrap">{getDateTimeString(task.Due, "No Due")}</span>
        </div>
        <div className="task-body">
          {note}
          <ul className="task-body-item task-list">
            {subTasks}
          </ul>
          <div className="task-body-item row">
            <a href={"#/edit?ID=" + task.ID} className="col-xs-6 col-md-3 text-primary"><i className="fa fa-edit"></i> Edit</a>
            <a href={"#/done?ID=" + task.ID} className="col-xs-6 col-md-3 text-success"><i className="fa fa-check"></i> Done</a>
            <a href={"#/delete?ID=" + task.ID} className="col-xs-6 col-md-3 text-danger"><i className="fa fa-trash"></i> Delete</a>
          </div>
        </div>
      </li>
    )
  }
})

var TaskList = React.createClass({
  componentDidMount: function() {
    $(".task-list li .task-header").click(function() {
      $(this).siblings(".task-body").slideToggle()
    })
  },
  render: function() {
    if (this.props.list && this.props.list.length > 0) {
      var tasks = this.props.tasks

      var tasksMap = {}
      for (var i = 0; i < tasks.length; i++) {
        tasksMap[tasks[i].ID] = tasks[i]
      }

      var taskNodes = this.props.list.map(function(taskID) {
        var _subTasks
        if (tasksMap[taskID].SubTaskIDs) {
          _subTasks = tasksMap[taskID].SubTaskIDs.map(function(id) {
            return tasksMap[id]
          })
        }

        return <TaskListNode key={taskID} task={tasksMap[taskID]} subTasks={_subTasks} />
      })
    } else {
      var taskNodes = <h3>No Task</h3>
    }

    return (
      <ul className="col-xs-12 col-sm-6 task-list">
        {taskNodes}
      </ul>
    )
  }
})

var TaskForm = React.createClass({
  getInitialState: function() {
    var state = {}
    state.Priority = 2
    if (this.props.task) {
      var task = this.props.task
      state.ID = task.ID
      state.Subject = task.Subject
      if (task.Due) {
        var due = getDateTime(task.Due)
        state.DueDate = due[0]
        state.DueTime = due[1]
      } else {
        state.NoDue = true
      }
      state.Priority = task.Priority || state.Priority
      if (task.Reminder) {
        var reminder = getDateTime(task.Reminder)
        state.ReminderDate = reminder[0]
        state.ReminderTime = reminder[1]
      } else {
        state.NoReminder = true
      }
      if (task.Next) {
        state.Next = true
        var next = getDateTime(task.Next)
        state.NextDate = next[0]
        state.NextTime = next[1]
      }
      state.Note = task.Note
      state.ParentTaskID = task.ParentTaskID
      if (task.Tags) {
        state.Tags = task.Tags.join(",")
      }
    }
    return state
  },
  handleChange: function(event) {
    var state = {}
    state[event.target.name] = event.target.value
    this.setState(state)
  },
  handleCheckboxChange: function(event) {
    var state = {}
    state[event.target.name] = event.target.checked
    this.setState(state)
  },
  handleSubmit: function(event) {
    event.preventDefault()
    ajax("/update", this.state, success, error)
  },
  render: function() {
    return (
      <form className="col-xs-12 col-sm-8" onSubmit={this.handleSubmit}>
        <div className="row">
          <div className="form-group col-xs-12 col-sm-6">
            <label>ID</label>
            <input type="text" name="ID" className="form-control" value={this.state.ID} readOnly />
          </div>
          <div className="form-group col-xs-12 col-sm-6">
            <label>ParentTaskID</label>
            <input type="text" name="ParentTaskID" className="form-control" value={this.state.ParentTaskID} readOnly />
          </div>
        </div>

        <div className="form-group">
          <label>Subject</label>
          <input type="text" name="Subject" className="form-control" value={this.state.Subject} onChange={this.handleChange} />
        </div>

        <div className="form-group">
          <label>Due</label>
          <div className="row">
            <div className="col-xs-12 checkbox">
              <label>
                <input type="checkbox" name="NoDue" checked={this.state.NoDue} onChange={this.handleCheckboxChange} />No Due
              </label>
            </div>
            <div className="col-xs-7">
              <input type="date" name="DueDate" className="form-control" value={this.state.DueDate} onChange={this.handleChange} />
            </div>
            <div className="col-xs-5">
              <input type="time" name="DueTime" className="form-control" value={this.state.DueTime} onChange={this.handleChange} />
            </div>
          </div>
        </div>

        <div className="form-group">
          <label>Priority</label>
          <div>
            <label className="radio-inline"><input type="radio" name="Priority" value="1" checked={(this.state.Priority == 1)} onChange={this.handleChange} /> 1</label>
            <label className="radio-inline"><input type="radio" name="Priority" value="2" checked={(this.state.Priority == 2)} onChange={this.handleChange} /> 2</label>
            <label className="radio-inline"><input type="radio" name="Priority" value="3" checked={(this.state.Priority == 3)} onChange={this.handleChange} /> 3</label>
          </div>
        </div>

        <div className="form-group">
          <label>Reminder</label>
          <div className="row">
            <div className="col-xs-12 checkbox">
              <label>
                <input type="checkbox" name="NoReminder" checked={this.state.NoReminder} onChange={this.handleCheckboxChange} />No Reminder
              </label>
            </div>
            <div className="col-xs-7">
              <input type="date" name="ReminderDate" className="form-control" value={this.state.ReminderDate} onChange={this.handleChange} />
            </div>
            <div className="col-xs-5">
              <input type="time" name="ReminderTime" className="form-control" value={this.state.ReminderTime} onChange={this.handleChange} />
            </div>
          </div>
        </div>

        <div className="form-group">
          <label>Next</label>
          <div className="row">
            <div className="col-xs-12 checkbox">
              <label>
                <input type="checkbox" name="Next" checked={this.state.Next} onChange={this.handleCheckboxChange} />Next
              </label>
            </div>
            <div className="col-xs-7">
              <input type="date" name="NextDate" className="form-control" value={this.state.NextDate} onChange={this.handleChange} />
            </div>
            <div className="col-xs-5">
              <input type="time" name="NextTime" className="form-control" value={this.state.NextTime} onChange={this.handleChange} />
            </div>
          </div>
        </div>

        <div className="form-group">
          <label>Note</label>
          <textarea name="Note" className="form-control" rows="6" value={this.state.Note} onChange={this.handleChange} />
        </div>

        <div className="form-group">
          <label>Tags</label>
          <input type="text" name="Tags" className="form-control" value={this.state.Tags} onChange={this.handleChange} />
        </div>

        <div className="form-group">
          <button type="submit" className="btn btn-default form-control">Submit</button>
        </div>

        <div className="form-group">
          <a type="button" href="" id="Done" className="btn btn-success form-control hidden">Done</a>
          <a type="button" href="" id="Delete" className="btn btn-danger form-control hidden">Delete</a>
        </div>
      </form>
    )
  }
})

var ParentSubTask = React.createClass({
  render: function() {
    var parentTask
    if (this.props.parentTask) {
      var _parentTask = this.props.parentTask
      parentTask = (
        <a href={"#/edit?ID=" + _parentTask.ID}>
          <h4><span className={"label " + ((_parentTask.Priority == 1) ? "label-danger" : "label-default")}>{_parentTask.Priority}</span> {_parentTask.Subject}</h4>
          <p className="text-right">{getDateTimeString(_parentTask.Due, "No Due")}</p>
        </a>
      )
    } else {
      parentTask = <p>None</p>
    }

    var subTasks
    if (this.props.subTasks) {
      subTasks = this.props.subTasks.map(function(task) {
        return (
          <a key={task.ID} href={"#/edit?ID=" + task.ID}>
            <h4><span className={"label " + ((task.Priority == 1) ? "label-danger" : "label-default")}>{task.Priority}</span> {task.Subject}</h4>
            <p className="text-right">{getDateTimeString(task.Due, "No Due")}</p>
          </a>
        )
      })
    } else {
      subTasks = <p>None</p>
    }

    return (
      <div className="col-xs-12 col-sm-4">
        <h3>Parent Task</h3>
        {parentTask}

        <h3>Sub Tasks</h3>
        {subTasks}
        <a href={"#/newSub?ID=" + this.props.taskID}>New</a>
      </div>
    )
  }
})

var Tags = React.createClass({
  render: function() {
    var tags = this.props.tags.map(function(tag) {
      return <li key={tag}><a href={"#/tag?Name=" + tag}>{tag}</a></li>
    })
    return (
      <ul className="col-xs-12 col-sm-6">
        {tags}
      </ul>
    )
  }
})

function getSuccessFunc(f) {
  return function (data) {
    switch (data.Status) {
    case "OK":
      if (f) {
        f(data.Data)
      }
      break
    case "Authenticated":
      window.localStorage.setItem("token", data.Data["token"])
      dispatch(getURIFromHash(), {})
      break
    case "Redirect":
      location.hash = data.Data["URI"]
      break
    case "AuthenticationFailed":
      ReactDOM.render(
        <Login />,
        document.getElementById("content")
      )
      break
    case "Error":
      alert(data.Data["ErrorMessage"])
      goBack()
      break
    }
  }
}

var success = getSuccessFunc()

function error(data) {
  alert("Ajax Error: " + data.responseText)
  goBack()
}

function dispatch(uri, data) {
  switch (true) {
  case /^\/home$/.test(uri):
    ajax(uri, {}, getSuccessFunc(function(data) {
      ReactDOM.render(
        <TaskList tasks={data["tasks"]} list={data["list"]} />,
        document.getElementById("content")
      )
      currentURI = uri
    }), error)
    break
  case /^\/new$/.test(uri):
    ReactDOM.render(
      <TaskForm />,
      document.getElementById("content")
    )
    currentURI = uri
    break
  case /^\/newSub/.test(uri):
    ReactDOM.render(
      <TaskForm task={{ParentTaskID: getParamFromURI(uri)["ID"]}} />,
      document.getElementById("content")
    )
    currentURI = uri
    break
  case /^\/edit/.test(uri):
    ajax(uri, {}, getSuccessFunc(function(data) {
      ReactDOM.render(
        <div>
          <TaskForm key={data["task"].ID} task={data["task"]} />
          <ParentSubTask taskID={data["task"].ID} parentTask={data["parentTask"]} subTasks={data["subTasks"]} />
        </div>,
        document.getElementById("content")
      )
      currentURI = uri
    }), error)
    break
  case /^\/done/.test(uri):
    ajax(uri, {}, success, error)
    break
  case /^\/delete/.test(uri):
    ajax(uri, {}, success, error)
    break
  case /^\/tags$/.test(uri):
    ajax(uri, {}, getSuccessFunc(function(data) {
      ReactDOM.render(
        <Tags tags={data["tags"]} />,
        document.getElementById("content")
      )
      currentURI = uri
    }), error)
    break
  case /^\/tag/.test(uri):
    ajax(uri, {}, getSuccessFunc(function(data) {
      ReactDOM.render(
        <TaskList tasks={data["tasks"]} list={data["list"]} />,
        document.getElementById("content")
      )
      currentURI = uri
    }), error)
    break
  default:
    goBack()
    break
  }
}

function goBack() {
  goingBack = true
  location.hash = currentURI
}

function handleHashChange() {
  if (!goingBack) {
    dispatch(getURIFromHash(), {})
  } else {
    goingBack = false
  }
}

var goingBack = false

var currentURI = ""

window.addEventListener("hashchange", handleHashChange)

handleHashChange()
