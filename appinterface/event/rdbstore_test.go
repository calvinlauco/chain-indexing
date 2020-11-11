package event_test

import (
	"github.com/crypto-com/chainindex/entity/event/test"
	. "github.com/crypto-com/chainindex/test"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appinterface_event "github.com/crypto-com/chainindex/appinterface/event"
	"github.com/crypto-com/chainindex/entity/event"
	"github.com/crypto-com/chainindex/infrastructure/pg"
	"github.com/crypto-com/chainindex/internal/primptr"
)

var _ = Describe("RdbEventStore", func() {
	WithTestPgxConn(func(pgxConn *pg.PgxConn, pgMigrate *pg.Migrate) {
		BeforeEach(func() {
			_ = pgMigrate.Reset()
			pgMigrate.MustUp()
		})

		AfterEach(func() {
			_ = pgMigrate.Reset()
		})

		Describe("Insert", func() {
			It("should insert an event properly without any error", func() {
				store := appinterface_event.NewRDbStore(pgxConn.ToHandle())

				event := test.NewFakeEvent()
				err := store.Insert(event)
				Expect(err).To(BeNil())
			})
		})

		Describe("InsertAll", func() {
			It("should insert multiple events properly without any error", func() {
				store := appinterface_event.NewRDbStore(pgxConn.ToHandle())

				events := []event.Event{test.NewFakeEvent(), test.NewFakeEvent()}
				err := store.InsertAll(events)
				Expect(err).To(BeNil())
			})
		})

		Describe("GetLatestHeight", func() {
			It("should return nil when events table does not have any record", func() {
				store := appinterface_event.NewRDbStore(pgxConn.ToHandle())

				actual, err := store.GetLatestHeight()
				Expect(err).To(BeNil())
				Expect(actual).To(Equal(primptr.Int64Nil()))
			})

			It("should get 1 after insert fake event with height 1", func() {
				store := appinterface_event.NewRDbStore(pgxConn.ToHandle())

				// Insert an event with latestHeight 1
				mockEvent := test.NewMockEvent()
				mockEvent.On("Height").Return(int64(1))
				mockEvent.On("Name").Return("MockEvent")
				mockEvent.On("Version").Return(0)
				mockEvent.On("Id").Return("mock-event-id")
				mockEvent.On("ToJSON").Return("\"MockEvent\"")
				err := store.Insert(mockEvent)
				Expect(err).To(BeNil())

				latestHeight, err := store.GetLatestHeight()
				Expect(err).To(BeNil())
				Expect(latestHeight).To(Equal(primptr.Int64(1)))
			})
		})
	})
})
