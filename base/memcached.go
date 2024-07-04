package base

import (
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
)

func (b *Base) loadMemcached() {
	b.MemQ = NewMemcachedQueue([]string{b.Env.MEMCACHED_URL}, "video_queue")
}

type MemcachedQueue struct {
	client    *memcache.Client
	queueName string
}

func NewMemcachedQueue(servers []string, queueName string) *MemcachedQueue {
	client := memcache.New(servers...)
	return &MemcachedQueue{client: client, queueName: queueName}
}

func (mq *MemcachedQueue) Enqueue(item string) error {
	queue, err := mq.getQueue()
	if err != nil && err != memcache.ErrCacheMiss {
		return err
	}
	queue = append(queue, item)
	return mq.setQueue(queue)
}

func (mq *MemcachedQueue) Dequeue() (string, error) {
	queue, err := mq.getQueue()
	if err != nil {
		return "", err
	}
	if len(queue) == 0 {
		return "", nil // Queue is empty
	}
	item := queue[0]
	queue = queue[1:]
	err = mq.setQueue(queue)
	if err != nil {
		return "", err
	}
	return item, nil
}

func (mq *MemcachedQueue) Size() (int, error) {
	queue, err := mq.getQueue()
	if err != nil {
		return 0, err
	}
	return len(queue), nil
}

func (mq *MemcachedQueue) getQueue() ([]string, error) {
	item, err := mq.client.Get(mq.queueName)
	if err != nil {
		return nil, err
	}
	var queue []string
	err = json.Unmarshal(item.Value, &queue)
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func (mq *MemcachedQueue) setQueue(queue []string) error {
	data, err := json.Marshal(queue)
	if err != nil {
		return err
	}
	return mq.client.Set(&memcache.Item{Key: mq.queueName, Value: data})
}
