package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/mlog"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"path"
	"regexp"
	"strings"
	"time"
)

type IPlugin interface {
	GetConfig() *model.Config
	GetUser(userId string) (*model.User, *model.AppError)
	CreateNewMeeting(user *model.User) (*NewMeetingResponse, error)
	CreatePost(post *model.Post) (*model.Post, *model.AppError)
}

func (p *Plugin) GetConfig() *model.Config {
	return p.API.GetConfig()
}

func (p *Plugin) GetUser(userId string) (*model.User, *model.AppError) {
	return p.API.GetUser(userId)
}

func (p *Plugin) CreatePost(post *model.Post) (*model.Post, *model.AppError) {
	return p.API.CreatePost(post)
}

type CurrentDate struct {
	Value time.Time
}

func getCommand() *model.Command {
	return &model.Command{
		Trigger:          "s4b",
		DisplayName:      "Skype for Business",
		Description:      "Skype for Business meeting.",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands: s4b",
		AutoCompleteHint: "[command]",
	}
}

func getCommandResponse(responseType, username string, text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     username,
		Type:         POST_MEETING_TYPE,
	}
}

func (p *Plugin) CreateNewMeeting(user *model.User) (*NewMeetingResponse, error) {
	applicationState, apiErr := p.fetchOnlineMeetingsUrl()
	if apiErr != nil {
		mlog.Error("Error fetching meetings resource url: " + apiErr.Message)
		return nil, &model.AppError{Message: "Error fetching meetings resource url: " + apiErr.Message}
	}

	newMeetingResponse, err := p.client.createNewMeeting(
		applicationState.OnlineMeetingsUrl,
		NewMeetingRequest{
			Subject:                   "Meeting created by " + user.Username,
			AutomaticLeaderAssignment: "SameEnterprise",
		},
		applicationState.Token,
	)
	if err != nil {
		mlog.Error("Error creating a new meeting: " + err.Error())
		return nil, &model.AppError{Message: "Error creating a new meeting: " + err.Error()}
	}

	return newMeetingResponse, nil
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	return executeCommand(p, c, args)
}

func executeCommand(p IPlugin, c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {

	serverConfiguration := p.GetConfig()
	user, err := p.GetUser(args.UserId)
	if err != nil {
		fmt.Println(err.Error())
		return nil, &model.AppError{Message: err.Error()}
	} else if user == nil {
		fmt.Println("User is nil")
		return nil, &model.AppError{Message: "User is nil"}
	}

	parsedArgs, e := parseArgs(args.Command, CurrentDate{Value: time.Now()})
	if e != nil {
		return nil, &model.AppError{Message: "Invalid arguments"}
	}

	newMeetingResponse, err2 := p.CreateNewMeeting(user)
	if err2 != nil {
		return nil, &model.AppError{Message: err2.Error()}
	}

	post := &model.Post{
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   parsedArgs.MeetingName,
		Type:      POST_MEETING_TYPE,
		Props: map[string]interface{}{
			"meeting_id":        newMeetingResponse.MeetingId,
			"meeting_link":      newMeetingResponse.JoinUrl,
			"meeting_personal":  false,
			"meeting_topic":     parsedArgs.MeetingName,
			"override_username": POST_MEETING_OVERRIDE_USERNAME,
			"meeting_status":    "SCHEDULED",
			"from_webhook":      "true",
			"start_time":        parsedArgs.StartTime.String(),
			"end_time":          parsedArgs.EndTime.String(),
			"override_icon_url": path.Join(*serverConfiguration.ServiceSettings.SiteURL, "plugins", manifest.ID, "api", "v1", "assets", "profile.png"),
		},
	}

	if _, err := p.CreatePost(post); err != nil {
		fmt.Println(err.Error())
		return nil, &model.AppError{Message: err.Error()}
	}

	return getCommandResponse(model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, user.Username, "testtext"), nil
}

type ParsedArgs struct {
	MeetingName string
	StartTime   time.Time
	EndTime     time.Time
}

func parseArgs(args string, currDate CurrentDate) (*ParsedArgs, error) {
	re := regexp.MustCompile(`"([^"]*)"`)
	match := re.FindAllString(args, -1)

	parsedArgs := ParsedArgs{}
	parsedArgs.MeetingName = strings.Trim(match[0], "\"")
	dateOfMeeting, e := time.Parse("2006-01-02", strings.ToUpper(strings.Trim(match[1], "\"")))
	if e != nil {
		return nil, e
	}

	startTime, e := time.Parse(time.Kitchen, strings.ToUpper(strings.Trim(match[2], "\"")))
	if e != nil {
		return nil, e
	}
	parsedArgs.StartTime = startTime
	endTime, e := time.Parse(time.Kitchen, strings.ToUpper(strings.Trim(match[3], "\"")))
	if e != nil {
		return nil, e
	}
	parsedArgs.EndTime = endTime

	now := currDate.Value
	parsedArgs.StartTime = time.Date(dateOfMeeting.Year(), dateOfMeeting.Month(), dateOfMeeting.Day(), startTime.Hour(), startTime.Minute(),
		0, 0, now.Location())
	parsedArgs.EndTime = time.Date(dateOfMeeting.Year(), dateOfMeeting.Month(), dateOfMeeting.Day(), endTime.Hour(), endTime.Minute(),
		0, 0, now.Location())

	return &parsedArgs, nil
}
