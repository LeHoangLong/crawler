package repositories

import (
	"context"
	"fmt"
	"sync"

	"github.com/playwright-community/playwright-go"
)

type PageRepository struct {
	MaxNumberOfPages                   uint32
	TotalNumberOfPages                 uint32
	FreePages                          []*playwright.Page
	mutex                              *sync.Mutex
	waitingQueue                       map[uint64]chan *playwright.Page
	waitingQueueCounter                uint64
	nextWaitingQueueToReceiveNewPageId uint64
	browser                            *playwright.Browser
}

func MakePageRepository(
	iMaxNumberOfPages uint32,
	iBrowser *playwright.Browser,
) PageRepository {
	return PageRepository{
		MaxNumberOfPages:                   iMaxNumberOfPages,
		TotalNumberOfPages:                 0,
		FreePages:                          []*playwright.Page{},
		mutex:                              &sync.Mutex{},
		waitingQueue:                       map[uint64]chan *playwright.Page{},
		waitingQueueCounter:                0,
		nextWaitingQueueToReceiveNewPageId: 0,
		browser:                            iBrowser,
	}
}

func (r *PageRepository) GetPage(iContext context.Context) (*playwright.Page, error) {
	fmt.Println("getting page")
	r.mutex.Lock()
	if len(r.FreePages) > 0 {
		page := r.FreePages[len(r.FreePages)-1]
		r.FreePages = r.FreePages[0 : len(r.FreePages)-1]
		r.mutex.Unlock()
		return page, nil
	} else if r.TotalNumberOfPages < r.MaxNumberOfPages {
		page, err := (*r.browser).NewPage()
		r.TotalNumberOfPages += 1
		fmt.Println("r.TotalNumberOfPages: ", r.TotalNumberOfPages)
		r.mutex.Unlock()
		return &page, err
	} else {
		waitChannelId := r.waitingQueueCounter
		waitChannel := make(chan *playwright.Page)
		r.waitingQueue[waitChannelId] = waitChannel
		r.waitingQueueCounter++
		r.mutex.Unlock()

		var ret *playwright.Page = nil
		select {
		case <-iContext.Done():
			/// do nothing
		case newPage := <-waitChannel:
			/// set to new page
			ret = newPage
		}

		/// remove waiting queue
		r.mutex.Lock()
		delete(r.waitingQueue, waitChannelId)
		r.mutex.Unlock()

		return ret, nil
	}
}

func (r *PageRepository) ReturnPage(page *playwright.Page) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if len(r.waitingQueue) > 0 {
		for r.nextWaitingQueueToReceiveNewPageId < r.waitingQueueCounter {
			if waitChannel, ok := r.waitingQueue[r.nextWaitingQueueToReceiveNewPageId]; ok {
				waitChannel <- page
				return
			} else {
				/// might have been cancelled, so increase id by 1 to try again
				r.nextWaitingQueueToReceiveNewPageId++
			}
		}
	} else {
		r.FreePages = append(r.FreePages, page)
	}
}
