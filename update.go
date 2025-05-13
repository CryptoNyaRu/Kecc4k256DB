package kecc4k256db

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	_4ByteMethodsEndpoint = "https://www.4byte.directory/api/v1/signatures/?format=json&ordering=created_at&page="
	_4ByteEventsEndpoint  = "https://www.4byte.directory/api/v1/event-signatures/?format=json&ordering=created_at&page="
)

const (
	Updating = iota
	Successful
	Warning
	Failed
)

type UpdateProgress struct {
	Status int
	Log    string
}

type Logger struct {
	Info    func(contents ...any)
	Success func(contents ...any)
	Warning func(contents ...any)
	Error   func(contents ...any)
}

func (k *Kecc4k256DB) UpdateAsync() (methodsProgress chan *UpdateProgress, eventsProgress chan *UpdateProgress) {
	k.updateLock.Lock()

	methodsProgress = make(chan *UpdateProgress, 256)
	eventsProgress = make(chan *UpdateProgress, 256)

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		defer func() {
			close(methodsProgress)
			waitGroup.Done()
		}()

		maintenance, err := k.Maintenance()
		if err != nil {
			methodsProgress <- &UpdateProgress{
				Status: Failed,
				Log:    fmt.Sprintf("Failed to get methods maintenance: %s", err),
			}
			return
		}

		for page := maintenance.MethodsPage + 1; ; page++ {
			var signatureList *signatureList
			for {
				signatureList, err = fetchMethods(int(page))
				if err != nil {
					methodsProgress <- &UpdateProgress{
						Status: Warning,
						Log:    fmt.Sprintf("Failed to fetch methods signature list: %s", err),
					}
					time.Sleep(time.Millisecond * 100)
					continue
				}
				break
			}

			var methodRecords []*MethodRecord
			for _, record := range signatureList.Results {
				if int64(record.Id) > maintenance.MethodsID {
					methodRecords = append(methodRecords, &MethodRecord{
						Selector: record.HexSignature,
						Method:   record.TextSignature,
					})
				}
			}
			if len(methodRecords) > 0 {
				if err = k.UpsertMethodRecords(methodRecords); err != nil {
					methodsProgress <- &UpdateProgress{
						Status: Failed,
						Log:    fmt.Sprintf("Failed to upsert method records: %s", err),
					}
					return
				}

				savedID := int64(signatureList.Results[len(signatureList.Results)-1].Id)
				var savedPage int64
				if len(methodRecords) == 100 {
					savedPage = page
				} else {
					savedPage = page - 1
				}
				if err = k.UpdateMethodsMaintenance(savedPage, savedID); err != nil {
					methodsProgress <- &UpdateProgress{
						Status: Failed,
						Log:    fmt.Sprintf("Failed to update methods maintenance: %s", err),
					}
					return
				}

				methodsProgress <- &UpdateProgress{
					Status: Updating,
					Log:    fmt.Sprintf("Method records upserted, page: %d, records in page: %d", page, len(signatureList.Results)),
				}
			}

			if signatureList.Next == nil {
				break
			}
		}

		records, _ := k.MethodRecords()
		methodsProgress <- &UpdateProgress{
			Status: Successful,
			Log:    fmt.Sprintf("Methods are up to date with records: %d", records),
		}
	}()

	go func() {
		defer func() {
			close(eventsProgress)
			waitGroup.Done()
		}()

		maintenance, err := k.Maintenance()
		if err != nil {
			eventsProgress <- &UpdateProgress{
				Status: Failed,
				Log:    fmt.Sprintf("Failed to get events maintenance: %s", err),
			}
			return
		}

		for page := maintenance.EventsPage + 1; ; page++ {
			var signatureList *signatureList
			for {
				signatureList, err = fetchEvents(int(page))
				if err != nil {
					eventsProgress <- &UpdateProgress{
						Status: Warning,
						Log:    fmt.Sprintf("Failed to fetch events signature list: %s", err),
					}
					time.Sleep(time.Millisecond * 100)
					continue
				}
				break
			}

			var eventRecords []*EventRecord
			for _, record := range signatureList.Results {
				if int64(record.Id) > maintenance.EventsID {
					eventRecords = append(eventRecords, &EventRecord{
						Signature: record.HexSignature,
						Event:     record.TextSignature,
					})
				}
			}
			if len(eventRecords) > 0 {
				if err = k.UpsertEventRecords(eventRecords); err != nil {
					eventsProgress <- &UpdateProgress{
						Status: Failed,
						Log:    fmt.Sprintf("Failed to upsert event records: %s", err),
					}
					return
				}

				savedID := int64(signatureList.Results[len(signatureList.Results)-1].Id)
				var savedPage int64
				if len(eventRecords) == 100 {
					savedPage = page
				} else {
					savedPage = page - 1
				}
				if err = k.UpdateEventsMaintenance(savedPage, savedID); err != nil {
					eventsProgress <- &UpdateProgress{
						Status: Failed,
						Log:    fmt.Sprintf("Failed to update events maintenance: %s", err),
					}
					return
				}

				eventsProgress <- &UpdateProgress{
					Status: Updating,
					Log:    fmt.Sprintf("Event records upserted, page: %d, records in page: %d", page, len(signatureList.Results)),
				}
			}

			if signatureList.Next == nil {
				break
			}
		}

		eventRecords, _ := k.EventRecords()
		eventsProgress <- &UpdateProgress{
			Status: Successful,
			Log:    fmt.Sprintf("Events are up to date with records: %d", eventRecords),
		}
	}()

	go func() {
		waitGroup.Wait()
		k.updateLock.Unlock()
	}()

	return methodsProgress, eventsProgress
}
func (k *Kecc4k256DB) UpdateSync(logger *Logger) {
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	methodsProgress, eventsProgress := k.UpdateAsync()

	go func() {
		for updateProgress := range methodsProgress {
			if logger != nil {
				switch updateProgress.Status {
				case Updating:
					logger.Info(updateProgress.Log)

				case Successful:
					logger.Success(updateProgress.Log)

				case Warning:
					logger.Warning(updateProgress.Log)

				case Failed:
					logger.Error(updateProgress.Log)
				}
			}
		}

		waitGroup.Done()
	}()

	go func() {
		for updateProgress := range eventsProgress {
			if logger != nil {
				switch updateProgress.Status {
				case Updating:
					logger.Info(updateProgress.Log)

				case Successful:
					logger.Success(updateProgress.Log)

				case Warning:
					logger.Warning(updateProgress.Log)

				case Failed:
					logger.Error(updateProgress.Log)
				}
			}
		}

		waitGroup.Done()
	}()

	waitGroup.Wait()
}

type signatureList struct {
	Next    *string `json:"next,omitempty"`
	Results []struct {
		Id            int    `json:"id"`
		TextSignature string `json:"text_signature"`
		HexSignature  string `json:"hex_signature"`
	} `json:"results"`
}

func fetchMethods(page int) (signatureList *signatureList, err error) {
	return fetchSignatureList(_4ByteMethodsEndpoint + fmt.Sprint(page))
}
func fetchEvents(page int) (signatureList *signatureList, err error) {
	return fetchSignatureList(_4ByteEventsEndpoint + fmt.Sprint(page))
}
func fetchSignatureList(url string) (signatureList *signatureList, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set(`Accept`, `application/json, text/plain, */*`)

	httpClient := &http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	rawResult, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(rawResult, &signatureList); err != nil {
		return nil, err
	}

	return signatureList, nil
}
