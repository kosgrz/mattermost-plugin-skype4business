package main

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/mattermost/mattermost-server/plugin/plugintest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommand(t *testing.T) {

	api := &plugintest.API{}
	siteURL := "test.com"
	api.On("GetConfig").Return(&model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}, (*model.AppError)(nil))
	api.On("GetUser", "testuserid").
		Return(&model.User{Username: "testusername"}, (*model.AppError)(nil))
	api.On("CreatePost", &model.Post{
		UserId:    "testuserid",
		ChannelId: "testchannelid",
		Message:   "Meeting scheduled.",
		Type:      "custom_s4b",
		Props: model.StringInterface{
			"from_webhook":      "true",
			"meeting_id":        "test",
			"meeting_link":      "test",
			"meeting_personal":  "test",
			"meeting_status":    "SCHEDULED",
			"meeting_topic":     "test",
			"override_icon_url": "test.com/plugins/skype4business/api/v1/assets/profile.png",
			"override_username": "Skype for Business Plugin",
		},
	}).Return(&model.Post{}, (*model.AppError)(nil))
	p := Plugin{}
	p.SetAPI(api)

	r, err := p.ExecuteCommand(&plugin.Context{}, &model.CommandArgs{UserId: "testuserid", ChannelId: "testchannelid"})

	assert.NotNil(t, r)
	assert.Equal(t, r.ResponseType, model.COMMAND_RESPONSE_TYPE_IN_CHANNEL)
	assert.Equal(t, r.Text, "testtext")
	assert.Equal(t, r.Username, "testusername")
	assert.Equal(t, r.Type, POST_MEETING_TYPE)
	assert.Nil(t, err)
}
