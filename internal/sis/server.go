package sis

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	config  *Config
	store   Store
	counter *counter
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
	return &Server{
		config:  c,
		store:   s,
		counter: &counter{},
	}
}

func (s *Server) Start() {
	log.Printf("http-server listening (%s)...\n", s.config.Addr)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/", s.getRoot())
	app.Get("/metrics", s.getMetrics())
	app.Get("/subscribers/:msisdn", s.getSubscriber())
	app.Put("/subscribers/:msisdn", s.putSubscriber())

	log.Fatal(app.Listen(s.config.Addr))
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
		case errors.Is(err, sql.ErrNoRows):
			return c.Status(fiber.StatusNotFound).JSON(response{ErrorID: 1, ErrorMsg: "Not found"})
		case err != nil:
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

		req := struct {
			BillingType  int16 `json:"billing_type"`
			LanguageType int16 `json:"language_type"`
			OperatorType int16 `json:"operator_type"`
		}{}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 2, ErrorMsg: "Invalid request body, err: " + err.Error()})
		}

		subscriber := &Subscriber{
			Msisdn:       m,
			BillingType:  req.BillingType,
			LanguageType: req.LanguageType,
			OperatorType: req.OperatorType,
		}

		if err := s.store.Set(c.Context(), subscriber); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response{ErrorID: 10, ErrorMsg: "InternalServerError DB"})
		}

		return c.SendStatus(fiber.StatusOK)
	}
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
