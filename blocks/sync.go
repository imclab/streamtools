package blocks

import (
	"container/heap"
	"encoding/json"
	"log"
	"reflect"
	"strings"
	"time"
)

func Sync(b *Block) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	type syncRule struct {
		Path string
		Lag  int
	}

	var rule *syncRule
	lagTime := time.Duration(0)
	emitTick := time.NewTimer(500 * time.Millisecond)

	for {
		select {
		case m := <-b.Routes["set_rule"]:
			if rule  == nil {
				rule = &syncRule{}
			}
			unmarshal(m, rule)
			lagTime = time.Duration(time.Duration(rule.Lag) * time.Second)
		case m := <-b.Routes["get_rule"]:
			if rule == nil {
				marshal(m, &syncRule{})
			} else {
				marshal(m, rule)
			}
		case msg := <-b.AddChan:
			updateOutChans(msg, b)
		case <-b.QuitChan:
			quit(b)
			return
		case <-emitTick.C:
		case msg := <-b.InChan:
			if rule == nil {
				break
			}

			keys := strings.Split(rule.Path, ".")
			msgTime, err := Get(msg, keys...)
			if err != nil {
				log.Println(err.Error())
			}
			msgTimeF, ok := msgTime.(float64)
			msgTimeI := int64(msgTimeF)
			if !ok {
				v, _ := json.Marshal(msg)
				log.Println(string(v))
				log.Println(reflect.TypeOf(msgTime))
				log.Println(msgTime)
				log.Println(msgTimeI)
				log.Println("could not cast time key to int")
			}

			// assuming the value is in MS
			// TODO: make this more explicit and/or flexible
			ms := time.Unix(0, int64(time.Duration(msgTimeI)*time.Millisecond))

			queueMessage := &PQMessage{
				val: msg,
				t:   ms,
			}

			heap.Push(pq, queueMessage)
		}
		now := time.Now()
		for {
			item, diff := pq.PeekAndShift(now, lagTime)
			if item == nil {
				// then the queue is empty. Pause for 5 seconds before checking again
				if diff == 0 {
					diff = time.Duration(500) * time.Millisecond
				}

				emitTick.Reset(diff)
				break
			}
			broadcast(b.OutChans, &item.(*PQMessage).val)
		}
	}

}
