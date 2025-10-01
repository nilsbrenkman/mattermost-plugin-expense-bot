package main

import (
	"encoding/json"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/pkg/errors"
)

type KVStore interface {
	GetUserDefaults(userID string) (*UserDefaults, error)
	SaveUserDefaults(user *UserDefaults) error
	GetDraft(userID string) (*Draft, error)
	SaveDraft(userID string, draft *Draft) error
	DeleteDraft(userID string) error
	GetExpense(expenseID string) (*Expense, error)
	SaveExpense(expense *Expense) error
}

type UserDefaults struct {
	UserID  string `json:"user_id"`
	Account string `json:"bank_account"`
	Name    string `json:"name"`
}

type Draft struct {
	UserID string            `json:"user_id"`
	State  string            `json:"state"`
	Data   map[string]string `json:"data"`
}

type Expense struct {
	ID          string   `json:"id"`
	PostID      string   `json:"post_id"`
	UserID      string   `json:"user_id"`
	State       string   `json:"state"`
	Account     string   `json:"bank_account"`
	Name        string   `json:"name"`
	Amount      string   `json:"amount"`
	Description string   `json:"description"`
	FileIDs     []string `json:"file_ids"`
}

type Store struct {
	api plugin.API
}

func NewKVStore(api plugin.API) KVStore {
	return Store{
		api: api,
	}
}

func (kv Store) GetUserDefaults(userID string) (*UserDefaults, error) {
	userData, err := kv.api.KVGet("user:" + userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get draft")
	}
	if len(userData) == 0 {
		return nil, nil
	}
	var user UserDefaults
	if err := json.Unmarshal(userData, &user); err != nil {
		return nil, errors.Wrap(err, "failed to decode draft json")
	}
	return &user, nil
}

func (kv Store) SaveUserDefaults(user *UserDefaults) error {
	userData, err := json.Marshal(user)
	if err != nil {
		return errors.Wrap(err, "failed to marshal draft")
	}
	appErr := kv.api.KVSet("user:"+user.UserID, userData)
	if appErr != nil {
		return errors.Wrap(err, "failed to store draft")
	}
	return nil
}

func (kv Store) GetDraft(userID string) (*Draft, error) {
	draftData, err := kv.api.KVGet("draft:" + userID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get draft")
	}
	if len(draftData) == 0 {
		return nil, nil
	}
	var draft Draft
	if err := json.Unmarshal(draftData, &draft); err != nil {
		return nil, errors.Wrap(err, "failed to decode draft json")
	}
	return &draft, nil
}

func (kv Store) SaveDraft(userID string, draft *Draft) error {
	draftData, err := json.Marshal(draft)
	if err != nil {
		return errors.Wrap(err, "failed to marshal draft")
	}
	appErr := kv.api.KVSet("draft:"+userID, draftData)
	if appErr != nil {
		return errors.Wrap(err, "failed to store draft")
	}
	return nil
}

func (kv Store) DeleteDraft(userID string) error {
	err := kv.api.KVDelete("draft:" + userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete draft")
	}
	return nil
}

func (kv Store) GetExpense(expenseID string) (*Expense, error) {
	expenseData, appErr := kv.api.KVGet("expense:" + expenseID)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get expense")
	}
	if len(expenseData) == 0 {
		return nil, nil
	}
	var expense Expense
	if err := json.Unmarshal(expenseData, &expense); err != nil {
		return nil, errors.Wrap(err, "failed to decode draft json")
	}
	return &expense, nil
}

func (kv Store) SaveExpense(expense *Expense) error {
	expenseData, err := json.Marshal(expense)
	if err != nil {
		return errors.Wrap(err, "failed to marshal draft")
	}
	appErr := kv.api.KVSet("expense:"+expense.ID, expenseData)
	if appErr != nil {
		return errors.Wrap(err, "failed to store expense")
	}
	return nil
}
