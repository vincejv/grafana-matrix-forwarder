package forwarder

import (
	"encoding/json"
	"grafana-matrix-forwarder/cfg"
	"grafana-matrix-forwarder/matrix"
	"grafana-matrix-forwarder/model"
	"io/ioutil"
	"log"
	"os"
)

const (
	alertMapFileName = "grafanaToMatrixMap.json"
)

type AlertForwarder struct {
	AppSettings                cfg.AppSettings
	Writer                     matrix.Writer
	alertToSentEventMap        map[string]sentMatrixEvent
	alertMapPersistenceEnabled bool
}

type sentMatrixEvent struct {
	EventID           string
	SentFormattedBody string
}

func NewForwarder(appSettings cfg.AppSettings, writer matrix.Writer) *AlertForwarder {
	forwarder := &AlertForwarder{
		AppSettings:                appSettings,
		Writer:                     writer,
		alertToSentEventMap:        map[string]sentMatrixEvent{},
		alertMapPersistenceEnabled: appSettings.PersistAlertMap,
	}
	forwarder.prePopulateAlertMap()
	return forwarder
}

// ForwardAlertToRooms sends the provided grafana.AlertPayload to the provided matrix.Writer by iterating over all the provided room IDs.
func (forwarder *AlertForwarder) ForwardAlertToRooms(roomIDs []string, alert model.Data) error {
	for _, roomID := range roomIDs {
		err := forwarder.ForwardAlertToRoom(roomID, alert)
		if err != nil {
			return err
		}
	}
	return nil
}

// ForwardAlertToRoom sends the provided grafana.AlertPayload to the provided matrix.Writer using the provided roomID
func (forwarder *AlertForwarder) ForwardAlertToRoom(roomID string, alert model.Data) (err error) {
	resolveWithReaction := forwarder.AppSettings.ResolveMode == cfg.ResolveWithReaction
	resolveWithReply := forwarder.AppSettings.ResolveMode == cfg.ResolveWithReply

	if sentEvent, ok := forwarder.alertToSentEventMap[alert.Id]; ok {
		if alert.State == model.AlertStateResolved && resolveWithReaction {
			delete(forwarder.alertToSentEventMap, alert.Id)
			return forwarder.sendReaction(roomID, sentEvent.EventID)
		}
		if alert.State == model.AlertStateResolved && resolveWithReply {
			delete(forwarder.alertToSentEventMap, alert.Id)
			return forwarder.sendReply(roomID, sentEvent)
		}
	}
	return forwarder.sendRegularMessage(roomID, alert, alert.Id)
}

func (forwarder *AlertForwarder) sendReaction(roomID string, eventID string) (err error) {
	_, err = forwarder.Writer.React(roomID, eventID, resolvedReactionStr)
	return
}

func (forwarder *AlertForwarder) sendReply(roomID string, event sentMatrixEvent) (err error) {
	replyMessageBody, err := executeTextTemplate(resolveReplyTemplate, event.SentFormattedBody)
	if err != nil {
		return
	}
	_, err = forwarder.Writer.Reply(roomID, event.EventID, resolveReplyPlainStr, replyMessageBody)
	return
}

func (forwarder *AlertForwarder) sendRegularMessage(roomID string, alert model.Data, alertID string) (err error) {
	formattedMessageBody, err := buildFormattedMessageBodyFromAlert(alert, forwarder.AppSettings)
	if err != nil {
		return
	}
	plainMessageBody := stripHtmlTagsFromString(formattedMessageBody)
	response, err := forwarder.Writer.Send(roomID, plainMessageBody, formattedMessageBody)
	if err == nil {
		forwarder.alertToSentEventMap[alertID] = sentMatrixEvent{
			EventID:           response.EventID.String(),
			SentFormattedBody: formattedMessageBody,
		}
		forwarder.persistAlertMap()
	}
	return
}

func (forwarder *AlertForwarder) prePopulateAlertMap() {
	fileData, err := ioutil.ReadFile(alertMapFileName)
	if err == nil {
		err = json.Unmarshal(fileData, &forwarder.alertToSentEventMap)
	}

	if err != nil {
		log.Printf("failed to load alert map - falling back on an empty map (%v)", err)
	}
}

func (forwarder *AlertForwarder) persistAlertMap() {
	if !forwarder.alertMapPersistenceEnabled {
		return
	}

	jsonData, err := json.Marshal(forwarder.alertToSentEventMap)
	if err == nil {
		err = ioutil.WriteFile(alertMapFileName, jsonData, os.ModePerm)
	}

	if err != nil {
		log.Printf("failed to persist alert map - functionality disabled (%v)", err)
		forwarder.alertMapPersistenceEnabled = false
	}
}