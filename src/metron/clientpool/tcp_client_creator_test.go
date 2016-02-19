package clientpool_test

import (
	"doppler/listeners"
	"metron/clientpool"

	"crypto/tls"

	"net"

	"github.com/cloudfoundry/gosteno"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TCPClientCreator", func() {

	var (
		tlsClientCreator *clientpool.TCPClientCreator
		logger           *gosteno.Logger
		tlsClientConfig  *tls.Config
		tlsListener      net.Listener
		address          string
		conns            chan net.Conn
	)

	BeforeEach(func() {
		logger = gosteno.NewLogger("TestLogger")
		var err error
		tlsClientConfig, err = listeners.NewTLSConfig("fixtures/client.crt", "fixtures/client.key", "fixtures/loggregator-ca.crt")
		Expect(err).NotTo(HaveOccurred())
		tlsClientConfig.ServerName = "doppler"

		tlsServerConfig, err := listeners.NewTLSConfig("fixtures/server.crt", "fixtures/server.key", "fixtures/loggregator-ca.crt")
		Expect(err).NotTo(HaveOccurred())

		tlsListener, err = tls.Listen("tcp", "127.0.0.1:0", tlsServerConfig)
		Expect(err).NotTo(HaveOccurred())

		address = tlsListener.Addr().String()
		conns = acceptTLSConnections(tlsListener)

		tlsClientCreator = clientpool.NewTCPClientCreator(logger, tlsClientConfig)
	})

	Describe("CreateClient", func() {
		It("makes clients", func() {
			client, err := tlsClientCreator.CreateClient(address)
			Expect(err).ToNot(HaveOccurred())
			Expect(client.Address()).To(Equal(address))
			Expect(client.Scheme()).To(Equal("tls"))
		})

		Context("with a working TLSListener", func() {
			It("connects", func() {
				_, err := tlsClientCreator.CreateClient(address)
				Expect(err).ToNot(HaveOccurred())
				Eventually(conns).Should(Receive())
			})
		})

		Context("without a TLSListener", func() {
			BeforeEach(func() {
				Expect(tlsListener.Close()).ToNot(HaveOccurred())
			})

			It("returns a nil client with an error", func() {
				client, err := tlsClientCreator.CreateClient(address)
				Expect(err).To(HaveOccurred())
				Expect(client).To(BeNil())
				Consistently(conns).ShouldNot(Receive())
			})
		})
	})
})
