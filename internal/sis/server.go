package sis

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	config  *Config
	store   Store
	counter *counter
	app     *fiber.App
}

type counter struct {
	subscribers uint64
}

type response struct {
	Value    string `json:"value,omitempty"`
	ErrorID  byte   `json:"error_id,omitempty"`
	ErrorMsg string `json:"error_msg,omitempty"`
}

func NewServer(c *Config, s Store) *Server {
	server := &Server{
		config:  c,
		store:   s,
		counter: &counter{},
		app:     fiber.New(fiber.Config{DisableStartupMessage: true}),
	}

	server.app.Get("/", server.getRoot())
	server.app.Get("/metrics", server.getMetrics())
	server.app.Get("/subscribers/:msisdn", server.getSubscriber())
	server.app.Put("/subscribers/:msisdn", server.putSubscriber())

	return server
}

func (s *Server) Start() error {
	log.Printf("http-server listening (%s)...\n", s.config.Addr)
	return s.app.Listen(s.config.Addr)
}

func (s *Server) Stop() error {
	log.Printf("http-server stopping...\n")
	if err := s.app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("http-server shutdown error: %v\n", err)
		return err
	}
	log.Printf("http-server stopped\n")
	return nil
}

func (s *Server) getRoot() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *Server) getMetrics() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).SendString(fmt.Sprintf(
			"subscribers_total %d\n",
			s.counter.subscribers,
		))
	}
}

func (s *Server) getSubscriber() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		atomic.AddUint64(&s.counter.subscribers, 1)
		msisdn := c.Params("msisdn")

		if !validateMsisdnLen(msisdn, s.config.MsisdnLength) {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 2, ErrorMsg: "Not supported MSISDN format: " + msisdn})
		}

		if cc, ok := validateCC(msisdn, s.config.CC); !ok {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 3, ErrorMsg: "Not supported CC: " + cc})
		}

		if ndc, ok := validateNDC(msisdn, s.config.NDCS); !ok {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 4, ErrorMsg: "Not supported NDC: " + ndc})
		}

		m, err := strconv.ParseInt(msisdn, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 2, ErrorMsg: "Not supported MSISDN format: " + msisdn})
		}

		subscriber, err := s.store.Get(c.Context(), m)
		switch {
		case errors.Is(err, ErrNotFound):
			return c.Status(fiber.StatusNotFound).JSON(response{ErrorID: 1, ErrorMsg: "Not found"})
		case err != nil:
			log.Printf("get subscriber %d: %v\n", m, err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{ErrorID: 10, ErrorMsg: "InternalServerError DB"})
		}

		return c.Status(fiber.StatusOK).JSON(subscriber)
	}
}

func (s *Server) putSubscriber() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		msisdn := c.Params("msisdn")

		if !validateMsisdnLen(msisdn, s.config.MsisdnLength) {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 2, ErrorMsg: "Not supported MSISDN format: " + msisdn})
		}

		if cc, ok := validateCC(msisdn, s.config.CC); !ok {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 3, ErrorMsg: "Not supported CC: " + cc})
		}

		if ndc, ok := validateNDC(msisdn, s.config.NDCS); !ok {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 4, ErrorMsg: "Not supported NDC: " + ndc})
		}

		m, err := strconv.ParseInt(msisdn, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 2, ErrorMsg: "Not supported MSISDN format: " + msisdn})
		}

		var req putSubscriberRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 5, ErrorMsg: "Invalid request body: " + err.Error()})
		}
		if err := req.validate(); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 5, ErrorMsg: "Invalid request body: " + err.Error()})
		}

		subscriber := &Subscriber{
			Msisdn:       m,
			BillingType:  *req.BillingType,
			LanguageType: *req.LanguageType,
			OperatorType: *req.OperatorType,
		}

		if err := s.store.Set(c.Context(), subscriber); err != nil {
			log.Printf("set subscriber %d: %v\n", m, err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{ErrorID: 10, ErrorMsg: "InternalServerError DB"})
		}

		return c.SendStatus(fiber.StatusOK)
	}
}

// putSubscriberRequest uses pointers so a missing field can be told apart
// from an explicit 0: PUT replaces the whole profile, so every field is required.
type putSubscriberRequest struct {
	BillingType  *int16 `json:"billing_type"`
	LanguageType *int16 `json:"language_type"`
	OperatorType *int16 `json:"operator_type"`
}

func (r putSubscriberRequest) validate() error {
	fields := []struct {
		name  string
		value *int16
	}{
		{"billing_type", r.BillingType},
		{"language_type", r.LanguageType},
		{"operator_type", r.OperatorType},
	}
	for _, f := range fields {
		if f.value == nil {
			return fmt.Errorf("%s is required", f.name)
		}
		if *f.value < 0 {
			return fmt.Errorf("%s must not be negative", f.name)
		}
	}
	return nil
}

func validateMsisdnLen(msisdn string, length int) bool {
	return len(msisdn) == length
}

func validateCC(msisdn, confCC string) (string, bool) {
	cc := msisdn[:3]
	if cc != confCC {
		return cc, false
	}
	return cc, true
}

func validateNDC(msisdn string, ndcs []int) (string, bool) {
	ndcStr := msisdn[3:5]

	ndc, err := strconv.Atoi(ndcStr)
	if err != nil {
		log.Println(err)
	}

	for _, n := range ndcs {
		if ndc == n {
			return ndcStr, true
		}
	}

	return ndcStr, false
}
