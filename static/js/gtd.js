function getDateTime(ts) {
  if (ts == 0) {
    return ["", ""];
  }

  var addZero = function (s) {
    return s.length == 2 ? s : "0" + s;
  };
  var dt = new Date(ts * 1000);
  var year = String(dt.getUTCFullYear());
  var month = addZero(String(dt.getUTCMonth() + 1));
  var date = addZero(String(dt.getUTCDate()));
  var hours = addZero(String(dt.getUTCHours()));
  var minutes = addZero(String(dt.getUTCMinutes()));
  return [year + "-" + month + "-" + date, hours + ":" + minutes];
}

function getDateTimeString(ts, defaultString) {
  if (ts == 0) {
    return defaultString;
  }

  var dt = getDateTime(ts);
  return dt[0] + " " + dt[1];
}

function getURIFromHash() {
  return location.hash.substring(1) || "/home";
}

function getParamFromURI(uri) {
  var params = {};
  var query = uri.slice(uri.indexOf("?") + 1);
  var vars = query.split("&");
  for (var i = 0; i < vars.length; i++) {
    var pair = vars[i].split("=");
    if (typeof params[pair[0]] === "undefined") {
      params[pair[0]] = decodeURIComponent(pair[1]);
    } else if (typeof params[pair[0]] === "string") {
      params[pair[0]] = [params[pair[0]], decodeURIComponent(pair[1])];
    } else {
      params[pair[0]].push(decodeURIComponent(pair[1]));
    }
  }
  return params;
}

var Login = React.createClass({
  getInitialState: function () {
    return { UserID: "", Password: "" };
  },
  handleChange: function (event) {
    var state = {};
    state[event.target.name] = event.target.value;
    this.setState(state);
  },
  handleSubmit: function (event) {
    event.preventDefault();
    ajax("/auth", this.state, success, error);
  },
  render: function () {
    return React.createElement(
      "form",
      { className: "col-xs-12 col-sm-4", onSubmit: this.handleSubmit },
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "UserID"
        ),
        React.createElement("input", { name: "UserID", type: "text", className: "form-control", onChange: this.handleChange })
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Password"
        ),
        React.createElement("input", { name: "Password", type: "password", className: "form-control", onChange: this.handleChange })
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "button",
          { type: "submit", className: "btn btn-default form-control" },
          "Login"
        )
      )
    );
  }
});

var TaskListNode = React.createClass({
  render: function () {
    var task = this.props.task;

    var note;
    if (task.Note) {
      note = React.createElement(
        "pre",
        { className: "task-body-item" },
        task.Note
      );
    }

    var subTasks;
    if (this.props.subTasks) {
      subTasks = this.props.subTasks.map(function (task) {
        return React.createElement(TaskListNode, { key: task.ID, task: task });
      });
    }

    return React.createElement(
      "li",
      { className: task.Priority == 1 ? "danger" : "default" },
      React.createElement(
        "div",
        { className: "task-header" },
        React.createElement(
          "b",
          null,
          React.createElement(
            "span",
            { className: "label " + (task.Priority == 1 ? "label-danger" : "label-default") },
            task.Priority
          ),
          " ",
          task.Subject
        ),
        React.createElement(
          "span",
          { className: "text-nowrap" },
          getDateTimeString(task.Due, "No Due")
        )
      ),
      React.createElement(
        "div",
        { className: "task-body" },
        note,
        React.createElement(
          "ul",
          { className: "task-body-item task-list" },
          subTasks
        ),
        React.createElement(
          "div",
          { className: "task-body-item row" },
          React.createElement(
            "a",
            { href: "#/edit?ID=" + task.ID, className: "col-xs-6 col-md-3 text-primary" },
            React.createElement("i", { className: "fa fa-edit" }),
            " Edit"
          ),
          React.createElement(
            "a",
            { href: "#/done?ID=" + task.ID, className: "col-xs-6 col-md-3 text-success" },
            React.createElement("i", { className: "fa fa-check" }),
            " Done"
          ),
          React.createElement(
            "a",
            { href: "#/delete?ID=" + task.ID, className: "col-xs-6 col-md-3 text-danger" },
            React.createElement("i", { className: "fa fa-trash" }),
            " Delete"
          )
        )
      )
    );
  }
});

var TaskList = React.createClass({
  componentDidMount: function () {
    $(".task-list li .task-header").click(function () {
      $(this).siblings(".task-body").slideToggle();
    });
  },
  render: function () {
    if (this.props.list && this.props.list.length > 0) {
      var tasks = this.props.tasks;

      var tasksMap = {};
      for (var i = 0; i < tasks.length; i++) {
        tasksMap[tasks[i].ID] = tasks[i];
      }

      var taskNodes = this.props.list.map(function (taskID) {
        var _subTasks;
        if (tasksMap[taskID].SubTaskIDs) {
          _subTasks = tasksMap[taskID].SubTaskIDs.map(function (id) {
            return tasksMap[id];
          });
        }

        return React.createElement(TaskListNode, { key: taskID, task: tasksMap[taskID], subTasks: _subTasks });
      });
    } else {
      var taskNodes = React.createElement(
        "h3",
        null,
        "No Task"
      );
    }

    return React.createElement(
      "ul",
      { className: "col-xs-12 col-sm-6 task-list" },
      taskNodes
    );
  }
});

var TaskForm = React.createClass({
  getInitialState: function () {
    var state = {};
    state.Priority = 2;
    if (this.props.task) {
      var task = this.props.task;
      state.ID = task.ID;
      state.Subject = task.Subject;
      if (task.Due) {
        var due = getDateTime(task.Due);
        state.DueDate = due[0];
        state.DueTime = due[1];
      } else {
        state.NoDue = true;
      }
      state.Priority = task.Priority || state.Priority;
      if (task.Reminder) {
        var reminder = getDateTime(task.Reminder);
        state.ReminderDate = reminder[0];
        state.ReminderTime = reminder[1];
      } else {
        state.NoReminder = true;
      }
      if (task.Next) {
        state.Next = true;
        var next = getDateTime(task.Next);
        state.NextDate = next[0];
        state.NextTime = next[1];
      }
      state.Note = task.Note;
      state.ParentTaskID = task.ParentTaskID;
      if (task.Tags) {
        state.Tags = task.Tags.join(",");
      }
    }
    return state;
  },
  handleChange: function (event) {
    var state = {};
    state[event.target.name] = event.target.value;
    this.setState(state);
  },
  handleCheckboxChange: function (event) {
    var state = {};
    state[event.target.name] = event.target.checked;
    this.setState(state);
  },
  handleSubmit: function (event) {
    event.preventDefault();
    ajax("/update", this.state, success, error);
  },
  render: function () {
    return React.createElement(
      "form",
      { className: "col-xs-12 col-sm-8", onSubmit: this.handleSubmit },
      React.createElement(
        "div",
        { className: "row" },
        React.createElement(
          "div",
          { className: "form-group col-xs-12 col-sm-6" },
          React.createElement(
            "label",
            null,
            "ID"
          ),
          React.createElement("input", { type: "text", name: "ID", className: "form-control", value: this.state.ID, readOnly: true })
        ),
        React.createElement(
          "div",
          { className: "form-group col-xs-12 col-sm-6" },
          React.createElement(
            "label",
            null,
            "ParentTaskID"
          ),
          React.createElement("input", { type: "text", name: "ParentTaskID", className: "form-control", value: this.state.ParentTaskID, readOnly: true })
        )
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Subject"
        ),
        React.createElement("input", { type: "text", name: "Subject", className: "form-control", value: this.state.Subject, onChange: this.handleChange })
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Due"
        ),
        React.createElement(
          "div",
          { className: "row" },
          React.createElement(
            "div",
            { className: "col-xs-12 checkbox" },
            React.createElement(
              "label",
              null,
              React.createElement("input", { type: "checkbox", name: "NoDue", checked: this.state.NoDue, onChange: this.handleCheckboxChange }),
              "No Due"
            )
          ),
          React.createElement(
            "div",
            { className: "col-xs-7" },
            React.createElement("input", { type: "date", name: "DueDate", className: "form-control", value: this.state.DueDate, onChange: this.handleChange })
          ),
          React.createElement(
            "div",
            { className: "col-xs-5" },
            React.createElement("input", { type: "time", name: "DueTime", className: "form-control", value: this.state.DueTime, onChange: this.handleChange })
          )
        )
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Priority"
        ),
        React.createElement(
          "div",
          null,
          React.createElement(
            "label",
            { className: "radio-inline" },
            React.createElement("input", { type: "radio", name: "Priority", value: "1", checked: this.state.Priority == 1, onChange: this.handleChange }),
            " 1"
          ),
          React.createElement(
            "label",
            { className: "radio-inline" },
            React.createElement("input", { type: "radio", name: "Priority", value: "2", checked: this.state.Priority == 2, onChange: this.handleChange }),
            " 2"
          ),
          React.createElement(
            "label",
            { className: "radio-inline" },
            React.createElement("input", { type: "radio", name: "Priority", value: "3", checked: this.state.Priority == 3, onChange: this.handleChange }),
            " 3"
          )
        )
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Reminder"
        ),
        React.createElement(
          "div",
          { className: "row" },
          React.createElement(
            "div",
            { className: "col-xs-12 checkbox" },
            React.createElement(
              "label",
              null,
              React.createElement("input", { type: "checkbox", name: "NoReminder", checked: this.state.NoReminder, onChange: this.handleCheckboxChange }),
              "No Reminder"
            )
          ),
          React.createElement(
            "div",
            { className: "col-xs-7" },
            React.createElement("input", { type: "date", name: "ReminderDate", className: "form-control", value: this.state.ReminderDate, onChange: this.handleChange })
          ),
          React.createElement(
            "div",
            { className: "col-xs-5" },
            React.createElement("input", { type: "time", name: "ReminderTime", className: "form-control", value: this.state.ReminderTime, onChange: this.handleChange })
          )
        )
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Next"
        ),
        React.createElement(
          "div",
          { className: "row" },
          React.createElement(
            "div",
            { className: "col-xs-12 checkbox" },
            React.createElement(
              "label",
              null,
              React.createElement("input", { type: "checkbox", name: "Next", checked: this.state.Next, onChange: this.handleCheckboxChange }),
              "Next"
            )
          ),
          React.createElement(
            "div",
            { className: "col-xs-7" },
            React.createElement("input", { type: "date", name: "NextDate", className: "form-control", value: this.state.NextDate, onChange: this.handleChange })
          ),
          React.createElement(
            "div",
            { className: "col-xs-5" },
            React.createElement("input", { type: "time", name: "NextTime", className: "form-control", value: this.state.NextTime, onChange: this.handleChange })
          )
        )
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Note"
        ),
        React.createElement("textarea", { name: "Note", className: "form-control", rows: "6", value: this.state.Note, onChange: this.handleChange })
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "label",
          null,
          "Tags"
        ),
        React.createElement("input", { type: "text", name: "Tags", className: "form-control", value: this.state.Tags, onChange: this.handleChange })
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "button",
          { type: "submit", className: "btn btn-default form-control" },
          "Submit"
        )
      ),
      React.createElement(
        "div",
        { className: "form-group" },
        React.createElement(
          "a",
          { type: "button", href: "", id: "Done", className: "btn btn-success form-control hidden" },
          "Done"
        ),
        React.createElement(
          "a",
          { type: "button", href: "", id: "Delete", className: "btn btn-danger form-control hidden" },
          "Delete"
        )
      )
    );
  }
});

var ParentSubTask = React.createClass({
  render: function () {
    var parentTask;
    if (this.props.parentTask) {
      var _parentTask = this.props.parentTask;
      parentTask = React.createElement(
        "a",
        { href: "#/edit?ID=" + _parentTask.ID },
        React.createElement(
          "h4",
          null,
          React.createElement(
            "span",
            { className: "label " + (_parentTask.Priority == 1 ? "label-danger" : "label-default") },
            _parentTask.Priority
          ),
          " ",
          _parentTask.Subject
        ),
        React.createElement(
          "p",
          { className: "text-right" },
          getDateTimeString(_parentTask.Due, "No Due")
        )
      );
    } else {
      parentTask = React.createElement(
        "p",
        null,
        "None"
      );
    }

    var subTasks;
    if (this.props.subTasks) {
      subTasks = this.props.subTasks.map(function (task) {
        return React.createElement(
          "a",
          { key: task.ID, href: "#/edit?ID=" + task.ID },
          React.createElement(
            "h4",
            null,
            React.createElement(
              "span",
              { className: "label " + (task.Priority == 1 ? "label-danger" : "label-default") },
              task.Priority
            ),
            " ",
            task.Subject
          ),
          React.createElement(
            "p",
            { className: "text-right" },
            getDateTimeString(task.Due, "No Due")
          )
        );
      });
    } else {
      subTasks = React.createElement(
        "p",
        null,
        "None"
      );
    }

    return React.createElement(
      "div",
      { className: "col-xs-12 col-sm-4" },
      React.createElement(
        "h3",
        null,
        "Parent Task"
      ),
      parentTask,
      React.createElement(
        "h3",
        null,
        "Sub Tasks"
      ),
      subTasks,
      React.createElement(
        "a",
        { href: "#/newSub?ID=" + this.props.taskID },
        "New"
      )
    );
  }
});

var Tags = React.createClass({
  render: function () {
    var tags = this.props.tags.map(function (tag) {
      return React.createElement(
        "li",
        { key: tag },
        React.createElement(
          "a",
          { href: "#/tag?Name=" + tag },
          tag
        )
      );
    });
    return React.createElement(
      "ul",
      { className: "col-xs-12 col-sm-6" },
      tags
    );
  }
});

function getSuccessFunc(f) {
  return function (data) {
    switch (data.Status) {
      case "OK":
        if (f) {
          f(data.Data);
        }
        break;
      case "Authenticated":
        window.localStorage.setItem("token", data.Data["token"]);
        dispatch(getURIFromHash(), {});
        break;
      case "Redirect":
        location.hash = data.Data["URI"];
        break;
      case "AuthenticationFailed":
        ReactDOM.render(React.createElement(Login, null), document.getElementById("content"));
        break;
      case "Error":
        alert(data.Data["ErrorMessage"]);
        goBack();
        break;
    }
  };
}

var success = getSuccessFunc();

function error(data) {
  alert("Ajax Error: " + data.responseText);
  goBack();
}

function dispatch(uri, data) {
  switch (true) {
    case /^\/home$/.test(uri):
      ajax(uri, {}, getSuccessFunc(function (data) {
        ReactDOM.render(React.createElement(TaskList, { tasks: data["tasks"], list: data["list"] }), document.getElementById("content"));
        currentURI = uri;
      }), error);
      break;
    case /^\/new$/.test(uri):
      ReactDOM.render(React.createElement(TaskForm, null), document.getElementById("content"));
      currentURI = uri;
      break;
    case /^\/newSub/.test(uri):
      ReactDOM.render(React.createElement(TaskForm, { task: { ParentTaskID: getParamFromURI(uri)["ID"] } }), document.getElementById("content"));
      currentURI = uri;
      break;
    case /^\/edit/.test(uri):
      ajax(uri, {}, getSuccessFunc(function (data) {
        ReactDOM.render(React.createElement(
          "div",
          null,
          React.createElement(TaskForm, { key: data["task"].ID, task: data["task"] }),
          React.createElement(ParentSubTask, { taskID: data["task"].ID, parentTask: data["parentTask"], subTasks: data["subTasks"] })
        ), document.getElementById("content"));
        currentURI = uri;
      }), error);
      break;
    case /^\/done/.test(uri):
      ajax(uri, {}, success, error);
      break;
    case /^\/delete/.test(uri):
      ajax(uri, {}, success, error);
      break;
    case /^\/tags$/.test(uri):
      ajax(uri, {}, getSuccessFunc(function (data) {
        ReactDOM.render(React.createElement(Tags, { tags: data["tags"] }), document.getElementById("content"));
        currentURI = uri;
      }), error);
      break;
    case /^\/tag/.test(uri):
      ajax(uri, {}, getSuccessFunc(function (data) {
        ReactDOM.render(React.createElement(TaskList, { tasks: data["tasks"], list: data["list"] }), document.getElementById("content"));
        currentURI = uri;
      }), error);
      break;
    default:
      goBack();
      break;
  }
}

function goBack() {
  goingBack = true;
  location.hash = currentURI;
}

function handleHashChange() {
  if (!goingBack) {
    dispatch(getURIFromHash(), {});
  } else {
    goingBack = false;
  }
}

var goingBack = false;

var currentURI = "";

window.addEventListener("hashchange", handleHashChange);

handleHashChange();

