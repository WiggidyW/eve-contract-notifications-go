package storage

import (
	"github.com/WiggidyW/eve-contract-notifications-go/hashcode"

	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	CREDS      string = `##FIRESTORE_CREDS##`
	PROJECT_ID string = "##FIRESTORE_PROJECT_ID##"

	COLLECTION string = "contract_notifications"
	DOCUMENT   string = "buyback_last_run"
	FIELD      string = "hash_codes"
)

type FirestoreStorage struct {
	client *firestore.Client
}

func NewFirestoreStorage(ctx context.Context) (*FirestoreStorage, error) {
	client, err := firestore.NewClient(
		ctx,
		PROJECT_ID,
		option.WithCredentialsJSON([]byte(CREDS)),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create firestore client: %v",
			err,
		)
	}
	return &FirestoreStorage{client}, nil
}

func (f *FirestoreStorage) Get(
	ctx context.Context,
) (hashcode.HashCodeSet, error) {
	// Retrieve the document.
	doc, err := f.client.Collection(COLLECTION).Doc(DOCUMENT).Get(ctx)
	if err != nil {
		status, ok := status.FromError(err)
		// If the collection doesn't exist, return an empty set.
		if ok && status.Code() == codes.NotFound {
			return hashcode.HashCodeSetNew(), nil
		}
		// Otherwise, return the error.
		return nil, fmt.Errorf(
			"failed to get document: %v",
			err,
		)
	}
	// // If the document doesn't exist, return an empty set.
	// if !doc.Exists() {
	// 	return HashCodeSetNew(), nil
	// }
	// Retrieve the hash codes from the document.
	hashCodesRaw, err := doc.DataAt(FIELD)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get hash codes from document: %v",
			err,
		)
	}
	// Convert the hash codes to an array of strings.
	hashCodesArr, ok := hashCodesRaw.([]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"failed to convert hash codes to array: %v",
			err,
		)
	}
	// Convert the array of strings to a set of hash codes.
	hashCodes := hashcode.HashCodeSetWithCapacity(len(hashCodesArr))
	for _, hashCodeRaw := range hashCodesArr {
		// Convert the hash code to a string.
		hashCode, ok := hashCodeRaw.(hashcode.HashCode)
		if !ok {
			return nil, fmt.Errorf(
				"failed to convert hash code to string: %v",
				err,
			)
		}
		// Add the hash code to the set.
		hashCodes.Add(hashCode)
	}

	return hashCodes, nil
}

func (f *FirestoreStorage) Set(
	ctx context.Context,
	hashCodes *[]hashcode.HashCode,
) error {
	_, err := f.client.Collection(COLLECTION).Doc(DOCUMENT).Update(
		ctx,
		[]firestore.Update{{
			Path:  FIELD,
			Value: *hashCodes,
		}},
	)
	if err != nil {
		return fmt.Errorf(
			"failed to update document: %v",
			err,
		)
	}
	return nil
}
