package http_test

import (
	"net/http"

	commonhttp "code.cloudfoundry.org/gorouter/common/http"
	"code.cloudfoundry.org/gorouter/common/http/fakes"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

const uuid_regex = `^[[:xdigit:]]{8}(-[[:xdigit:]]{4}){3}-[[:xdigit:]]{12}$`

// 64-bit random hexadecimal string
const b3_id_regex = `^[[:xdigit:]]{16}$`

var _ = Describe("Headers", func() {
	Describe("SetVcapRequestIdHeader", func() {
		var (
			logger lager.Logger
			req    *http.Request
		)
		BeforeEach(func() {
			logger = lagertest.NewTestLogger("headers-test")
			var err error
			req, err = http.NewRequest("GET", "test.endpoint", nil)
			Expect(err).ToNot(HaveOccurred())
		})
		JustBeforeEach(func() {
			commonhttp.SetVcapRequestIdHeader(req, logger)
		})

		Context("when X-Vcap-Request-Id is not set", func() {
			It("sets the X-Vcap-Request-Id header", func() {
				reqID := req.Header.Get(commonhttp.VcapRequestIdHeader)
				Expect(reqID).ToNot(BeEmpty())
				Expect(reqID).To(MatchRegexp(uuid_regex))
			})

			It("logs the header", func() {
				reqID := req.Header.Get(commonhttp.VcapRequestIdHeader)
				Expect(logger).To(gbytes.Say("vcap-request-id-header-set"))
				Expect(logger).To(gbytes.Say(reqID))
			})

		})

		Context("when X-Vcap-Request-Id is set", func() {
			BeforeEach(func() {
				req.Header.Set(commonhttp.VcapRequestIdHeader, "BOGUS-HEADER")
			})

			It("overwrites the X-Vcap-Request-Id header", func() {
				reqID := req.Header.Get(commonhttp.VcapRequestIdHeader)
				Expect(reqID).ToNot(BeEmpty())
				Expect(reqID).ToNot(Equal("BOGUS-HEADER"))
				Expect(reqID).To(MatchRegexp(uuid_regex))
			})

			It("logs the header", func() {
				reqID := req.Header.Get(commonhttp.VcapRequestIdHeader)
				Expect(logger).To(gbytes.Say("vcap-request-id-header-set"))
				Expect(logger).To(gbytes.Say(reqID))
			})
		})
	})

	Describe("SetTraceHeaders", func() {
		var respWriter http.ResponseWriter

		BeforeEach(func() {
			respWriter = httpfakes.NewFakeResponseWriter()
		})

		JustBeforeEach(func() {
			commonhttp.SetTraceHeaders(respWriter, "1.1.1.1", "example.com")
		})

		It("sets the trace headers on the response", func() {
			Expect(respWriter.Header().Get(commonhttp.VcapRouterHeader)).To(Equal("1.1.1.1"))
			Expect(respWriter.Header().Get(commonhttp.VcapBackendHeader)).To(Equal("example.com"))
			Expect(respWriter.Header().Get(commonhttp.CfRouteEndpointHeader)).To(Equal("example.com"))
		})
	})

	Describe("SetB3TraceHeader", func() {
		var (
			logger lager.Logger
			req    *http.Request
		)
		BeforeEach(func() {
			var err error
			req, err = http.NewRequest("GET", "test.endpoint", nil)
			Expect(err).ToNot(HaveOccurred())
		})
		JustBeforeEach(func() {
			commonhttp.SetB3TraceIdHeader(req, logger)
		})

		Context("when logger is set", func() {
			BeforeEach(func() {
				logger = lagertest.NewTestLogger("headers-test")
			})
			Context("when X-B3-TraceId is not set", func() {
				It("generates a new b3 id and sets the X-B3-TraceId header with a random 64-bit hexadecimal string", func() {
					reqID := req.Header.Get(commonhttp.B3TraceIdHeader)
					Expect(reqID).ToNot(BeEmpty())
					Expect(reqID).To(MatchRegexp(b3_id_regex))
				})

				It("logs the header", func() {
					reqID := req.Header.Get(commonhttp.B3TraceIdHeader)
					Expect(logger).To(gbytes.Say("b3-trace-id-header-set"))
					Expect(logger).To(gbytes.Say(reqID))
				})

			})

			Context("when X-B3-TraceId is set", func() {
				BeforeEach(func() {
					req.Header.Set(commonhttp.B3TraceIdHeader, "BOGUS-HEADER")
				})

				It("should not override the X-B3-TraceId header", func() {
					Expect(req.Header.Get(commonhttp.B3TraceIdHeader)).To(Equal("BOGUS-HEADER"))
				})

				It("logs the header", func() {
					Expect(logger).To(gbytes.Say("b3-trace-id-header-exists"))
					Expect(logger).To(gbytes.Say("BOGUS-HEADER"))
				})
			})
		})

		Context("when logger is nil", func() {
			It("does not fail when X-B3-TraceId is not set", func() {
				reqID := req.Header.Get(commonhttp.B3TraceIdHeader)
				Expect(reqID).ToNot(BeEmpty())
				Expect(reqID).To(MatchRegexp(b3_id_regex))
			})
			Context("when X-B3-TraceId is set", func() {
				BeforeEach(func() {
					req.Header.Set(commonhttp.B3TraceIdHeader, "BOGUS-HEADER")
				})
				It("does not fail when X-B3-TraceId is set", func() {
					Expect(req.Header.Get(commonhttp.B3TraceIdHeader)).To(Equal("BOGUS-HEADER"))
				})
			})
		})
	})

	Describe("SetB3SpanIdHeader", func() {
		var (
			logger lager.Logger
			req    *http.Request
		)
		BeforeEach(func() {
			var err error
			req, err = http.NewRequest("GET", "test.endpoint", nil)
			Expect(err).ToNot(HaveOccurred())
		})
		JustBeforeEach(func() {
			commonhttp.SetB3SpanIdHeader(req, logger)
		})

		Context("when logger is set", func() {
			BeforeEach(func() {
				logger = lagertest.NewTestLogger("headers-test")
			})
			Context("when tracing is enabled", func() {
				It("generates a new b3 span id", func() {
					reqID := req.Header.Get(commonhttp.B3SpanIdHeader)
					Expect(reqID).ToNot(BeEmpty())
					Expect(reqID).To(MatchRegexp(b3_id_regex))
				})

				It("logs the header", func() {
					reqID := req.Header.Get(commonhttp.B3SpanIdHeader)
					Expect(logger).To(gbytes.Say("b3-span-id-header-set"))
					Expect(logger).To(gbytes.Say(reqID))
				})

			})
		})

		Context("when logger is nil", func() {
			It("does not fail when X-B3-Span is not set", func() {
				reqID := req.Header.Get(commonhttp.B3SpanIdHeader)
				Expect(reqID).ToNot(BeEmpty())
				Expect(reqID).To(MatchRegexp(b3_id_regex))
			})
			Context("when X-B3-SpanId header is set", func() {
				BeforeEach(func() {
					req.Header.Set(commonhttp.B3TraceIdHeader, "sample-span-id")
				})
				It("sets a new X-B3-SpanId", func() {
					Expect(req.Header.Get(commonhttp.B3SpanIdHeader)).ToNot(Equal("sample-span-id"))
				})
			})
		})
	})

	Describe("ValidateCfAppInstance", func() {
		var (
			appInstanceHeader string
		)

		Context("when given a complete app instance header", func() {
			BeforeEach(func() {
				appInstanceHeader = "app-id:1"
			})

			It("returns the app id and app index", func() {
				appID, appIndex, err := commonhttp.ValidateCfAppInstance(appInstanceHeader)
				Expect(err).ToNot(HaveOccurred())
				Expect(appID).To(Equal("app-id"))
				Expect(appIndex).To(Equal("1"))
			})
		})

		Context("when given an incomplete app instance header", func() {
			BeforeEach(func() {
				appInstanceHeader = "app-id:"
			})

			It("returns an error", func() {
				_, _, err := commonhttp.ValidateCfAppInstance(appInstanceHeader)
				Expect(err).To(HaveOccurred())
			})
		})
		Context("when only the app id is given", func() {
			BeforeEach(func() {
				appInstanceHeader = "app-id"
			})

			It("returns an error", func() {
				_, _, err := commonhttp.ValidateCfAppInstance(appInstanceHeader)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
