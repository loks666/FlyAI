package mq

import (
	"github.com/loks666/FlyAI/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// MQ连接对象
var conn *amqp.Connection
var channel *amqp.Channel

// 交换机名称
const exchangeName = "wechat-message"

// Init
// @description: 初始化MQ
func Init() {
	if !config.Conf.Mq.Enable {
		log.Println("未启用MQ")
		return
	}
	// 读取MQ连接配置
	mqUrl := config.Conf.Mq.RabbitMQ.GetURL()
	if mqUrl == "" {
		log.Panicf("MQ配置异常")
	}

	var err error
	// 创建MQ连接
	if conn, err = amqp.Dial(mqUrl); err != nil {
		log.Panicf("RabbitMQ连接失败: %s", err)
	}

	//获取channel
	if channel, err = conn.Channel(); err != nil {
		log.Panicf("打开Channel失败: %s", err)
	}
	log.Println("RabbitMQ连接成功")
	go Receive()
	log.Println("开始监听消息")
}

// Receive
// @description: 接收消息
func Receive() (err error) {
	// 创建交换机
	if err = channel.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		//true表示这个exchange不可以被client用来推送消息，仅用来进行exchange和exchange之间的绑定
		false,
		false,
		nil,
	); err != nil {
		log.Printf("声明Exchange失败: %s", err)
		return
	}
	// 创建队列
	var q amqp.Queue
	q, err = channel.QueueDeclare(
		"", //随机生产队列名称
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Printf("无法声明Queue: %s", err)
		return
	}
	// 绑定队列到 exchange 中
	err = channel.QueueBind(
		q.Name,
		//在pub/sub模式下，这里的key要为空
		"",
		exchangeName,
		false,
		nil)
	if err != nil {
		log.Printf("绑定队列失败: %s", err)
		return
	}

	// 消费消息
	var messages <-chan amqp.Delivery
	messages, err = channel.Consume(
		q.Name,
		"",
		false, // 不自动ack，手动处理，这样即使消费者挂掉，消息也不会丢失
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("无法使用队列: %s", err)
		return
	}
	for {
		msg, ok := <-messages
		if !ok {
			log.Printf("获取消息失败")
			return Receive()
		}
		log.Printf("收到消息: %s", msg.Body)
		parse(msg.Body)
		// ACK消息
		if err = msg.Ack(true); err != nil {
			log.Printf("ACK消息失败: %s", err)
			continue
		}
	}

}
