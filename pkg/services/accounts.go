// Copyright 2024 Itential Inc. All Rights Reserved
// Unauthorized copying of this file, via any medium is strictly prohibited
// Proprietary and confidential

package services

import (
	"errors"
	"fmt"

	"github.com/itential/ipctl/internal/logging"
	"github.com/itential/ipctl/pkg/client"
)

// Account represents a configured account configured on Itential Platform that
// can access the system.
type Account struct {
	// The account ID is a unique identifier that is assigned by the server and
	// can be used to identify a specific account in Platform.
	Id string `json:"_id"`

	// The email address for the account
	Email string `json:"email"`

	// The first name of the user for this account
	FirstName string `json:"firstname"`

	// Returns whether or not the account is inactive
	Inactive bool `json:"inactive"`

	// Returns whether or not the account is currently logged at the time of
	// the API call
	LoggedIn bool `json:"loggedIn"`

	// Identifies the origin of the account.  Since all accounts are federated
	// from other systems, this field provides an indication as to which
	// external system the account is sourced from.
	Provenance string `json:"provenance"`

	// The username associated with this account.  The username should be
	// unique within the system

	// The username associated with this account.  The username should be
	// unique within the system
	Username string `json:"username"`
}

// AccountService provides API access for mananging Itential Platform accounts.
type AccountService struct {
	BaseService
}

// NewAccountService creates and returns a new instance of AccountService using the
// client connection specified by c. The service provides methods for managing
// Itential Platform user accounts.
func NewAccountService(c client.Client) *AccountService {
	return &AccountService{
		BaseService: NewBaseService(c),
	}
}

// GetAll retrieves all user accounts from the Itential Platform by calling
// GET /authorization/accounts. This method handles pagination automatically,
// fetching all accounts across multiple pages if necessary. Returns a slice
// of Account structs or an error if the request fails.
func (svc *AccountService) GetAll() ([]Account, error) {
	logging.Trace()

	type Response struct {
		Results []Account `json:"results"`
		Total   int       `json:"total"`
	}

	var accounts []Account
	var limit = 100
	var skip = 0

	for {
		var res *Response

		if err := svc.GetRequest(&Request{
			uri:    "/authorization/accounts",
			params: &QueryParams{Limit: limit, Skip: skip},
		}, &res); err != nil {
			return nil, err
		}

		for _, ele := range res.Results {
			accounts = append(accounts, ele)
		}

		if len(accounts) == res.Total {
			break
		}

		skip += limit
	}

	logging.Info("Found %v account(s)", len(accounts))

	return accounts, nil
}

// Get retrieves a specific user account by ID from the Itential Platform by
// calling GET /authorization/accounts/{id}. Returns a pointer to the Account
// struct or an error if the account is not found or the request fails.
func (svc *AccountService) Get(id string) (*Account, error) {
	logging.Trace()

	var res *Account

	var uri = fmt.Sprintf("/authorization/accounts/%s", id)

	if err := svc.BaseService.Get(uri, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// Deactivate sets an account to inactive status by calling PATCH /authorization/accounts/{id}
// with inactive=true. The account will no longer be able to access the system.
// Returns an error if the request fails or the account is not found.
func (svc *AccountService) Deactivate(id string) error {
	logging.Trace()
	return svc.PatchRequest(&Request{
		uri:  fmt.Sprintf("/authorization/accounts/%s", id),
		body: map[string]interface{}{"inactive": true},
	}, nil)
}

// Activate sets an account to active status by calling PATCH /authorization/accounts/{id}
// with inactive=false. The account will be able to access the system again.
// Returns an error if the request fails or the account is not found.
func (svc *AccountService) Activate(id string) error {
	logging.Trace()
	return svc.PatchRequest(&Request{
		uri:  fmt.Sprintf("/authorization/accounts/%s", id),
		body: map[string]interface{}{"inactive": false},
	}, nil)
}

// GetByName retrieves an account by username using client-side filtering.
// DEPRECATED: Business logic method - prefer using resources.AccountResource.GetByName
// When multiple accounts share the same username, an active account is
// preferred over an inactive one.
func (svc *AccountService) GetByName(name string) (*Account, error) {
	logging.Trace()

	accounts, err := svc.GetAll()
	if err != nil {
		return nil, err
	}

	var inactiveMatch *Account

	for i := range accounts {
		if accounts[i].Username == name {
			if !accounts[i].Inactive {
				return &accounts[i], nil
			}

			if inactiveMatch == nil {
				inactiveMatch = &accounts[i]
			}
		}
	}

	if inactiveMatch != nil {
		return inactiveMatch, nil
	}

	return nil, errors.New("account not found")
}
