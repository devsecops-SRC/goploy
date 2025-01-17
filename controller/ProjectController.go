package controller

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"goploy/core"
	"goploy/model"
	"goploy/utils"
)

// Project struct
type Project Controller

// GetList project list
func (project Project) GetList(w http.ResponseWriter, gp *core.Goploy) {
	type RespData struct {
		Project    model.Projects   `json:"projectList"`
		Pagination model.Pagination `json:"pagination"`
	}
	pagination, err := model.PaginationFrom(gp.URLQuery)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}

	var projectList model.Projects
	if gp.UserInfo.Role == core.RoleAdmin || gp.UserInfo.Role == core.RoleManager {
		projectList, pagination, err = model.Project{}.GetList(pagination)
	} else {
		projectList, pagination, err = model.Project{}.GetListInGroupIDs(strings.Split(gp.UserInfo.ManageGroupStr, ","), pagination)
	}

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Data: RespData{Project: projectList, Pagination: pagination}}
	response.JSON(w)
}

// GetOption Project list
func (project Project) GetOption(w http.ResponseWriter, gp *core.Goploy) {
	type RespData struct {
		Projects model.Projects `json:"projectList"`
	}

	projectList, err := model.Project{}.GetAll()

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Data: RespData{Projects: projectList}}
	response.JSON(w)
}

// GetBindServerList project detail
func (project Project) GetBindServerList(w http.ResponseWriter, gp *core.Goploy) {
	type RespData struct {
		ProjectServers model.ProjectServers `json:"projectServerMap"`
	}
	id, err := strconv.ParseInt(gp.URLQuery.Get("id"), 10, 64)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: "id参数错误"}
		response.JSON(w)
		return
	}
	projectServersMap, err := model.ProjectServer{ProjectID: id}.GetBindServerListByProjectID()
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Data: RespData{ProjectServers: projectServersMap}}
	response.JSON(w)
}

// GetBindUserList project detail
func (project Project) GetBindUserList(w http.ResponseWriter, gp *core.Goploy) {
	type RespData struct {
		ProjectUsers model.ProjectUsers `json:"projectUserMap"`
	}
	id, err := strconv.ParseInt(gp.URLQuery.Get("id"), 10, 64)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: "id参数错误"}
		response.JSON(w)
		return
	}
	projectUsersMap, err := model.ProjectUser{ProjectID: id}.GetBindUserListByProjectID()
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Data: RespData{ProjectUsers: projectUsersMap}}
	response.JSON(w)
}

// GetBindUserList project detail
func (project Project) GetBindProjectList(w http.ResponseWriter, gp *core.Goploy) {
	type RespData struct {
		ProjectUsers model.ProjectUsers `json:"projectUserMap"`
	}
	userID, err := strconv.ParseInt(gp.URLQuery.Get("userId"), 10, 64)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: "id参数错误"}
		response.JSON(w)
		return
	}
	projectUsersMap, err := model.ProjectUser{UserID: userID}.GetBindProjectListByUserID()
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Data: RespData{ProjectUsers: projectUsersMap}}
	response.JSON(w)
}

// Add one project
func (project Project) Add(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		GroupID           int64   `json:"groupId"`
		Name              string  `json:"name"`
		URL               string  `json:"url"`
		Path              string  `json:"path"`
		Environment       string  `json:"Environment"`
		Branch            string  `json:"branch"`
		AfterPullScript   string  `json:"afterPullScript"`
		AfterDeployScript string  `json:"afterDeployScript"`
		RsyncOption       string  `json:"rsyncOption"`
		AutoDeploy        uint8   `json:"autoDeploy"`
		ServerIDs         []int64 `json:"serverIds"`
		UserIDs           []int64 `json:"userIds"`
		NotifyType        uint8   `json:"notifyType"`
		NotifyTarget      string  `json:"notifyTarget"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}

	if _, err := utils.ParseCommandLine(reqData.RsyncOption); err != nil {
		response := core.Response{Code: core.Deny, Message: "Rsync Option错误，请输入正确的参数格式"}
		response.JSON(w)
		return
	}

	_, err = model.Project{Name: reqData.Name}.GetDataByName()
	if err != sql.ErrNoRows {
		response := core.Response{Code: core.Deny, Message: "项目名称已存在"}
		response.JSON(w)
		return
	}

	projectID, err := model.Project{
		GroupID:           reqData.GroupID,
		Name:              reqData.Name,
		URL:               reqData.URL,
		Path:              reqData.Path,
		Environment:       reqData.Environment,
		Branch:            reqData.Branch,
		AfterPullScript:   reqData.AfterPullScript,
		AfterDeployScript: reqData.AfterDeployScript,
		RsyncOption:       reqData.RsyncOption,
		AutoDeploy:        reqData.AutoDeploy,
		NotifyType:        reqData.NotifyType,
		NotifyTarget:      reqData.NotifyTarget,
		CreateTime:        time.Now().Unix(),
		UpdateTime:        time.Now().Unix(),
	}.AddRow()

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	projectServersModel := model.ProjectServers{}
	for _, serverID := range reqData.ServerIDs {
		projectServerModel := model.ProjectServer{
			ProjectID:  projectID,
			ServerID:   serverID,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		}
		projectServersModel = append(projectServersModel, projectServerModel)
	}

	if err := projectServersModel.AddMany(); err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	projectUsersModel := model.ProjectUsers{}
	for _, userID := range reqData.UserIDs {
		projectUserModel := model.ProjectUser{
			ProjectID:  projectID,
			UserID:     userID,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		}
		projectUsersModel = append(projectUsersModel, projectUserModel)
	}

	if err := projectUsersModel.AddMany(); err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	go repoCreate(projectID)
	response := core.Response{Message: "添加成功"}
	response.JSON(w)
}

// Edit one Project
func (project Project) Edit(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		ID                int64  `json:"id"`
		GroupID           int64  `json:"groupId"`
		Name              string `json:"name"`
		URL               string `json:"url"`
		Path              string `json:"path"`
		Environment       string `json:"Environment"`
		Branch            string `json:"branch"`
		AfterPullScript   string `json:"afterPullScript"`
		AfterDeployScript string `json:"afterDeployScript"`
		RsyncOption       string `json:"rsyncOption"`
		AutoDeploy        uint8  `json:"autoDeploy"`
		NotifyType        uint8  `json:"notifyType"`
		NotifyTarget      string `json:"notifyTarget"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}

	if _, err := utils.ParseCommandLine(reqData.RsyncOption); err != nil {
		response := core.Response{Code: core.Deny, Message: "Rsync Option错误，请输入正确的参数格式"}
		response.JSON(w)
		return
	}

	projectData, err := model.Project{Name: reqData.Name}.GetDataByName()
	if err != nil {
		if err != sql.ErrNoRows {
			response := core.Response{Code: core.Deny, Message: err.Error()}
			response.JSON(w)
			return
		}
	} else {
		if projectData.ID != reqData.ID {
			response := core.Response{Code: core.Deny, Message: "项目名称已存在"}
			response.JSON(w)
			return
		}
	}

	err = model.Project{
		ID:                reqData.ID,
		GroupID:           reqData.GroupID,
		Name:              reqData.Name,
		URL:               reqData.URL,
		Path:              reqData.Path,
		Environment:       reqData.Environment,
		Branch:            reqData.Branch,
		AfterPullScript:   reqData.AfterPullScript,
		AfterDeployScript: reqData.AfterDeployScript,
		RsyncOption:       reqData.RsyncOption,
		AutoDeploy:        reqData.AutoDeploy,
		NotifyType:        reqData.NotifyType,
		NotifyTarget:      reqData.NotifyTarget,
		UpdateTime:        time.Now().Unix(),
	}.EditRow()

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	go repoCreate(reqData.ID)
	response := core.Response{Message: "修改成功"}
	response.JSON(w)
}

// Remove one Project
func (project Project) Remove(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		ID int64 `json:"id"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	err = model.Project{
		ID:         reqData.ID,
		UpdateTime: time.Now().Unix(),
	}.RemoveRow()

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Message: "删除成功"}
	response.JSON(w)
}

// AddServer one project
func (project Project) AddServer(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		ProjectID int64   `json:"projectId"`
		ServerIDs []int64 `json:"serverIds"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	projectID := reqData.ProjectID

	projectServersModel := model.ProjectServers{}
	for _, serverID := range reqData.ServerIDs {
		projectServerModel := model.ProjectServer{
			ProjectID:  projectID,
			ServerID:   serverID,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		}
		projectServersModel = append(projectServersModel, projectServerModel)
	}

	if err := projectServersModel.AddMany(); err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Message: "添加成功"}
	response.JSON(w)
}

// AddUser one project
func (project Project) AddUser(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		ProjectID int64   `json:"projectId"`
		UserIDs   []int64 `json:"userIds"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	projectID := reqData.ProjectID

	projectUsersModel := model.ProjectUsers{}
	for _, userID := range reqData.UserIDs {
		projectUserModel := model.ProjectUser{
			ProjectID:  projectID,
			UserID:     userID,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		}
		projectUsersModel = append(projectUsersModel, projectUserModel)
	}

	if err := projectUsersModel.AddMany(); err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Message: "添加成功"}
	response.JSON(w)
}

// RemoveProjectServer one Project
func (project Project) RemoveProjectServer(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		ProjectServerID int64 `json:"projectServerId"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	err = model.ProjectServer{
		ID: reqData.ProjectServerID,
	}.DeleteRow()

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Message: "删除成功"}
	response.JSON(w)
}

// RemoveProjectUser one Project
func (project Project) RemoveProjectUser(w http.ResponseWriter, gp *core.Goploy) {
	type ReqData struct {
		ProjectUserID int64 `json:"projectUserId"`
	}
	var reqData ReqData
	err := json.Unmarshal(gp.Body, &reqData)
	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	err = model.ProjectUser{
		ID: reqData.ProjectUserID,
	}.DeleteRow()

	if err != nil {
		response := core.Response{Code: core.Deny, Message: err.Error()}
		response.JSON(w)
		return
	}
	response := core.Response{Message: "删除成功"}
	response.JSON(w)
}

func repoCreate(projectID int64) {
	project, err := model.Project{ID: projectID}.GetData()
	if err != nil {
		core.Log(core.TRACE, "projectID:"+strconv.FormatInt(projectID, 10)+" 无此项目")
		return
	}
	srcPath := core.RepositoryPath + project.Name
	if _, err := os.Stat(srcPath); err != nil {
		if err := os.RemoveAll(srcPath); err != nil {
			core.Log(core.TRACE, "projectID:"+strconv.FormatInt(project.ID, 10)+" 项目移除失败")
			return
		}
		repo := project.URL
		cmd := exec.Command("git", "clone", repo, srcPath)
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			core.Log(core.ERROR, "projectID:"+strconv.FormatInt(project.ID, 10)+" 项目初始化失败:"+err.Error())
			return
		}

		if project.Branch != "master" {
			checkout := exec.Command("git", "checkout", "-b", project.Branch, "origin/"+project.Branch)
			checkout.Dir = srcPath
			var checkoutOutbuf, checkoutErrbuf bytes.Buffer
			checkout.Stdout = &checkoutOutbuf
			checkout.Stderr = &checkoutErrbuf
			if err := checkout.Run(); err != nil {
				core.Log(core.ERROR, "projectID:"+strconv.FormatInt(project.ID, 10)+checkoutErrbuf.String())
				os.RemoveAll(srcPath)
				return
			}

		}
		core.Log(core.TRACE, "projectID:"+strconv.FormatInt(project.ID, 10)+" 项目初始化成功")

	}
	return
}
