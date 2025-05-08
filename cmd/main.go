package main

import (
	"fmt"
	"log"
	"select0rs"
	"time"
)

func main() {
	kecc4k256DB, err := kecc4k256db.Open("./kecc4k256.db")
	if err != nil {
		log.Fatalf("Failed to open: %s\n", err)
	}
	journalMode, err := kecc4k256DB.JournalMode()
	if err != nil {
		log.Fatalf("Failed to get journal mode: %s\n", err)
	}
	log.Printf("Database connection established, journal mode: %s\n\n", journalMode)

	maintenance, err := kecc4k256DB.Maintenance()
	if err != nil {
		log.Fatalf("Failed to get maintenance: %s\n", err)
	}
	methodRecords, err := kecc4k256DB.MethodRecords()
	if err != nil {
		log.Fatalf("Failed to get method records: %s\n", err)
	}
	eventRecords, err := kecc4k256DB.EventRecords()
	if err != nil {
		log.Fatalf("Failed to get event records: %s\n", err)
	}
	fmt.Println(fmt.Sprintf("[Methods]\nPage       : %d\nID         : %d\nRecords    : %d\nMaintenance: %s\n", maintenance.MethodsPage, maintenance.MethodsID, methodRecords, time.Unix(maintenance.MethodsMaintenanceTime, 0).String()))
	fmt.Println(fmt.Sprintf("[Events]\nPage       : %d\nID         : %d\nRecords    : %d\nMaintenance: %s\n", maintenance.EventsPage, maintenance.EventsPage, eventRecords, time.Unix(maintenance.EventsMaintenanceTime, 0).String()))

	selector := "0x0178fe3f"
	method := "getData(uint256)"
	log.Printf("selector: %s, method: %s\n", selector, method)
	log.Printf("%s %s\n", selector, method)
	log.Println(kecc4k256DB.GetMethodsBySelector(selector))
	log.Println(kecc4k256DB.GetSelectorByMethod(method))
	fmt.Println()

	signature := "0x00f80de212f43b06ae1124cfbc40d7cf760d91ce3d0133c263cbb00d81602a3e"
	event := "Router(address)"
	log.Printf("signature: %s, event: %s\n", signature, event)
	log.Println(kecc4k256DB.GetEventBySignature(signature))
	log.Println(kecc4k256DB.GetSignatureByEvent(event))
	fmt.Println()

	// UpdateSync
	{
		log.Println("UpdateSync will start in 3 seconds...")
		time.Sleep(3 * time.Second)

		kecc4k256DB.UpdateSync(&kecc4k256db.Logger{
			Info:    log.Println,
			Success: log.Println,
			Warning: log.Println,
			Error:   log.Println,
		})

		log.Println("UpdateSync done")
	}

	//// UpdateAsync()
	//{
	//	log.Println("UpdateAsync will start in 3 seconds...")
	//	time.Sleep(3 * time.Second)
	//
	//	waitGroup := sync.WaitGroup{}
	//	waitGroup.Add(2)
	//
	//	methodsProgress, eventsProgress := kecc4k256DB.UpdateAsync()
	//
	//	go func() {
	//		for updateProgress := range methodsProgress {
	//			switch updateProgress.Status {
	//			case kecc4k256db.Updating:
	//				log.Println(updateProgress.Log)
	//
	//			case kecc4k256db.Successful:
	//				log.Println(updateProgress.Log)
	//
	//			case kecc4k256db.Warning:
	//				log.Println(updateProgress.Log)
	//
	//			case kecc4k256db.Failed:
	//				log.Println(updateProgress.Log)
	//			}
	//		}
	//
	//		waitGroup.Done()
	//	}()
	//
	//	go func() {
	//		for updateProgress := range eventsProgress {
	//			switch updateProgress.Status {
	//			case kecc4k256db.Updating:
	//				log.Println(updateProgress.Log)
	//
	//			case kecc4k256db.Successful:
	//				log.Println(updateProgress.Log)
	//
	//			case kecc4k256db.Warning:
	//				log.Println(updateProgress.Log)
	//
	//			case kecc4k256db.Failed:
	//				log.Println(updateProgress.Log)
	//			}
	//		}
	//
	//		waitGroup.Done()
	//	}()
	//
	//	waitGroup.Wait()
	//
	//	log.Println("UpdateAsync done")
	//}
}
