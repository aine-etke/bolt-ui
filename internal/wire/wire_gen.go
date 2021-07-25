// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package wire

import (
	"github.com/boreq/velo/adapters"
	"github.com/boreq/velo/application"
	"github.com/boreq/velo/internal/config"
	"github.com/boreq/velo/internal/service"
	"github.com/boreq/velo/ports/http"
	"go.etcd.io/bbolt"
)

// Injectors from wire.go:

func BuildTransactableAdapters(tx *bbolt.Tx) (*application.TransactableAdapters, error) {
	database := adapters.NewDatabase(tx)
	transactableAdapters := &application.TransactableAdapters{
		Database: database,
	}
	return transactableAdapters, nil
}

func BuildTestTransactableAdapters(tx *bbolt.Tx, mocks Mocks) (*application.TransactableAdapters, error) {
	database := adapters.NewDatabase(tx)
	transactableAdapters := &application.TransactableAdapters{
		Database: database,
	}
	return transactableAdapters, nil
}

func BuildApplicationForTest(db *bbolt.DB) (TestApplication, error) {
	wireAdaptersProvider := newAdaptersProvider()
	transactionProvider := adapters.NewTransactionProvider(db, wireAdaptersProvider)
	browseHandler := application.NewBrowseHandler(transactionProvider)
	applicationApplication := &application.Application{
		Browse: browseHandler,
	}
	mocks := Mocks{}
	testApplication := TestApplication{
		Application: applicationApplication,
		Mocks:       mocks,
		DB:          db,
	}
	return testApplication, nil
}

func BuildService(conf *config.Config) (*service.Service, error) {
	db, err := newBolt(conf)
	if err != nil {
		return nil, err
	}
	wireAdaptersProvider := newAdaptersProvider()
	transactionProvider := adapters.NewTransactionProvider(db, wireAdaptersProvider)
	browseHandler := application.NewBrowseHandler(transactionProvider)
	applicationApplication := &application.Application{
		Browse: browseHandler,
	}
	tokenAuthProvider := http.NewTokenAuthProvider(applicationApplication)
	handler, err := http.NewHandler(applicationApplication, tokenAuthProvider)
	if err != nil {
		return nil, err
	}
	server := http.NewServer(handler)
	serviceService := service.NewService(server)
	return serviceService, nil
}

// wire.go:

type TestApplication struct {
	Application *application.Application
	Mocks
	DB *bbolt.DB
}

type Mocks struct {
}
