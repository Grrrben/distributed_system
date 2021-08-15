package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const ServerPort = ":3000"
const ServicesUrl = "http://localhost" + ServerPort + "/services"

// package internal registry holding all Registrations
type registry struct {
	registrations []Registration
	mu            *sync.RWMutex
}

func (r *registry) add(reg Registration) error {
	r.mu.Lock()
	r.registrations = append(r.registrations, reg)
	r.mu.Unlock()

	err := r.sendRequiredServices(reg)
	if err != nil {
		return err
	}

	r.notify(patch{
		Added: []patchEntry{
			patchEntry{
				Name: reg.ServiceName,
				URL:  reg.ServiceUrl,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) remove(url string) (string, error) {
	var p patchEntry
	for i := range r.registrations {
		if r.registrations[i].ServiceUrl == url {
			p = patchEntry{
				Name: r.registrations[i].ServiceName,
				URL:  r.registrations[i].ServiceUrl,
			}

			name := string(r.registrations[i].ServiceName)
			r.mu.Lock()
			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)
			r.mu.Unlock()

			// loop the altered []Registration to notify them
			for range r.registrations {
				r.notify(patch{Removed: []patchEntry{p}})
			}

			return name, nil
		}
	}

	return "", fmt.Errorf("service with url %v not found", url)
}

func (r *registry) sendRequiredServices(reg Registration) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var p patch
	for _, serviceReg := range r.registrations {
		for _, requiredService := range reg.RequiredServices {
			if serviceReg.ServiceName == requiredService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceUrl,
				})
			}
		}
	}

	err := r.sendPatch(p, reg.ServiceUpdateUrl)
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) sendPatch(p patch, url string) error {
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) notify(fullPatch patch) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reg := range r.registrations {
		go func(reg Registration) {
			updated := false
			p := patch{}
			for _, requiredService := range reg.RequiredServices {
				for _, added := range fullPatch.Added {
					if added.Name == requiredService {
						p.Added = append(p.Added, added)
						updated = true
					}
				}
			}

			for _, requiredService := range reg.RequiredServices {
				for _, removed := range fullPatch.Removed {
					if removed.Name == requiredService {
						p.Removed = append(p.Removed, removed)
						updated = true
					}
				}
			}

			if updated {
				err := r.sendPatch(p, reg.ServiceUpdateUrl)
				if err != nil {
					log.Println(err)
				}
			}
		}(reg)
	}
}

func (r *registry) heartbeat(t time.Duration) {
	for {
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			go func(reg Registration) {
				defer wg.Done()

				failedAttempts := 0
				for attempts := 0; attempts < 3; attempts++ {
					res, err := http.Get(reg.HeartbeatUrl)

					if err != nil {
						// not sure what to do here, as we don't know the error
						log.Println(err)
						failedAttempts++
						if failedAttempts > 2 {
							fmt.Printf("health check failed for %s after 3 attempts\n", reg.ServiceName)
							_, err := r.remove(reg.ServiceUrl)
							if err != nil {
								log.Println(err)
							}
							break
						}
						continue
					}

					if res.StatusCode != http.StatusOK {
						// heartbeat failed, remove it from registrations
						failedAttempts++
						_, err := r.remove(reg.ServiceUrl)
						if err != nil {
							log.Println(err)
						}
						fmt.Printf("health check failed for %s\n", reg.ServiceName)

						// try again in a sec
						time.Sleep(1 * time.Second)
						continue
					}

					// reaching this means http.StatusOK
					fmt.Printf("health check OK for %s\n", reg.ServiceName)

					if failedAttempts > 0 {
						// add it as it was removed before...
						r.add(reg)
						failedAttempts = 0
					}

					break
				}
			}(reg)
		}
		wg.Wait()
		time.Sleep(t)
	}
}

var reg = registry{registrations: make([]Registration, 0), mu: new(sync.RWMutex)}

var once sync.Once

func SetUpHeartBeatCheck() {
	once.Do(func() {
		go reg.heartbeat(3 * time.Second)
	})
}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println("Request received")
	switch req.Method {
	case http.MethodPost:
		var r Registration
		dec := json.NewDecoder(req.Body)
		err := dec.Decode(&r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("Adding service: %v with URL: %s\n", r.ServiceName, r.ServiceUrl)
		err = reg.add(r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case http.MethodDelete:
		payload, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		url := string(payload) // as it was a []byte
		name, err := reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Removed %s from registry", name)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
