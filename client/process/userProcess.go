package process

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-chat/client/logger"
	"go-chat/client/model"
	"go-chat/client/utils"
	common "go-chat/common/message"
	"net"
	"os"
)

type UserProcess struct{}

// 登陆成功菜单显示：
func showAfterLoginMenu() {
	logger.Info("\n----------------login succeed!----------------\n")
	logger.Info("\t\tselect what you want to do\n")
	logger.Info("\t\t1. Show all online users\n")
	logger.Info("\t\t2. Send group message\n")
	logger.Info("\t\t3. Point-to-point communication\n")
	logger.Info("\t\t4. Exit\n")
	var key int
	var content string

	fmt.Scanf("%d\n", &key)
	switch key {
	case 1:
		messageProcess := MessageProcess{}
		err := messageProcess.GetOnlineUerList()
		if err != nil {
			logger.Error("Some error occured when get online user list, error: %v\n", err)
		}
		return
	case 2:
		logger.Notice("Say something:\n")
		fmt.Scanf("%s\n", &content)
		currentUser := model.CurrentUser
		messageProcess := MessageProcess{}
		err := messageProcess.SendGroupMessageToServer(0, currentUser.UserName, content)
		if err != nil {
			logger.Error("Some error occured when send data to server: %v\n", err)
		} else {
			logger.Success("Send group message succeed!\n\n")
		}
	case 3:
		var targetUserName, message string

		logger.Notice("Select one friend by user name\n")
		fmt.Scanf("%s\n", &targetUserName)
		logger.Notice("Input message:\n")
		fmt.Scanf("%s\n", &message)

		messageProcess := MessageProcess{}
		conn, err := messageProcess.PointToPointCommunication(targetUserName, model.CurrentUser.UserName, message)
		if err != nil {
			logger.Error("Some error occured when point to point comunication: %v\n", err)
			return
		}

		errMsg := make(chan error)
		go Response(conn, errMsg)
		err = <-errMsg

		if err.Error() != "<nil>" {
			logger.Error("Send message errror: %v\n", err)
		}
	case 4:
		logger.Warn("Exit...\n")
		os.Exit(0)
	default:
		logger.Info("Selected invalied!\n")
	}
}

// 用户登陆
func (up UserProcess) Login(userName, password string) (err error) {
	// connect server
	conn, err := net.Dial("tcp", "localhost:8888")

	if err != nil {
		logger.Error("Connect server error: %v", err)
		return
	}

	var message common.Message
	message.Type = common.LoginMessageType
	// 生成 loginMessage
	var loginMessage common.LoginMessage
	loginMessage.UserName = userName
	loginMessage.Password = password

	// func Marshal(v interface{}) ([]byte, error)
	// 先序列话需要传到服务器的数据
	data, err := json.Marshal(loginMessage)
	if err != nil {
		logger.Error("Some error occured when parse you data, error: %v\n", err)
		return
	}

	// 首先发送数据 data 的长度到服务器端
	// 将一个字符串的长度转为一个表示长度的切片
	message.Data = string(data)
	message.Type = common.LoginMessageType
	data, _ = json.Marshal(message)

	dispatcher := utils.Dispatcher{Conn: conn}
	err = dispatcher.SendData(data)
	if err != nil {
		return
	}

	errMsg := make(chan error)
	go Response(conn, errMsg)
	err = <-errMsg

	if err != nil {
		return
	}

	for {
		showAfterLoginMenu()
	}
}

// 处理用户注册
func (up UserProcess) Register(userName, password, password_confirm string) (err error) {
	if password != password_confirm {
		err = errors.New("Confirm password not match")
		return
	}

	conn, err := net.Dial("tcp", "localhost:8888")

	if err != nil {
		logger.Error("Connect server error: %v", err)
		return
	}

	// 定义消息
	var messsage common.Message

	// 生成 registerMessage
	var registerMessage common.RegisterMessage
	registerMessage.UserName = userName
	registerMessage.Password = password
	registerMessage.PasswordConfirm = password_confirm

	data, err := json.Marshal(registerMessage)
	if err != nil {
		logger.Error("Client soem error: %v\n", err)
	}

	// 构造需要传递给服务器的数据
	messsage.Data = string(data)
	messsage.Type = common.RegisterMessageType

	data, err = json.Marshal(messsage)
	if err != nil {
		logger.Error("RegisterMessage json Marshal error: %v\n", err)
		return
	}

	dispatcher := utils.Dispatcher{Conn: conn}
	err = dispatcher.SendData(data)
	if err != nil {
		logger.Error("Send data erro!\n")
		return
	}

	errMsg := make(chan error)
	go Response(conn, errMsg)
	err = <-errMsg

	return
}
