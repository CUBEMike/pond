package pond

import (
	"io/ioutil"
	"log"
	"net/http"
	"runtime"

	"github.com/garyburd/redigo/redis"
)

type Pond struct {
	queue chan *Rock
}

func NewPond() *Pond {
	p := new(Pond)
	p.queue = make(chan *Rock)

	p.startWorkers()
	p.startBroadcasters()

	// Wake up conn!
	conn := pool.Get()
	conn.Do("PING")
	defer conn.Close()

	return p
}

func (p *Pond) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body := req.Body
        switch req.Method {
        case "POST":
                if body != nil {
                        bytes, _ := ioutil.ReadAll(body)
                        rock := NewRock(bytes)

                        if !rock.alreadySent() {
                                p.queue <- rock
                        }
                }

        case "GET":


        }
}

func (p *Pond) storeMessage(msg []byte) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("LPUSH", queue_key, msg)
	if err != nil {
		panic(err)
	}

	log.Println("---> Message stored", msg)
}

func (p *Pond) worker(i int) {
	log.Println("---> Starting the worker", i)

	var rock *Rock

	for {
		rock = <-p.queue
		if !rock.alreadySent() {
			log.Printf("----> [w:%d] Incoming message!: %s", i, string(rock.Message))
			p.storeMessage(rock.Message)
		}
	}
}

func (p *Pond) broadcaster(i int) {
	log.Println("---> Starting the broadcaster", i)
	conn := pool.Get()
	defer conn.Close()

	for {
		msg, _ := redis.Bytes(conn.Do("BRPOPLPUSH", queue_key, backup_key, 5))
		if msg != nil {
			rock := NewRock(msg)

			if !rock.alreadySent() {
				go p.sendToTheRiver(rock)
				rock.StoreForReading()
				rock.FlagAsSent()

				log.Printf("----> [b:%d] Sent message %s", i, rock.Message)
			}
		}

	}
}

func (p *Pond) sendToTheRiver(rock *Rock) {
	conn := pool.Get()
	defer conn.Close()

        redis.Values(conn.Do("HGETALL", friends_key))

        //http.Post("http://example.com/upload", "image/jpeg", &buf)
}

func (p *Pond) startWorkers() {
	cpu_count := runtime.NumCPU() / 2
	runtime.GOMAXPROCS(cpu_count)

	log.Printf("--> Starting %d workers\n", cpu_count)

	for i := 0; i < cpu_count; i++ {
		go p.worker(i)
	}
}

func (p *Pond) startBroadcasters() {
	cpu_count := runtime.NumCPU() / 2
	log.Printf("--> Starting %d broadcasters\n", cpu_count)

	for i := 0; i < cpu_count; i++ {
		go p.broadcaster(i)
	}
}
