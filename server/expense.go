package main

import (
	"fmt"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

func (p *Plugin) createExpense(userID string, draft *Draft) error {
	expense := &Expense{
		ID:          model.NewId(),
		UserID:      draft.UserID,
		State:       ExpenseStateSubmitted,
		Account:     draft.Data["iban"],
		Name:        draft.Data["name"],
		Amount:      draft.Data["amount"],
		Description: draft.Data["description"],
		FileIDs:     []string{draft.Data["file"]},
	}

	message, err := p.formatExpense(expense)
	if err != nil {
		return errors.Wrap(err, "failed to format expense")
	}
	dm := p.sendPinnedDM(userID, message, true)
	if dm == nil {
		return errors.New("failed to create post")
	}

	expense.PostID = dm.Id
	err = p.kvstore.SaveExpense(expense)
	if err != nil {
		return errors.Wrap(err, "failed to save expense")
	}

	return p.sendChannelMessage(expense)
}

func (p *Plugin) formatExpense(expense *Expense) (string, error) {
	var state string
	switch expense.State {
	case ExpenseStateSubmitted:
		state = ":hourglass_flowing_sand: **Submitted**"
	case ExpenseStatePaid:
		state = ":white_check_mark: **Paid**"
	case ExpenseStateRejected:
		state = ":x: **Rejected**"
	}
	file, appErr := p.API.GetFileInfo(expense.FileIDs[0])
	if appErr != nil {
		return "", errors.Wrap(appErr, "failed to get file")
	}
	message := fmt.Sprintf("|Status|%s|\n|-|-|\n|Bank account|%s|\n|Name|%s|\n|Amount|%s|\n|Description|%s|\n|File|[%s](%s)|\n",
		state,
		expense.Account,
		expense.Name,
		expense.Amount,
		expense.Description,
		file.Name,
		fmt.Sprintf("%s/api/v4/files/%s", p.getBaseURL(), file.Id),
	)
	return message, nil
}

func (p *Plugin) getBaseURL() string {
	cfg := p.API.GetConfig()
	if cfg == nil || cfg.ServiceSettings.SiteURL == nil {
		p.API.LogWarn("SiteURL is not configured")
		return ""
	}
	return *cfg.ServiceSettings.SiteURL
}

func (p *Plugin) sendChannelMessage(expense *Expense) error {
	channel, appErr := p.API.GetChannel(p.getConfiguration().ChannelID)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to get channel")
	}
	user, appError := p.API.GetUser(expense.UserID)
	if appError != nil {
		return errors.Wrap(appError, "failed to get user")
	}
	title := fmt.Sprintf("**Expense claim from %s %s**", user.FirstName, user.LastName)
	message, err := p.formatExpense(expense)
	if err != nil {
		return errors.Wrap(err, "failed to format expense")
	}
	post, appErr := p.API.CreatePost(&model.Post{
		UserId:    p.botID,
		ChannelId: channel.Id,
		Message:   fmt.Sprintf("%s\n\n%s", title, message),
	})
	if appErr != nil {
		return errors.Wrap(appErr, "failed to create post")
	}
	actions := []*model.PostAction{
		{
			Id:    "paid",
			Name:  "Paid",
			Type:  model.PostActionTypeButton,
			Style: "success",
			Integration: &model.PostActionIntegration{
				URL: fmt.Sprintf("/plugins/com.mattermost.plugin-expense-bot/api/expenses/%s/%s", expense.ID, ExpenseStatePaid),
			},
		},
		{
			Id:    "reject",
			Name:  "Reject",
			Type:  model.PostActionTypeButton,
			Style: "danger",
			Integration: &model.PostActionIntegration{
				URL: fmt.Sprintf("/plugins/com.mattermost.plugin-expense-bot/api/expenses/%s/%s", expense.ID, ExpenseStateRejected),
			},
		},
	}
	attachment := []*model.SlackAttachment{{
		AuthorName: "",
		Actions:    actions,
	}}
	model.ParseSlackAttachment(post, attachment)
	_, appErr = p.API.UpdatePost(post)
	if appErr != nil {
		return errors.Wrap(appErr, "failed to update post")
	}
	return nil
}
