package service

import (
	"context"
	"github.com/loks666/FlyAI/client"
	"log"
	"sync"
	"time"
)

type NonOwnerReplyAction int64

const (
	NonOwnerReplyIgnore NonOwnerReplyAction = iota
	NonOwnerReplyRandom
	NonOwnerReplyBalanceAndMute
)

const (
	nonOwnerMuteDuration    = time.Hour
	nonOwnerPendingDuration = 2 * time.Minute
)

const beginNonOwnerReplyScript = `
local state = redis.call('GET', KEYS[1])
if not state then
  redis.call('SET', KEYS[1], 'first_pending', 'PX', ARGV[1])
  return 1
end
if state == 'seen' then
  redis.call('SET', KEYS[1], 'balance_pending', 'PX', ARGV[1])
  return 2
end
return 0
`

const completeNonOwnerReplyScript = `
local state = redis.call('GET', KEYS[1])
local action = tonumber(ARGV[1])
local sent = ARGV[2] == '1'
if action == 1 then
  if sent then
    if state == 'first_pending' then redis.call('SET', KEYS[1], 'seen', 'PX', ARGV[3]) end
  else
    if state == 'first_pending' then redis.call('DEL', KEYS[1]) end
  end
elseif action == 2 then
  if sent then
    redis.call('SET', KEYS[1], 'muted', 'PX', ARGV[3])
  else
    if state == 'balance_pending' then redis.call('SET', KEYS[1], 'seen') end
  end
end
return 1
`

type memoryReplyState struct {
	phase     string
	expiresAt time.Time
}

type nonOwnerMemoryState struct {
	mu     sync.Mutex
	states map[string]memoryReplyState
	now    func() time.Time
}

func newNonOwnerMemoryState() *nonOwnerMemoryState {
	return &nonOwnerMemoryState{states: make(map[string]memoryReplyState), now: time.Now}
}

func (s *nonOwnerMemoryState) begin(userId string) NonOwnerReplyAction {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	state, exists := s.states[userId]
	if exists && !state.expiresAt.IsZero() && !now.Before(state.expiresAt) {
		delete(s.states, userId)
		exists = false
	}
	if !exists {
		s.states[userId] = memoryReplyState{phase: "first_pending", expiresAt: now.Add(nonOwnerPendingDuration)}
		return NonOwnerReplyRandom
	}
	if state.phase == "seen" {
		s.states[userId] = memoryReplyState{phase: "balance_pending", expiresAt: now.Add(nonOwnerPendingDuration)}
		return NonOwnerReplyBalanceAndMute
	}
	return NonOwnerReplyIgnore
}

func (s *nonOwnerMemoryState) observeBegin(userId string, action NonOwnerReplyAction) {
	if action == NonOwnerReplyIgnore {
		return
	}
	now := s.now()
	phase := "first_pending"
	if action == NonOwnerReplyBalanceAndMute {
		phase = "balance_pending"
	}
	s.mu.Lock()
	s.states[userId] = memoryReplyState{phase: phase, expiresAt: now.Add(nonOwnerPendingDuration)}
	s.mu.Unlock()
}

func (s *nonOwnerMemoryState) complete(userId string, action NonOwnerReplyAction, sent bool) {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	state := s.states[userId]
	switch action {
	case NonOwnerReplyRandom:
		if state.phase != "first_pending" {
			return
		}
		if sent {
			s.states[userId] = memoryReplyState{phase: "seen", expiresAt: now.Add(nonOwnerMuteDuration)}
		} else {
			delete(s.states, userId)
		}
	case NonOwnerReplyBalanceAndMute:
		if sent {
			s.states[userId] = memoryReplyState{phase: "muted", expiresAt: now.Add(nonOwnerMuteDuration)}
		} else if state.phase == "balance_pending" {
			s.states[userId] = memoryReplyState{phase: "seen"}
		}
	}
}

func (s *nonOwnerMemoryState) isSilenced(userId string) bool {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	state, exists := s.states[userId]
	if !exists {
		return false
	}
	if !state.expiresAt.IsZero() && !now.Before(state.expiresAt) {
		delete(s.states, userId)
		return false
	}
	return state.phase == "first_pending" || state.phase == "balance_pending" || state.phase == "muted"
}

var (
	nonOwnerReplyMemoryState = newNonOwnerMemoryState()
	redisLimitLogMu          sync.Mutex
	lastRedisLimitLog        time.Time
)

func nonOwnerReplyKey(userId string) string {
	return "flyai:non-owner-reply:v2:" + userId
}

// BeginNonOwnerReply returns the next global reply action for one sender wxid.
// The group id is intentionally excluded so changing groups cannot reset state.
func BeginNonOwnerReply(userId string) NonOwnerReplyAction {
	if userId == "" {
		return NonOwnerReplyIgnore
	}
	if redisClient := client.GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		result, err := redisClient.Eval(ctx, beginNonOwnerReplyScript, []string{nonOwnerReplyKey(userId)}, nonOwnerPendingDuration.Milliseconds()).Int64()
		cancel()
		if err == nil {
			action := NonOwnerReplyAction(result)
			nonOwnerReplyMemoryState.observeBegin(userId, action)
			return action
		}
		logRedisLimitFallback(err)
	}
	return nonOwnerReplyMemoryState.begin(userId)
}

// CompleteNonOwnerReply commits the state only after the bridge confirms that
// the reply was sent. The second successful reply starts the one-hour mute TTL.
func CompleteNonOwnerReply(userId string, action NonOwnerReplyAction, sent bool) {
	if userId == "" || action == NonOwnerReplyIgnore {
		return
	}
	if redisClient := client.GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		sentValue := "0"
		if sent {
			sentValue = "1"
		}
		err := redisClient.Eval(ctx, completeNonOwnerReplyScript, []string{nonOwnerReplyKey(userId)}, int64(action), sentValue, nonOwnerMuteDuration.Milliseconds()).Err()
		cancel()
		if err != nil {
			logRedisLimitFallback(err)
		}
	}
	nonOwnerReplyMemoryState.complete(userId, action, sent)
}

// ShouldIgnoreNonOwnerMessage is used by private-message handlers so a sender
// muted after the second group mention cannot bypass the rule via private chat.
func ShouldIgnoreNonOwnerMessage(userId string) bool {
	if userId == "" {
		return false
	}
	if redisClient := client.GetRedisClient(); redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		state, err := redisClient.Get(ctx, nonOwnerReplyKey(userId)).Result()
		cancel()
		if err == nil {
			return state == "first_pending" || state == "balance_pending" || state == "muted"
		}
	}
	return nonOwnerReplyMemoryState.isSilenced(userId)
}

func logRedisLimitFallback(err error) {
	redisLimitLogMu.Lock()
	defer redisLimitLogMu.Unlock()
	now := time.Now()
	if now.Sub(lastRedisLimitLog) >= time.Minute {
		log.Printf("Redis 非管理员回复状态不可用，已降级为进程内存: %v", err)
		lastRedisLimitLog = now
	}
}
