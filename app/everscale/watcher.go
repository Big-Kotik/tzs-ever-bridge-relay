package everscale

import (
	"encoding/json"
	"fmt"
	goton "github.com/move-ton/ton-client-go"
	"github.com/move-ton/ton-client-go/domain"
	"log"
	"time"
)

type EventWatcher struct {
	Ton     *goton.Ton
	ABI     *domain.Abi
	Address string
	Cache   map[string]struct{}
}

func NewEventWatcher(address string, ton *goton.Ton, abi *domain.Abi) *EventWatcher {
	return &EventWatcher{ton, abi, address, make(map[string]struct{})}
}

func (cw *EventWatcher) RunWatcher(channel chan *domain.DecodedMessageBody, filter func(message *domain.DecodedMessageBody) bool) {
	timeResult, err := cw.Ton.Net.Query(&domain.ParamsOfQuery{
		Query: "query { info { time } }",
	})

	if err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	log.Println(string(timeResult.Result))

	var timeStruct struct {
		Data struct {
			Info struct {
				Time uint64 `json:"time"`
			} `json:"info"`
		} `json:"data"`
	}

	if err := json.Unmarshal(timeResult.Result, &timeStruct); err != nil {
		log.Printf("err: %s\n", err)
		return
	}

	lastCheckTime := timeStruct.Data.Info.Time / 1000

	for {
		collection, err := cw.Ton.Net.QueryCollection(&domain.ParamsOfQueryCollection{
			Collection: "messages",
			Filter:     json.RawMessage(fmt.Sprintf(`{"src":{"eq":"%s"}, "msg_type":{"eq":2}, "created_at":{"gt":%d}}`, cw.Address, lastCheckTime)),
			Result:     "id, body, created_at",
			Order:      []*domain.OrderBy{{"created_at", "ASC"}},
		})

		if err != nil {
			log.Printf("err: %s\n", err)
			return
		}

		for _, elem := range collection.Result {
			var msg struct {
				ID   string `json:"id"`
				Body string `json:"body"`
				Time uint64 `json:"created_at"`
			}

			err := json.Unmarshal(elem, &msg)
			if err != nil {
				log.Printf("err: %s\n", err)
			}

			lastCheckTime = msg.Time

			res, err := cw.Ton.Abi.DecodeMessageBody(&domain.ParamsOfDecodeMessageBody{
				Body: msg.Body,
				Abi:  cw.ABI,
			})
			if err != nil {
				log.Printf("err: %s\n", err)
				continue
			}

			if filter(res) {
				channel <- res
			}
		}

		time.Sleep(2 * time.Second)
	}
}
