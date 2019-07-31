package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"path"
	"strings"
	"time"
)

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

func (p *Plugin) getCommandResponse(responseType, username string, text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     username,
		Type:         POST_MEETING_TYPE,
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {

	serverConfiguration := p.API.GetConfig()
	user, err := p.API.GetUser(args.UserId)
	if err != nil {
		fmt.Println(err.Error())
		return nil, &model.AppError{Message: err.Error()}
	} else if user == nil {
		fmt.Println("User is nil")
		return nil, &model.AppError{Message: "User is nil"}
	}

	post := &model.Post{
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   "Meeting scheduled.",
		Type:      POST_MEETING_TYPE,
		Props: map[string]interface{}{
			"meeting_id":        "test",
			"meeting_link":      "test",
			"meeting_personal":  "test",
			"meeting_topic":     "test",
			"override_username": POST_MEETING_OVERRIDE_USERNAME,
			"meeting_status":    "SCHEDULED",
			"from_webhook":      "true",
			"override_icon_url": path.Join(*serverConfiguration.ServiceSettings.SiteURL, "plugins", manifest.ID, "api", "v1", "assets", "profile.png"),
		},
	}

	if _, err := p.API.CreatePost(post); err != nil {
		fmt.Println(err.Error())
		return nil, &model.AppError{Message: err.Error()}
	}

	return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, user.Username, "testtext"), nil
}

type ParsedArgs struct {
	StartTime time.Time
}

func (p *Plugin) parseArgs(args string) (*ParsedArgs, error) {
	parsedArgs := ParsedArgs{}
	arrayArgs := strings.Split(args, " ")

	if len(arrayArgs) == 3 {
		startTime, e := time.Parse(time.Kitchen, strings.ToUpper(arrayArgs[2]))
		if e != nil {
			return nil, e
		}
		parsedArgs.StartTime = startTime
	}

	return &parsedArgs, nil
}
