package main

import (
	"fmt"
	"sync"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	botID string

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	p.kvstore = NewKVStore(p.API)

	botID, appErr := p.client.Bot.EnsureBot(&model.Bot{
		Username:    "expensebot",
		DisplayName: "ExpenseBot",
		Description: "Bot for submitting expenses and tracking their status. Talk receipts to me. :kissing_heart:",
	})
	if appErr != nil {
		return fmt.Errorf("failed ensuring bot: %w", appErr)
	}
	p.botID = botID

	config := new(configuration)
	err := p.API.LoadPluginConfiguration(&config)
	if err != nil {
		return fmt.Errorf("failed to load plugin configuration: %w", err)
	}
	p.setConfiguration(config)

	p.API.LogInfo("ExpenseBot plugin activated.")

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	return nil
}
