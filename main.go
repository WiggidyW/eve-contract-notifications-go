package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"github.com/WiggidyW/eve-contract-notifications-go/contracts"
	"github.com/WiggidyW/eve-contract-notifications-go/discord"
	"github.com/WiggidyW/eve-contract-notifications-go/hashcode"
	"github.com/WiggidyW/eve-contract-notifications-go/storage"
)

const (
	EXPIRING_SOON time.Duration = 48 * time.Hour
)

type NotifyOf struct {
	New      *[]contracts.Contract
	Expiring *[]contracts.Contract
}

func init() {
	functions.HTTP("Run", run)
}

func run(_ http.ResponseWriter, _ *http.Request) {
	main()
}

func main() {
	ctx := context.Background()

	wg := sync.WaitGroup{}
	errChan := make(chan error, 2)
	contractChan := make(chan *[]contracts.Contract)
	// Add a buffer to this channel so we don't hang forever on error.
	discordChan := make(chan discord.NotificationContracts, 1)

	wg.Add(1)
	go func(
		ctx context.Context,
		wg *sync.WaitGroup,
		errChan chan error,
		contractChan chan *[]contracts.Contract,
	) {
		defer wg.Done()
		contractGetter, err := contracts.NewContractGetter(ctx)
		if err != nil {
			errChan <- err
			return
		}
		contracts, err := contractGetter.Get(ctx)
		if err != nil {
			errChan <- err
		} else {
			contractChan <- contracts
		}
	}(ctx, &wg, errChan, contractChan)

	wg.Add(1)
	go func(
		wg *sync.WaitGroup,
		errChan chan error,
		discordChan chan discord.NotificationContracts,
	) {
		defer wg.Done()
		d, err := discord.OpenDiscordSession()
		if err != nil {
			errChan <- fmt.Errorf(
				"failed to open discord session: %v",
				err,
			)
			return
		}
		defer d.Close()
		contracts := <-discordChan
		err = discord.WriteContracts(d, contracts)
		if err != nil {
			errChan <- fmt.Errorf(
				"failed to write contracts to discord: %v",
				err,
			)
		}

	}(&wg, errChan, discordChan)

	// Create the storage client.
	storage, err := storage.NewStorage(ctx)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	// Get the previous run hash codes.
	prevRun, err := storage.Get(ctx)
	if err != nil {
		log.Fatalf("failed to get previous run hash codes: %v", err)
	}

	// Store all hash codes from this run.
	var hashCodes []hashcode.HashCode
	// Find and store contracts that weren't in the previous run.
	newContracts := make([]contracts.Contract, 0)
	// Find and store contracts that are expiring soon.
	expContracts := make([]contracts.Contract, 0)

	// Get the contracts from the channel.
	select {
	case err := <-errChan:
		log.Fatal(err)
	case contractsPtr := <-contractChan:
		contracts := *contractsPtr
		hashCodes = make([]hashcode.HashCode, 0, len(contracts))

		for _, contract := range contracts {
			hashCode := contract.HashCode
			hashCodes = append(hashCodes, hashCode)
			if !prevRun.Contains(hashCode) {
				newContracts = append(newContracts, contract)
			}
			if contract.Expires.Before(time.Now().Add(EXPIRING_SOON)) {
				expContracts = append(expContracts, contract)
			}
		}
	}

	// Write the contracts to discord.
	discordChan <- discord.NotificationContracts{
		New:      &newContracts,
		Expiring: &expContracts,
	}

	// Store the hash codes from this run. Don't panic yet if this fails.
	err = storage.Set(ctx, &hashCodes)

	// Wait for the goroutines to finish.
	wg.Wait()

	// Panic now if storing the hash codes failed. Discord has been notified.
	if err != nil {
		log.Fatalf("failed to store hash codes: %v", err)
	}

	// Check for errors.
	select {
	case err := <-errChan:
		log.Fatal(err)
	default:
	}
}
