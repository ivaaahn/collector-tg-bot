package fsm

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

type FSM struct {
	log      internal.Logger
	Contexts map[int64]*UserContext
}

func NewFSM(log internal.Logger) *FSM {
	return &FSM{log, make(map[int64]*UserContext)}
}

func (r *FSM) Init(userID int64, endpoint string) {
	r.Contexts[userID] = newUseCaseData("", make(map[string]string), endpoint)
	r.log.Infof("Init for user %q – endpoint %q", userID, endpoint)
}

func (r *FSM) Upsert(userID int64, data map[string]string, newStage string) {
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

func (r *FSM) GetContext(userID int64) *UserContext {
	return r.Contexts[userID]
}

func (r *FSM) IsActiveSessionExists(userID int64) bool {
	session, exists := r.Contexts[userID]
	if !exists {
		return false
	}

	if session.Stage == "" {
		return false
	}

	return true
}

func (r *FSM) Clear(userID int64) {
	delete(r.Contexts, userID)
	fmt.Println(r.Contexts)
}
