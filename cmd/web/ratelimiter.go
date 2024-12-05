package main

import (
	"net/http"
	"time"
)

type Visitor struct {
	Count int
}

func (app *Application) rateLimiterMW(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if visitor, ok := app.Limits[ip]; ok {
			if visitor.Count > 3 {
				app.ErrorPage(w, http.StatusTooManyRequests, "DDoSer go away")
				return
			}
			app.Limits[ip].Count++
		} else {
			app.Limits[r.RemoteAddr] = &Visitor{}
		}

		//for keyIP, visitor := range app.Limits {
		//	fmt.Println("ip:", keyIP, "\t", "visitor counter:", visitor.Count)
		//}

		next.ServeHTTP(w, r)
	}
}

func disCounter(limits map[string]*Visitor) {
	for {
		for _, visitor := range limits {
			if visitor.Count > 0 {
				visitor.Count--
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func createLimiter() map[string]*Visitor {
	limits := make(map[string]*Visitor)
	go disCounter(limits)
	return limits
}
