// HTTP Controller for a MessageBox service
// This contains all directly related http responses as well direct logic control with the database

package controller

import (
	"MessageBox/dataservice"
	"MessageBox/model"
	"MessageBox/util/logger"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"time"
)

type MessageBox struct {}

// CreateUser godoc
// @Summary Creates a new user
// @Description Create user by username
// @Accept  json
// @Produce  json
// @Param UserRegistration body model.UserRegistration true "User registration"
// @Success 200 {object} model.UserRegistration
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /users [post]
func (mb *MessageBox) CreateUser(ctx echo.Context) error {
	user := new(model.UserRegistration)
	if err := ctx.Bind(user); err != nil {
		return err
	}
	if user.UserName == "" {
		return ctx.JSON(http.StatusBadRequest, "Bad request")
	}
	result, err := dataservice.CB.GetUser(user.UserName)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	if result != nil {
		return ctx.JSON(http.StatusConflict, user)
	}
	result, err = dataservice.CB.Insert(user)
	if err != nil {
		return err
	}
	logger.Log.Debug("Insert user success: %v")
	return ctx.JSON(http.StatusCreated, user)
}

// GetUserMessages godoc
// @Summary Get the mailbox for a username
// @Description Grab all messages in a user mailbox
// @Accept  json
// @Produce  json
// @Param username path string true "username"
// @Success 200 {object} []model.Message
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /users/{username}/mailbox [get]
func (mb *MessageBox) GetUserMessages(ctx echo.Context) error {
	user := ctx.Param("username")
	if user == "" {
		return ctx.JSON(http.StatusNotFound, fmt.Sprintf("User: %v, not found", user))
	}
	result, err := dataservice.CB.GetUser(user)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	if result == nil {
		return ctx.JSON(http.StatusNotFound, "Not found")
	}
	userObj := model.User{}
	userString, _ := json.Marshal(result)
	if err := json.Unmarshal(userString, &userObj); err != nil {
		return err
	}

	//Sort by RFC3339
	sort.Slice(userObj.Messages, func(i, j int) bool {
		t1, _ := time.Parse(time.RFC3339, userObj.Messages[i].SentAt)
		t2, _ := time.Parse(time.RFC3339, userObj.Messages[j].SentAt)
		return t1.After(t2)
	})
	return ctx.JSON(http.StatusOK, userObj.Messages)
}

// ReplyMessage godoc
// @Summary Reply to a message!
// @Description A user can reply to a message based on the message Id
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Param ReplyMessage body model.ReplyMessage true "Reply Message"
// @Success 200 {object} model.Message
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /messages/{id}/replies [post]
func (mb *MessageBox) ReplyMessage(ctx echo.Context) error {
	messageId := ctx.Param("id")
	if messageId == "" {
		return ctx.JSON(http.StatusNotFound,fmt.Sprintf("ID: %v, not found", messageId))
	}
	idCast, err := strconv.Atoi(messageId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}

	// Get the message we are replying to by ID
	result, err := dataservice.CB.GetMessage(idCast, false)
	if result != nil {

		//Create a new reply message object and bind it to the request
		replyMessage := new(model.ReplyMessage)
		if err := ctx.Bind(replyMessage); err != nil {
			return err
		}

		messageObj := model.Message{}
		messageString, _ := json.Marshal(result)
		var msgResponse interface{}
		if err := json.Unmarshal(messageString, &messageObj); err != nil {
			return err
		}
		if _, ok := mb.getRecipientM(&messageObj).(model.UserRegistration); ok {
			msgResponse = mb.userReply(*replyMessage, messageObj.Recipient.UserName, messageObj.Id)
		}
		if _, ok := mb.getRecipientM(&messageObj).(model.GroupRecipient); ok {
			msgResponse = mb.groupReply(*replyMessage, messageObj.Recipient.GroupName, messageObj.Id)
		}
		if msgResponse == nil {
			return ctx.JSON(http.StatusInternalServerError, "Error with a reply!")
		}
		return ctx.JSON(http.StatusAccepted, msgResponse)
	}
	return ctx.JSON(http.StatusNotFound, "Message Not found!")
}

// CreateGroups godoc
// @Summary Create a new group
// @Description Create a new group defined with a groupname and usernames to be added
// @Accept  json
// @Produce  json
// @Param GroupCreation body model.GroupCreation true "group registration"
// @Success 200 {object} model.GroupCreation
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /groups [post]
func (mb *MessageBox) CreateGroups(ctx echo.Context) error {
	group := new(model.GroupCreation)
	if err := ctx.Bind(group); err != nil {
		return err
	}
	if group.Groupname == "" {
		return ctx.JSON(http.StatusBadRequest, "Bad Request")
	}
	result, err := dataservice.CB.GetGroup(group.Groupname)
	if err != nil {
		return err
	}
	if result != nil {
		return ctx.JSON(http.StatusConflict, result)
	}
	result, err = dataservice.CB.Insert(group)
	if err != nil {
		return err
	}
	logger.Log.Debugf("Group %v creation success", group.Groupname)
	return ctx.JSON(http.StatusCreated, group)
}

// SendMessage godoc
// @Summary Send a message to a user
// @Description Send a message to a user or group
// @Accept  json
// @Produce  json
// @Param ComposedMessage body model.ComposedMessage true "A composed message"
// @Success 200 {object} model.ComposedMessage
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /messages [post]
func (mb *MessageBox) SendMessage(ctx echo.Context) error {
	message := new(model.ComposedMessage)
	if err := ctx.Bind(message); err != nil {
		return err
	}

	//Type validation that we are
	if _, ok := mb.getRecipientC(message).(model.UserRegistration); ok {
		username := message.Recipient.UserName
		result, err := dataservice.CB.GetUser(username)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, err)
		}
		if result != nil {
			messageObj, err := mb.updateUserWithMessage(result.(map[string]interface{}), *message, false)
			if err != nil {
				return ctx.JSON(http.StatusInternalServerError, err)
			}
			return ctx.JSON(http.StatusAccepted, messageObj)
		} else {
			return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Recipient: %v , not found", username))
		}
	}

	if _, ok := mb.getRecipientC(message).(model.GroupRecipient); ok {
		groupname := message.Recipient.GroupName
		result, err := dataservice.CB.GetGroup(groupname)
		if err != nil {
			return err
		}
		if result != nil {
			msgResponse := mb.sendMessageToGroup(message, groupname, false)
			return ctx.JSON(http.StatusAccepted, msgResponse)
		}
	}
	return ctx.JSON(http.StatusBadRequest, "Bad Request")
}

// GetReplies godoc
// @Summary Get all the replies from a message
// @Description Get a reply from a message by message ID
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} []model.Message
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /messages/{id}/replies [get]
func (mb *MessageBox) GetReplies(ctx echo.Context) error {
	msgId := ctx.Param("id")
	if msgId == "" {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("Bad Request"))
	}
	idCast, err := strconv.Atoi(msgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	result, err := dataservice.CB.GetMessage(idCast, true)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	if result == nil {
		return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Msg ID: %v, not found", msgId))
	}
	var msgArr []interface{}
	resultCast := result.([]interface{})

	//Reply messages will have RE key. We just unmarshal the message and check for a RE to append to the response.
	//This algorithm could be improved by a more powerful N1QL query which i didn't have time for :)
	if len(resultCast) > 1 {
		for _, elm := range resultCast {
			msgObj := model.Message{}
			msgString, _ := json.Marshal(elm)
			if err := json.Unmarshal(msgString, &msgObj); err != nil {
				return err
			}
			if msgObj.Re == 0 {
				continue
			}
			msgArr = append(msgArr, msgObj)
		}
	} else {
		msgObj := model.Message{}
		msgString, _ := json.Marshal(result)
		if err := json.Unmarshal(msgString, &msgObj); err != nil {
			return err
		}
		if msgObj.Re != 0 {
			msgArr = append(msgArr, msgObj)
		}
	}
	if msgArr == nil {
		return ctx.JSON(http.StatusNotFound, "Not Found")
	}
	return ctx.JSON(http.StatusOK, msgArr)
}

// GetMessage godoc
// @Summary Get a message
// @Description Get a message by ID
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} model.Message
// @Failure 400,404 {object} api.HTTPError
// @Failure 500 {object} api.HTTPError
// @Failure default {object} api.HTTPError
// @Router /messages/{id} [get]
func (mb *MessageBox) GetMessage(ctx echo.Context) error {
	msgId := ctx.Param("id")
	if msgId == "" {
		return ctx.JSON(http.StatusBadRequest, fmt.Sprintf("Bad Request"))
	}
	idCast, err := strconv.Atoi(msgId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	result, err := dataservice.CB.GetMessage(idCast, false)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	if result == nil {
		return ctx.JSON(http.StatusNotFound, fmt.Sprintf("Msg ID: %v, not found", msgId))
	}
	msgObj := model.Message{}
	msgString, _ := json.Marshal(result)
	if err := json.Unmarshal(msgString, &msgObj); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, msgObj)
}


// getRecipientC
// Validate and return whether a Composed message is for a user or group
func (mb *MessageBox) getRecipientC(message *model.ComposedMessage) interface{} {
	if message.Recipient.UserName != "" {
		return model.UserRegistration{UserName: message.Recipient.UserName}
	}
	if message.Recipient.GroupName != "" {
		return model.GroupRecipient{Groupname: message.Recipient.GroupName}
	}
	return nil
}

// getRecipientM
// Validate and return whether a Message is for a user or group
func (mb *MessageBox) getRecipientM(message *model.Message) interface{} {
	if message.Recipient.UserName != "" {
		return model.UserRegistration{UserName: message.Recipient.UserName}
	}
	if message.Recipient.GroupName != "" {
		return model.GroupRecipient{Groupname: message.Recipient.GroupName}
	}
	return nil
}

// updateUserWithMessage
// Takes a cb response object for a user, and a composed message, then updates the user doc in the key/value store
func (mb *MessageBox) updateUserWithMessage(cbR map[string]interface{}, message model.ComposedMessage, reply bool) (interface{}, error) {
	userObj, err := mb.marshalUserObj(cbR)
	if err != nil {
		return nil, err
	}

	messageObj, err := mb.createNewMessageObj(message, reply)
	if err != nil {
		return nil, err
	}

	userObj.Messages = append(userObj.Messages, *messageObj)
	logger.Log.Debugf("New message added to user: %v, %v", userObj.UserName, messageObj)

	var docId = cbR["id"].(string)
	if _, err := dataservice.CB.Update(docId, userObj); err != nil {
		return nil, err
	}
	return messageObj, nil
}

// marshalUserObj
// simple offshore method to marshal a user object and return the address to it
func (mb *MessageBox) marshalUserObj(obj interface{}) (*model.User, error) {
	userObj := model.User{}
	userString, _ := json.Marshal(obj)
	if err := json.Unmarshal(userString, &userObj); err != nil {
		return nil, err
	}
	return &userObj, nil
}

// createNewMessageObj
// similar too marshalUserObj except it will validate a ID and set a new time field
func (mb *MessageBox) createNewMessageObj(message model.ComposedMessage, reply bool) (*model.Message, error){
	messageObj := model.Message{}
	messageString, _ := json.Marshal(message)
	if err := json.Unmarshal(messageString, &messageObj); err != nil {
		return nil, err
	}
	messageObj.SentAt = time.Now().Format(time.RFC3339)
	if messageObj.Id == 0 {
		messageObj.Id = rand.Intn(500) + 1
	}
	if reply {
		messageObj.Re = rand.Intn(500) + 1
	}
	return &messageObj, nil
}

// sendMessageToGroup
// takes a composed message, and the group name and will update all the users in the group with the new message
func (mb *MessageBox) sendMessageToGroup(message *model.ComposedMessage, groupname string , reply bool) interface{} {
	result, err := dataservice.CB.GetUsersInGroup(groupname)
	if err != nil {
		return err
	}
	var msgResponse *model.Message
	for _, elm := range result.([]interface{}) {
		userObj, _ := elm.(map[string]interface{})
		msg, err := mb.updateUserWithMessage(userObj, *message, reply)
		if err != nil {
			return err
		}
		if _, ok := msg.(*model.Message); ok == true {
			msgResponse = msg.(*model.Message)
		}
	}
	return msgResponse
}

// userReply
// replies a message directly to a user
func (mb *MessageBox) userReply(reply model.ReplyMessage, username string, messageId int) interface{} {

	//Create a new message for the recipient they are replying to
	message := model.ComposedMessage{
		Sender: reply.Sender,
		Subject: reply.Subject,
		Body: reply.Body,
		Recipient: model.MessageRecipient{
			UserName: username,
			GroupName: "",
		},
		Id: messageId,
	}
	//Get the user information for the recipient on the replying end and add to their inbox
	cbR, err := dataservice.CB.GetUser(username)
	if err != nil {
		return err
	}
	msgResponse, err := mb.updateUserWithMessage(cbR.(map[string]interface{}), message, true)
	if err != nil {
		return err
	}
	return msgResponse
}

// groupReply
// replies a message to a group of users.
func (mb *MessageBox) groupReply(reply model.ReplyMessage, groupname string, messageId int) interface{} {
	message := model.ComposedMessage{
		Sender: reply.Sender,
		Subject: reply.Subject,
		Body: reply.Body,
		Recipient: model.MessageRecipient{
			UserName: "",
			GroupName: groupname,
		},
		Id: messageId,
	}
	msgResponse := mb.sendMessageToGroup(&message, groupname, true)
	return msgResponse
}