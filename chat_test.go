package main

import (
	"runtime"
	"testing"
	"time"
)

func TestChatClient_SendEvent(t *testing.T) {
	client := NewChatClient(WS_URL)

	done := make(chan bool, 0)

	defer func() {
		client.Close()
	}()

	client.OnConnected(func() {

		ticker := time.NewTicker(time.Second * 2);
		Done := make(chan bool, 0)


		for {
			select {
				case <-ticker.C:
					client.SendEvent(CHAT_EVENT_MESSAGE, &MessageArgs{
						Command: EVENT_PLAYING,
						Params: []interface{}{
							&Song{
								Name:     "Test",
								Duration: 100.0,
								Index:    0.0,
							},
							0,
							0,
							0,
						},
					})
				case <-Done:
					goto done
			}
		}

		done:

		time.AfterFunc(time.Second*20, func() {
			log.Info("20s")

			Done <- true
		})
	})

	client.Listen()

	time.AfterFunc(time.Second*30, func() {
		log.Info("30s timeout")
		done <- true
		client.Close()
	})

	<-done
}

func TestNewChatClient(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	client := NewChatClient(WS_URL)

	//go NewSendClient()

	listener, err := client.Listen()
	done := make(chan bool, 0)

	if err != nil {
		t.Error(err)
		return
	}

	time.AfterFunc(time.Second*30, func() {
		t.Log("30s timeout")
		done <- true
		client.Close()
	})

	for true {
		select {
		case <-done:
			goto end
		case msg := <-listener:
			t.Log(msg)
			break
		}
	}

end:
}
