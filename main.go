package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	models "github.com/latonaio/salesforce-data-models"
	"github.com/latonaio/aion-core/pkg/go-client/msclient"
	"github.com/latonaio/aion-core/pkg/log"
	"github.com/latonaio/data-interface-for-salesforce-contract-bulk-get/internal/handlers"
	"github.com/latonaio/data-interface-for-salesforce-contract-bulk-get/internal/resources"
)

func main() {
	// Create Kanban client
	ctx, cancel := context.WithCancel(context.Background())
	if err := models.NewDBConPool(ctx); err != nil {
		panic(err)
	}
	defer cancel()
	if err := newKanbanClient(ctx); err != nil {
		log.Fatalf("failed to get kanban client: %v", err)
	}
	log.Printf("successful get kanban client")
	defer kanbanClient.Close()

	// Get Kanban channel by Kanban client
	kanbanCh := kanbanClient.GetKanbanCh()
	log.Printf("successful get kanban channel")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM)
	for {
		select {
		case s := <-signalCh:
			fmt.Printf("received signal: %s", s.String())
			goto END
		case k := <-kanbanCh:
			if k == nil {
				goto NODATA
			}

			// Get metadata from Kanban
			fromMetadata, err := msclient.GetMetadataByMap(k)
			if err != nil {
				log.Printf("failed to get metadata.err: %d", err)
				continue
			}

			log.Printf("got metadata from kanban")
			log.Printf("metadata: %v", fromMetadata)

			toMetadata, err := handle(fromMetadata)
			if err != nil {
				log.Printf("failed to handle\n%v", err)
			}

			// Write metadata for kanban
			if toMetadata != nil {
				if err := writeKanban(toMetadata); err != nil {
					log.Printf("failed to write kanban: %v", err)
					continue
				}
				log.Printf("write metadata to kanban")
				log.Printf("metadata: %v", toMetadata)
			}
		}
	NODATA:
	}
END:
}

func handle(fromMetadata map[string]interface{}) (map[string]interface{}, error) {
	ck, ok := fromMetadata["connection_type"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid connection key")
	}

	if ck == "response" {
		if err := handlers.HandleContract(fromMetadata); err != nil {
			log.Printf("failed to handle contract: %v", err)
		}
		return nil, nil
	}

	if ck == "request" {
		// Get Contract from metadata
		Contract, err := resources.NewContract(fromMetadata)
		if err != nil {
			log.Printf("failed to construct Contract: %v", err)
		}

		// Build metadata for kanban
		toMetadata, err := Contract.BuildMetadata()
		if err != nil {
			return nil, fmt.Errorf("failed to build metadata: %v", err)
		}
		return toMetadata, nil
	}
	return nil, errors.New("failed to handle")
}
