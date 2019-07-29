package main

import (
	"fmt"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"path"
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

func (p *Plugin) getCommandResponse(responseType, text string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: responseType,
		Text:         text,
		Username:     "test",
		Type:         POST_MEETING_TYPE,
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {

	serverConfiguration := p.API.GetConfig()

	post := &model.Post{
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   "Meeting started at %s.",
		Type:      POST_MEETING_TYPE,
		Props: map[string]interface{}{
			"meeting_id":        "test",
			"meeting_link":      "test",
			"meeting_personal":  "test",
			"meeting_topic":     "test",
			"override_username": POST_MEETING_OVERRIDE_USERNAME,
			"meeting_status":    "STARTED",
			"from_webhook":      "true",
			"override_icon_url": path.Join(*serverConfiguration.ServiceSettings.SiteURL, "plugins", manifest.ID, "api", "v1", "assets", "profile.png"),
		},
	}

	if _, err := p.API.CreatePost(post); err != nil {
		fmt.Println(err.Error())
	}

	return p.getCommandResponse(model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, "test"), nil
}
