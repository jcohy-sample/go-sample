package process

import (
	"com/jcohy/sample/chatroom/client/utils"
	message "com/jcohy/sample/chatroom/common"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type UserProcess struct {
}

func (this *UserProcess) Register(userId int, userPwd string, userName string) (err error) {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("net.Dial err=", err)
		return
	}

	defer conn.Close()
	// 1、声明消息类型
	var mes message.Message
	mes.Type = message.RegisterMesType

	// 2、填充消息内容
	var registerMes message.RegisterMes
	registerMes.User.UserId = userId
	registerMes.User.UserName = userName
	registerMes.User.UserPwd = userPwd

	// 3、序列化
	data, err := json.Marshal(registerMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	// 4、 把 data 赋给 mes.Data 字段
	mes.Data = string(data)

	// 5、将 mes 进行序列化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	// 6、创建一个 Transfer 实例
	tf := &utils.Transfer{
		Conn: conn,
	}

	// 7、发送数据
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("注册发送信息错误 err=", err)
	}

	// 8、接收数据 mes 就是 RegisterResMes
	mes, err = tf.ReadPkg()
	if err != nil {
		fmt.Println("readPkg(conn) err=", err)
		return
	}

	// 9、反序列化
	var registerResMes message.RegisterResMes
	err = json.Unmarshal([]byte(mes.Data), &registerResMes)
	if registerResMes.Code == 200 {
		fmt.Println("注册成功, 你重新登录一把")
		os.Exit(0)
	} else {
		fmt.Println(registerResMes.Error)
		os.Exit(0)
	}
	return
}

func (this *UserProcess) Login(userId int, userPwd string) (err error) {
	conn, err := net.Dial("tcp", "localhost:8888")
	if err != nil {
		fmt.Println("net.Dial err=", err)
		return
	}
	defer conn.Close()

	//2. 准备通过conn发送消息给服务
	var mes message.Message
	mes.Type = message.LoginMesType
	//3. 创建一个LoginMes 结构体
	var loginMes message.LoginMes
	loginMes.UserId = userId
	loginMes.UserPwd = userPwd

	//4. 将loginMes 序列化
	data, err := json.Marshal(loginMes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}
	// 5. 把data赋给 mes.Data字段
	mes.Data = string(data)

	// 6. 将 mes进行序列化化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err=", err)
		return
	}

	// 7. 到这个时候 data就是我们要发送的消息
	// 7.1 先把 data的长度发送给服务器
	// 先获取到 data的长度->转成一个表示长度的byte切片
	var pkgLen uint32
	pkgLen = uint32(len(data))
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[0:4], pkgLen)
	// 发送长度
	n, err := conn.Write(buf[:4])
	if n != 4 || err != nil {
		fmt.Println("conn.Write(bytes) fail", err)
		return
	}

	fmt.Printf("客户端，发送消息的长度=%d 内容=%s", len(data), string(data))

	// 发送消息本身
	_, err = conn.Write(data)
	if err != nil {
		fmt.Println("conn.Write(data) fail", err)
		return
	}

	//休眠20
	// time.Sleep(20 * time.Second)
	// fmt.Println("休眠了20..")
	// 这里还需要处理服务器端返回的消息.
	//创建一个Transfer 实例
	tf := &utils.Transfer{
		Conn: conn,
	}
	mes, err = tf.ReadPkg() // mes 就是

	if err != nil {
		fmt.Println("readPkg(conn) err=", err)
		return
	}

	//将mes的Data部分反序列化成 LoginResMes
	var loginResMes message.LoginResMes
	err = json.Unmarshal([]byte(mes.Data), &loginResMes)
	if loginResMes.Code == 200 {
		//初始化CurUser
		CurUser.Conn = conn
		CurUser.UserId = userId
		CurUser.UserStatus = message.UserOnline

		//fmt.Println("登录成功")
		//可以显示当前在线用户列表,遍历loginResMes.UsersId
		fmt.Println("当前在线用户列表如下:")
		for _, v := range loginResMes.UsersId {
			//如果我们要求不显示自己在线,下面我们增加一个代码
			if v == userId {
				continue
			}

			fmt.Println("用户id:\t", v)
			//完成 客户端的 onlineUsers 完成初始化
			user := &message.User{
				UserId:     v,
				UserStatus: message.UserOnline,
			}
			onlineUsers[v] = user
		}
		fmt.Print("\n\n")

		//这里我们还需要在客户端启动一个协程
		//该协程保持和服务器端的通讯.如果服务器有数据推送给客户端
		//则接收并显示在客户端的终端.
		go serverProcessMes(conn)

		//1. 显示我们的登录成功的菜单[循环]..
		for {
			ShowMenu()
		}

	} else {
		fmt.Println(loginResMes.Error)
	}

	return
}
