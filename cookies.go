package lpse

import (
	"net/http"
	"sync"
)

type cookies struct {
	cs    []*http.Cookie
	mutex *sync.Mutex
}

func newCookies() *cookies {
	return &cookies{
		cs:    make([]*http.Cookie, 0),
		mutex: new(sync.Mutex),
	}
}

func (cks *cookies) get() []*http.Cookie {
	var cookies []*http.Cookie
	cks.mutex.Lock()
	cookies = cks.cs
	cks.mutex.Unlock()

	return cookies
}

func (cks *cookies) set(newCs []*http.Cookie) {
	cks.mutex.Lock()
	for _, newC := range newCs {
		found := false
		for i, oldC := range cks.cs {
			if newC.Name == oldC.Name && newC.Value != "" {
				cks.cs[i] = newC
				found = true
				break
			}
		}

		if !found {
			cks.cs = append(cks.cs, newC)
		}
	}

	cks.mutex.Unlock()
}
