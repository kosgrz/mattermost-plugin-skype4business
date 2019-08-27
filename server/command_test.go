package main

import (
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {

	siteURL := "test.com"

	p := PluginMock{}
	p.On("GetConfig").Return(&model.Config{
		ServiceSettings: model.ServiceSettings{
			SiteURL: &siteURL,
		},
	}, (*model.AppError)(nil))
	p.On("GetUser", "testuserid").
		Return(&model.User{Username: "testusername"}, nil)
	p.On("CreateNewMeeting", &model.User{
		Username: "testusername",
	}).Return(&NewMeetingResponse{
		JoinUrl:   "testurl",
		MeetingId: "testid",
	}, nil)
	p.On("CreatePost", &model.Post{
		UserId:    "testuserid",
		ChannelId: "testchannelid",
		Message:   "test custom meeting name",
		Type:      "custom_s4b",
		Props: model.StringInterface{
			"from_webhook":      "true",
			"meeting_id":        "testid",
			"meeting_link":      "testurl",
			"meeting_personal":  false,
			"meeting_status":    "SCHEDULED",
			"meeting_topic":     "test custom meeting name",
			"start_time":        "0000-01-01 21:15:00 +0000 UTC",
			"end_time":          "0000-01-01 21:30:00 +0000 UTC",
			"override_icon_url": "test.com/plugins/skype4business/api/v1/assets/profile.png",
			"override_username": "Skype for Business Plugin",
		},
	}).Return(&model.Post{}, (*model.AppError)(nil))

	r, err := executeCommand(&p, &plugin.Context{}, &model.CommandArgs{
		UserId:    "testuserid",
		ChannelId: "testchannelid",
		Command:   "/s4b \"test custom meeting name\" \"9:15pm\" \"9:30pm\"",
	})

	assert.NotNil(t, r)
	assert.Equal(t, r.ResponseType, model.COMMAND_RESPONSE_TYPE_IN_CHANNEL)
	assert.Equal(t, r.Text, "testtext")
	assert.Equal(t, r.Username, "testusername")
	assert.Equal(t, r.Type, POST_MEETING_TYPE)
	assert.Nil(t, err)
}

func TestParsingArgs(t *testing.T) {
	testLocation, _ := time.LoadLocation("Asia/Shanghai")
	testCurrDate := time.Date(2010, 12, 10, 0, 0, 0, 0, testLocation)

	testArgs := "/s4b \"test name\" \"8:30am\" \"9:00am\""

	parsedArgs, e := parseArgs(testArgs, CurrentDate{Value: testCurrDate})

	assert.NotNil(t, parsedArgs)
	assert.Equal(t, "test name", parsedArgs.MeetingName)
	assert.Equal(t, "2010-12-10 08:30:00 +0800 CST", parsedArgs.StartTime.String())
	assert.Equal(t, "2010-12-10 09:00:00 +0800 CST", parsedArgs.EndTime.String())
	assert.Nil(t, e)
}

type PluginMock struct {
	mock.Mock
}

func (p *PluginMock) GetConfig() *model.Config {
	ret := p.Called()
	return ret.Get(0).(*model.Config)
}

func (p *PluginMock) GetUser(userId string) (*model.User, *model.AppError) {
	ret := p.Called(userId)

	if ret.Get(0) != nil && ret.Get(1) == nil {
		return ret.Get(0).(*model.User), nil
	} else {
		return nil, ret.Get(1).(*model.AppError)
	}
}

func (p *PluginMock) CreateNewMeeting(user *model.User) (*NewMeetingResponse, error) {
	ret := p.Called(user)

	if ret.Get(0) != nil && ret.Get(1) == nil {
		return ret.Get(0).(*NewMeetingResponse), nil
	} else {
		return nil, ret.Get(1).(*model.AppError)
	}
}

func (p *PluginMock) CreatePost(post *model.Post) (*model.Post, *model.AppError) {
	ret := p.Called(post)

	if ret.Get(0) != nil && ret.Get(1) == nil {
		return ret.Get(0).(*model.Post), nil
	} else {
		return nil, ret.Get(1).(*model.AppError)
	}
}
