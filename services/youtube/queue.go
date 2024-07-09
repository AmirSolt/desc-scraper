package youtube

import (
	"sync"
)

type Queue struct {
	queue []string
	mu    sync.Mutex
}

func (queue *Queue) Size() int {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	return len(queue.queue)
}

func (queue *Queue) Enqueue(item string) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.queue = append(queue.queue, item)
}

func (queue *Queue) Dequeue() (string, bool) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	if len(queue.queue) == 0 {
		return "", false
	}
	item := queue.queue[0]
	queue.queue = queue.queue[1:]
	return item, true
}

func (queue *Queue) EnqueueAll(items []string) {
	queue.mu.Lock()
	defer queue.mu.Unlock()
	queue.queue = append(queue.queue, items...)
}

// func main() {
// 	queue := &Queue{}

// 	var wg sync.WaitGroup
// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			queue.Enqueue(fmt.Sprintf("task-%d", id))
// 		}(i)
// 	}

// 	wg.Wait()

// 	fmt.Println("Queue size after enqueueing:", queue.Size())

// 	for i := 0; i < 10; i++ {
// 		wg.Add(1)
// 		go func(id int) {
// 			defer wg.Done()
// 			item, ok := queue.Dequeue()
// 			if ok {
// 				fmt.Printf("Task %d got item: %s\n", id, item)
// 			} else {
// 				fmt.Printf("Task %d found the queue empty\n", id)
// 			}
// 		}(i)
// 	}

// 	wg.Wait()

// 	fmt.Println("Queue size after dequeuing:", queue.Size())
// }
