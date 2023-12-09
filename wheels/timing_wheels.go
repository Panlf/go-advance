package wheels

import (
	"container/list"
	"sync"
	"time"
)

/*
*
单机版时间轮核心流程

	1、建立一个环状数据结构
	2、每个刻度对应一个时间范围
	3、创建定时任务时，根据距今的相对时长，推算出需要向后推移的刻度值
	4、倘若来到环形数组的结尾，则重新从起点开始计算，但是记录时把执行轮次数加1
	5、一个刻度可能存在多笔定时任务，所以每个刻度需要挂载一个定时任务链表
*/
type TimeWheel struct {
	//单例工具 保证时间轮停止操作只能执行一次
	sync.Once

	//	时间轮运行时间间隔
	interval time.Duration
	//	时间轮定时器
	ticker *time.Ticker
	//	停止时间轮的channel
	stopc chan struct{}
	//	新增定时任务的入口 channel
	addTaskCh chan *taskElement
	//	删除定时任务的入口
	removeTaskCh chan string
	// 通过list组成的环状数组。通过遍历环状数组的方式实现时间轮
	//	定时任务数量较大，每个slot槽内可能存在多个定时任务，因此通过list进行组装
	slots []*list.List
	//	当前遍历到的环状数组的索引
	curSlot int
	//	定时任务 key 到任务节点的映射，便于在 list 中删除任务节点
	keyToETask map[string]*list.Element
}

// 封装了一笔定时任务的明细信息
type taskElement struct {
	//	内聚了定时任务执行逻辑的闭包函数
	task func()
	// 定时任务挂载在环状数组中的索引位置
	pos int
	// 定时任务的延迟轮次，指的是 curSlot 指针还要扫描环状数组多少轮，才能满足执行该任务条件
	cycle int
	//	定时任务的唯一标识键
	key string
}

// 创建单机版时间轮 slotNum 时间轮环状数组长度	interval 扫描时间间隔
func NewTimeWheel(slotNum int, interval time.Duration) *TimeWheel {
	// 环状数组长度默认为 10
	if slotNum <= 0 {
		slotNum = 10
	}

	// 扫描时间间隔默认为1秒
	if interval <= 0 {
		interval = time.Second
	}

	//	初始化时间轮实例
	t := TimeWheel{
		interval:     interval,
		ticker:       time.NewTicker(interval),
		stopc:        make(chan struct{}),
		keyToETask:   make(map[string]*list.Element),
		slots:        make([]*list.List, 0, slotNum),
		addTaskCh:    make(chan *taskElement),
		removeTaskCh: make(chan string),
	}

	for i := 0; i < slotNum; i++ {
		t.slots = append(t.slots, list.New())
	}

	//	异步启动时间轮常驻 goroutine
	go t.run()

	return &t
}

func (t *TimeWheel) run() {
	defer func() {
		if err := recover(); err != nil {
			panic(err)
		}
	}()

	// 通过 for + select 的代码结构运行一个常驻 goroutine 是常规操作

	for {
		select {
		// 停止时间轮
		case <-t.stopc:
			return
		case <-t.ticker.C:
			t.tick()
		case task := <-t.addTaskCh:
			t.addTask(task)
		case removeKey := <-t.removeTaskCh:
			t.removeTask(removeKey)
		}
	}
}

// 停止时间轮
func (t *TimeWheel) Stop() {
	// 通过单例工具，保证channel 只能被关闭一次，避免 panic
	t.Do(
		func() {
			// 定制定时器 ticker
			t.ticker.Stop()
			//关闭定时器运行的stopc
			close(t.stopc)
		})
}

// 添加定时任务到时间轮中
func (t *TimeWheel) AddTask(key string, task func(), executeAt time.Time) {
	// 根据执行时间推算得到定时任务从属的slot位置，以及需要延迟的轮次
	pos, cycle := t.getPosAndCircle(executeAt)
	//定时任务通过 channel 进行投递
	t.addTaskCh <- &taskElement{
		pos:   pos,
		cycle: cycle,
		task:  task,
		key:   key,
	}
}

// 根据执行时间推算得到定时任务从属的slot位置，以及需要延迟的轮次
func (t *TimeWheel) getPosAndCircle(executeAt time.Time) (int, int) {
	delay := int(time.Until(executeAt))
	// 定时任务的延迟轮次
	cycle := delay / (len(t.slots) * int(t.interval))

	// 定时任务从属的环状数组 index
	pos := (t.curSlot + delay/int(t.interval)) % len(t.slots)
	return pos, cycle
}

// 常驻 goroutine 接收到创建定时任务后的处理逻辑
func (t *TimeWheel) addTask(task *taskElement) {
	// 获取到定时任务从属的环状数组 index 以及对应的list
	list := t.slots[task.pos]
	// 倘若定时任务 key 之前已存在，则需要先删除定时任务
	if _, ok := t.keyToETask[task.key]; ok {
		t.removeTask(task.key)
	}
	// 将定时任务追加到 list 尾部
	eTask := list.PushBack(task)
	// 建立定时任务 key 到将定时任务所处的节点
	t.keyToETask[task.key] = eTask
}

// 删除定时任务，投递信号
func (t *TimeWheel) RemoveTask(key string) {
	t.removeTaskCh <- key
}

// 时间轮常驻 goroutine 接收到删除任务信号后，执行的删除任务逻辑
func (t *TimeWheel) removeTask(key string) {
	eTask, ok := t.keyToETask[key]
	if !ok {
		return
	}
	// 将定时任务节点从映射 map 中移除
	delete(t.keyToETask, key)

	// 获取到定时任务节点后，将其从 list 中移除
	task, _ := eTask.Value.(*taskElement)
	_ = t.slots[task.pos].Remove(eTask)
}

// 常驻 goroutine 每次接收到定时信号后用于执行定时任务的逻辑
func (t *TimeWheel) tick() {
	// 根据 curSolt 获取到当前所处的环状数组索引位置，取出对应的list
	list := t.slots[t.curSlot]

	// 在方法返回前，推进 curSlot 指针的位置，进行环状遍历
	defer t.circularIncr()

	// 批量处理满足执行条件的定时任务
	t.execute(list)
}

// 执行定时任务，每次处理一个list
func (t *TimeWheel) execute(l *list.List) {
	//遍历list
	for e := l.Front(); e != nil; {
		// 获取到每个节点对应的定时任务信息
		taskElement, _ := e.Value.(*taskElement)
		// 倘若任务还存在延迟轮次，则只对 cycle 计数器进行扣减，本轮不作任务的执行
		if taskElement.cycle > 0 {
			taskElement.cycle--
			e = e.Next()
			continue
		}

		// 当前节点对应定时任务已达成执行条件，开启一个 goroutine 负责执行任务
		go func() {
			defer func() {
				if err := recover(); err != nil {
					panic(err)
				}
			}()
			taskElement.task()
		}()

		// 任务已执行，需要把对应的任务节点从 list 中删除
		next := e.Next()
		l.Remove(e)

		// 把任务 key 从映射 map 中删除
		delete(t.keyToETask, taskElement.key)

		e = next
	}

}

// 每次 tick 后需要推进 curSlot 指针的位置，slots在逻辑意义上是环状数组，所以在到达尾部时需要重新回到头部
func (t *TimeWheel) circularIncr() {
	t.curSlot = (t.curSlot + 1) % len(t.slots)
}
