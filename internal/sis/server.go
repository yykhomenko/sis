package sis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	shutdownTimeout = 5 * time.Second
	readyTimeout    = time.Second
)

// Error IDs returned in the error payload.
const (
	errNotFound       byte = 1
	errMsisdnFormat   byte = 2
	errUnsupportedCC  byte = 3
	errUnsupportedNDC byte = 4
	errInvalidBody    byte = 5
	errDB             byte = 10
	errDBTimeout      byte = 11
)

type Server struct {
	config  *Config
	store   Store
	metrics *metrics
	ndcs    []string
	ndcLen  int // longest configured NDC, used in error messages
	app     *fiber.App
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
		metrics: newMetrics(),
		app:     fiber.New(fiber.Config{DisableStartupMessage: true}),
	}

	for _, ndc := range c.NDCS {
		ndcStr := strconv.Itoa(ndc)
		server.ndcs = append(server.ndcs, ndcStr)
		server.ndcLen = max(server.ndcLen, len(ndcStr))
	}

	server.app.Get("/", server.getRoot())
	server.app.Get("/ready", server.getReady())
	server.app.Get("/metrics", server.getMetrics())

	subscribers := server.app.Group("/subscribers", server.countRequests())
	subscribers.Get("/:msisdn", server.getSubscriber())
	subscribers.Put("/:msisdn", server.putSubscriber())

	return server
}

func (s *Server) Start() error {
	log.Printf("http-server listening (%s)...\n", s.config.Addr)
	return s.app.Listen(s.config.Addr)
}

func (s *Server) Stop() error {
	log.Printf("http-server stopping...\n")
	if err := s.app.ShutdownWithTimeout(shutdownTimeout); err != nil {
		log.Printf("http-server shutdown error: %v\n", err)
		return err
	}
	log.Printf("http-server stopped\n")
	return nil
}

// getRoot is the liveness probe: the process is up and serving HTTP.
func (s *Server) getRoot() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	}
}

// getReady is the readiness probe: the service can reach its store.
func (s *Server) getReady() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), readyTimeout)
		defer cancel()

		if err := s.store.Ping(ctx); err != nil {
			log.Printf("readiness check failed: %v\n", err)
			return c.SendStatus(fiber.StatusServiceUnavailable)
		}
		return c.SendStatus(fiber.StatusOK)
	}
}

func (s *Server) getMetrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "text/plain; version=0.0.4; charset=utf-8")
		return c.SendString(s.metrics.render())
	}
}

// countRequests records method and response status of every subscriber request.
func (s *Server) countRequests() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		status := c.Response().StatusCode()
		if err != nil {
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}
		s.metrics.inc(c.Method(), status)
		return err
	}
}

func (s *Server) getSubscriber() fiber.Handler {
	return func(c *fiber.Ctx) error {
		msisdn, apiErr := s.parseMsisdn(c.Params("msisdn"))
		if apiErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(apiErr)
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), s.config.DbTimeout)
		defer cancel()

		subscriber, err := s.store.Get(ctx, msisdn)
		if errors.Is(err, ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response{ErrorID: errNotFound, ErrorMsg: "Not found"})
		}
		if err != nil {
			return s.storeError(c, "get", msisdn, err)
		}

		return c.JSON(subscriber)
	}
}

func (s *Server) putSubscriber() fiber.Handler {
	return func(c *fiber.Ctx) error {
		msisdn, apiErr := s.parseMsisdn(c.Params("msisdn"))
		if apiErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(apiErr)
		}

		var req putSubscriberRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: errInvalidBody, ErrorMsg: "Invalid request body: " + err.Error()})
		}
		if err := req.validate(); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response{ErrorID: errInvalidBody, ErrorMsg: "Invalid request body: " + err.Error()})
		}

		subscriber := &Subscriber{
			Msisdn:       msisdn,
			BillingType:  *req.BillingType,
			LanguageType: *req.LanguageType,
			OperatorType: *req.OperatorType,
		}

		ctx, cancel := context.WithTimeout(c.UserContext(), s.config.DbTimeout)
		defer cancel()

		if err := s.store.Set(ctx, subscriber); err != nil {
			return s.storeError(c, "set", msisdn, err)
		}

		return c.SendStatus(fiber.StatusOK)
	}
}

// storeError logs a store failure and maps it to a 504 (timeout) or 500 response.
func (s *Server) storeError(c *fiber.Ctx, op string, msisdn int64, err error) error {
	log.Printf("%s subscriber %d: %v\n", op, msisdn, err)
	if errors.Is(err, context.DeadlineExceeded) {
		return c.Status(fiber.StatusGatewayTimeout).JSON(response{ErrorID: errDBTimeout, ErrorMsg: "DB timeout"})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(response{ErrorID: errDB, ErrorMsg: "InternalServerError DB"})
}

// parseMsisdn checks length, digits, CC and NDC of an MSISDN and converts it to a number.
func (s *Server) parseMsisdn(msisdn string) (int64, *response) {
	if len(msisdn) != s.config.MsisdnLength || !isDigits(msisdn) {
		return 0, &response{ErrorID: errMsisdnFormat, ErrorMsg: "Not supported MSISDN format: " + msisdn}
	}

	cc := s.config.CC
	if !strings.HasPrefix(msisdn, cc) {
		return 0, &response{ErrorID: errUnsupportedCC, ErrorMsg: "Not supported CC: " + msisdn[:min(len(cc), len(msisdn))]}
	}

	rest := msisdn[len(cc):]
	supported := false
	for _, ndc := range s.ndcs {
		if strings.HasPrefix(rest, ndc) {
			supported = true
			break
		}
	}
	if !supported {
		return 0, &response{ErrorID: errUnsupportedNDC, ErrorMsg: "Not supported NDC: " + rest[:min(s.ndcLen, len(rest))]}
	}

	n, err := strconv.ParseInt(msisdn, 10, 64)
	if err != nil {
		return 0, &response{ErrorID: errMsisdnFormat, ErrorMsg: "Not supported MSISDN format: " + msisdn}
	}
	return n, nil
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
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
