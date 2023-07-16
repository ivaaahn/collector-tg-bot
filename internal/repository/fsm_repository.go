package repo

import (
	"collector/internal"
	"fmt"
)

type UserContext struct {
	Stage    string
	Data     map[string]string
	Endpoint string
}

func newUseCaseData(stage string, data map[string]string, endpoint string) *UserContext {
	return &UserContext{Data: data, Stage: stage, Endpoint: endpoint}
}

type FSMRepository struct {
	log      internal.Logger
	Contexts map[int64]*UserContext
}

func NewFSMRepository(log internal.Logger) *FSMRepository {
	return &FSMRepository{log, make(map[int64]*UserContext)}
}

func (r *FSMRepository) Init(userID int64, endpoint string) {
	r.Contexts[userID] = newUseCaseData("", make(map[string]string), endpoint)
	r.log.Infof("Init for user %q – endpoint %q", userID, endpoint)
}

func (r *FSMRepository) Upsert(userID int64, data map[string]string, newStage string) {
	userContext, ok := r.Contexts[userID]
	if !ok {
		r.log.Warnf("Context not initialized... huh?")
		return
	}

	userContext.Stage = newStage
	for k, v := range data {
		userContext.Data[k] = v
	}

	fmt.Printf("%v\n", userContext)
}

func (r *FSMRepository) GetContext(userID int64) *UserContext {
	return r.Contexts[userID]
}

func (r *FSMRepository) IsActiveSessionExists(userID int64) bool {
	session, exists := r.Contexts[userID]
	if !exists {
		return false
	}

	if session.Stage == "" {
		return false
	}

	return true
}

func (r *FSMRepository) Clear(userID int64) {
	delete(r.Contexts, userID)
	fmt.Println(r.Contexts)
}
