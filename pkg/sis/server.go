package sis

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
	"sync/atomic"
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
	ErrorID  byte   `json:"errorID,omitempty"`
	ErrorMsg string `json:"errorMsg,omitempty"`
}

func NewServer(c *Config, s Store) *Server {
	return &Server{
		config:  c,
		store:   s,
		counter: &counter{},
	}
}

func (s *Server) Start() {
	log.Println("http-server listening...")
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/", s.getRoot())
	app.Get("/metrics", s.getMetrics())
	app.Get("/subscribers/:msisdn", s.getHashes())

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

func (s *Server) getHashes() func(c *fiber.Ctx) error {
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

		//m, err := strconv.Atoi(msisdn[:len(s.config.CC)])
		m, err := strconv.ParseInt(msisdn, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: 2, ErrorMsg: "Not supported MSISDN format: " + msisdn})
		}

		info, err := s.store.Get(m)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(response{ErrorID: 1, ErrorMsg: "Not found"})
		}

		return c.Status(fiber.StatusOK).JSON(info)
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
