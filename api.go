package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/djavorszky/ddn-api/registry"
	"github.com/djavorszky/ddn-common/errs"
	"github.com/djavorszky/ddn-common/inet"
	"github.com/djavorszky/ddn-common/logger"
	"github.com/djavorszky/ddn-common/model"
	"github.com/djavorszky/ddn-common/status"
	"github.com/djavorszky/sutils"
)

func apiPage(w http.ResponseWriter, r *http.Request) {
	api := fmt.Sprintf("%s/web/html/api.html", workdir)

	t, err := template.ParseFiles(api)
	if err != nil {
		logger.Error("Failed parsing files: %v", err)
		return
	}

	t.Execute(w, nil)
}

func apiSafe2Restart(w http.ResponseWriter, r *http.Request) {
	imports := make(map[string]int)

	// Check if server and agents are restartable
	entries, err := db.FetchAll()
	if err != nil {
		logger.Error("failed FetchAll: %v", err)
		msg := inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.QueryFailed,
		}

		inet.SendResponse(w, http.StatusInternalServerError, msg)
		return
	}

	for _, entry := range entries {
		if entry.InProgress() {
			imports[entry.AgentName]++
		}
	}

	result := inet.MapMessage{Status: http.StatusOK, Message: make(map[string]string)}

	conns := registry.List()

	if len(imports) == 0 {
		result.Message["server"] = "yes"

		for _, c := range conns {
			result.Message[c.ShortName] = "yes"
		}

		inet.SendResponse(w, http.StatusOK, result)
		return
	}

	result.Message["server"] = "no"

	for _, c := range conns {
		if imports[c.ShortName] == 0 {
			result.Message[c.ShortName] = "yes"
			continue
		}

		result.Message[c.ShortName] = fmt.Sprintf("No, %d imports running", imports[c.ShortName])
	}

	inet.SendResponse(w, http.StatusOK, result)
}

// stores a web push notification subscription to the database
func apiSaveSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		subscription model.PushSubscription
		err          error
	)

	err = json.NewDecoder(r.Body).Decode(&subscription)
	if err != nil {
		logger.Error("couldn't decode json request: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.JSONDecodeFailed,
		})
		return
	}

	if ok := sutils.Present(subscription.Endpoint, subscription.Keys.P256dh, subscription.Keys.Auth); !ok {
		logger.Error("Missing or empty subscription parameters were received from the /api/save-subscription API call!")
		//TODO
		// log the received request body
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingParameters,
		})
		return
	}

	userCookie, err := r.Cookie("user")
	if err != nil {
		logger.Error("getting user cookie failed: %v", err)
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingUserCookie,
		})
		return
	}

	err = db.InsertPushSubscription(&subscription, userCookie.Value)
	if err != nil {
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.PersistFailed,
		})
		return
	}

	msg := inet.Message{Status: status.Success, Message: fmt.Sprintf("Subscription has been saved to back end.")}

	inet.SendResponse(w, http.StatusOK, msg)
}

// removes a web push notification subscription from the database
func apiRemoveSubscription(w http.ResponseWriter, r *http.Request) {
	var (
		subscription model.PushSubscription
		err          error
	)

	err = json.NewDecoder(r.Body).Decode(&subscription)
	if err != nil {
		logger.Error("couldn't decode json request: %v", err)

		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.JSONDecodeFailed,
		})
		return
	}

	if ok := sutils.Present(subscription.Endpoint, subscription.Keys.P256dh, subscription.Keys.Auth); !ok {
		logger.Error("Missing or empty subscription parameters were received from the /api/remove-subscription API call!")
		//TODO
		// log the received request body
		inet.SendResponse(w, http.StatusBadRequest, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingParameters,
		})
		return
	}

	userCookie, err := r.Cookie("user")
	if err != nil {
		logger.Error("getting user cookie failed: " + err.Error())
		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusBadRequest,
			Message: errs.MissingUserCookie,
		})
		return
	}

	err = db.DeletePushSubscription(&subscription, userCookie.Value)
	if err != nil {
		logger.Error("failed deleting push subscription: %v", err)

		inet.SendResponse(w, http.StatusInternalServerError, inet.Message{
			Status:  http.StatusInternalServerError,
			Message: errs.DropFailed,
		})
		return
	}

	msg := inet.Message{Status: status.Success, Message: fmt.Sprintf("Subscription has been removed from back end.")}

	inet.SendResponse(w, http.StatusOK, msg)
}
