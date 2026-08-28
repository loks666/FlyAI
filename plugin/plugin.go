package plugin

import (
	"github.com/loks666/FlyAI/model/dto"
	"hash/fnv"
	"log"
	"os"
	"strconv"
	"sync"
)

// MessageHandler 消息处理函数
type MessageHandler func(msg *dto.Message)

// MessageDispatcher 消息分发处理接口
// 跟 DispatchMessage 结合封装成 MessageHandler
type MessageDispatcher interface {
	Dispatch(msg *dto.Message)
}

// DispatchMessage 跟 MessageDispatcher 结合封装成 MessageHandler
func DispatchMessage(dispatcher MessageDispatcher) func(msg *dto.Message) {
	return func(msg *dto.Message) { dispatcher.Dispatch(msg) }
}

// MessageDispatcher impl

// MessageContextHandler 消息处理函数
type MessageContextHandler func(ctx *MessageContext)

type MessageContextHandlerGroup []MessageContextHandler

// MessageContext 消息处理上下文对象
type MessageContext struct {
	index           int
	abortIndex      int
	messageHandlers MessageContextHandlerGroup
	*dto.Message
}

// Next 主动调用下一个消息处理函数(或开始调用)
func (c *MessageContext) Next() {
	c.index++
	for c.index <= len(c.messageHandlers) {
		if c.IsAbort() {
			return
		}
		handle := c.messageHandlers[c.index-1]
		handle(c)
		c.index++
	}
}

// IsAbort 判断是否被中断
func (c *MessageContext) IsAbort() bool {
	return c.abortIndex > 0
}

// Abort 中断当前消息处理, 不会调用下一个消息处理函数, 但是不会中断当前的处理函数
func (c *MessageContext) Abort() {
	c.abortIndex = c.index
}

// AbortHandler 获取当前中断的消息处理函数
func (c *MessageContext) AbortHandler() MessageContextHandler {
	if c.abortIndex > 0 {
		return c.messageHandlers[c.abortIndex-1]
	}
	return nil
}

// MatchFunc 消息匹配函数,返回为true则表示匹配
type MatchFunc func(*dto.Message) bool

// MatchFuncList 将多个MatchFunc封装成一个MatchFunc
func MatchFuncList(matchFuncs ...MatchFunc) MatchFunc {
	return func(message *dto.Message) bool {
		for _, matchFunc := range matchFuncs {
			if !matchFunc(message) {
				return false
			}
		}
		return true
	}
}

type matchNode struct {
	matchFunc MatchFunc
	group     MessageContextHandlerGroup
}

type matchNodes []*matchNode

// MessageMatchDispatcher impl MessageDispatcher interface
//
//	dispatcher := NewMessageMatchDispatcher()
//	dispatcher.OnText(func(msg *dto.Message){
//			msg.ReplyText("hello")
//	})
//	bot := DefaultBot()
//	bot.MessageHandler = DispatchMessage(dispatcher)
type MessageMatchDispatcher struct {
	async        bool
	matchNodes   matchNodes
	workerOnce   sync.Once
	workerQueues []chan *MessageContext
}

// NewMessageMatchDispatcher Constructor
func NewMessageMatchDispatcher() *MessageMatchDispatcher {
	return &MessageMatchDispatcher{}
}

// SetAsync 设置是否异步处理
func (m *MessageMatchDispatcher) SetAsync(async bool) {
	m.async = async
	if async {
		m.startWorkers()
	}
}

// Dispatch impl MessageDispatcher
// 遍历 MessageMatchDispatcher 所有的消息处理函数
// 获取所有匹配上的函数
// 执行处理的消息处理方法
func (m *MessageMatchDispatcher) Dispatch(msg *dto.Message) {
	var group MessageContextHandlerGroup
	for _, node := range m.matchNodes {
		if node.matchFunc(msg) {
			group = append(group, node.group...)
		}
	}
	ctx := &MessageContext{Message: msg, messageHandlers: group}
	if m.async {
		m.startWorkers()
		m.workerQueues[m.workerIndex(msg)] <- ctx
	} else {
		m.do(ctx)
	}
}

func positiveEnv(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (m *MessageMatchDispatcher) startWorkers() {
	m.workerOnce.Do(func() {
		workerCount := positiveEnv("PLUGIN_WORKERS", 4)
		queueSize := positiveEnv("PLUGIN_QUEUE_SIZE", 500)
		perWorkerQueue := queueSize / workerCount
		if perWorkerQueue < 1 {
			perWorkerQueue = 1
		}
		m.workerQueues = make([]chan *MessageContext, workerCount)
		for index := range m.workerQueues {
			queue := make(chan *MessageContext, perWorkerQueue)
			m.workerQueues[index] = queue
			go func(jobs <-chan *MessageContext) {
				for ctx := range jobs {
					m.doSafely(ctx)
				}
			}(queue)
		}
		log.Printf("消息插件工作池已启动: workers=%d queue=%d", workerCount, queueSize)
	})
}

func (m *MessageMatchDispatcher) doSafely(ctx *MessageContext) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("消息插件工作器发生异常: %v", recovered)
		}
	}()
	m.do(ctx)
}

func (m *MessageMatchDispatcher) workerIndex(msg *dto.Message) int {
	key := msg.FromUser
	if key == "" {
		key = msg.GroupUser
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(key))
	return int(hasher.Sum32() % uint32(len(m.workerQueues)))
}

func (m *MessageMatchDispatcher) do(ctx *MessageContext) {
	ctx.Next()
}

// RegisterHandler 注册消息处理函数, 根据自己的需求自定义
// matchFunc返回true则表示处理对应的handlers
func (m *MessageMatchDispatcher) RegisterHandler(matchFunc MatchFunc, handlers ...MessageContextHandler) {
	if matchFunc == nil {
		panic("MatchFunc can not be nil")
	}
	node := &matchNode{matchFunc: matchFunc, group: handlers}
	m.matchNodes = append(m.matchNodes, node)
}
