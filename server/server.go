package common

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"

	"sync"

	"os/signal"

	"os"

	"syscall"

	"io"

	log "gitlab.ucloudadmin.com/wu/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServiceInf interface {
	SetCloseC(closeC chan struct{})
	SetShutdownWg(wg *sync.WaitGroup)
}

type Service struct {
	Name       string
	CloseC     chan struct{}
	shutdownWg *sync.WaitGroup
}

func (service *Service) Finished() {
	service.shutdownWg.Done()
}

func (service *Service) SetCloseC(closeC chan struct{}) {
	service.CloseC = closeC
}

func (service *Service) SetShutdownWg(wg *sync.WaitGroup) {
	service.shutdownWg = wg
}

type GrpcService struct {
	Service
	Server *grpc.Server
	quitC  chan interface{}
}

func NewGrpcService(name string, opts ...grpc.ServerOption) *GrpcService {
	return &GrpcService{
		Service{Name: name},
		grpc.NewServer(opts...),
		make(chan interface{}),
	}
}

func (service *GrpcService) Start(address string) error {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	log.WithField("addr", address).Info("Listen at")
	go service.waitForClose()
	if err := service.Server.Serve(l); err != nil && err != io.EOF {
		log.WithError(err).Error("service unexpect down")
		return err
	}
	return nil
}

func (service *GrpcService) waitForClose() {
	<-service.CloseC
	service.Server.GracefulStop()
	service.Finished()
}

type ServerOption struct {
	// 配置文件名(不带后缀)
	ConfigFileName string
}

func RegisterService(service ServiceInf) {
	gsvr.addService(service)
}

func WaitForEnd() {
	gsvr.waitForEnd()
}

type server struct {
	closeC     chan struct{}
	shutdownWg *sync.WaitGroup
	signalC    chan os.Signal
}

func newServer() *server {
	return &server{
		closeC:     make(chan struct{}),
		shutdownWg: &sync.WaitGroup{},
		signalC:    make(chan os.Signal, 1),
	}
}

func (s *server) close() {
	close(s.closeC)
	// wait all servie close
	s.shutdownWg.Wait()
}

func (s *server) addService(service ServiceInf) {
	service.SetCloseC(s.closeC)
	service.SetShutdownWg(s.shutdownWg)
	s.shutdownWg.Add(1)
}

func (s *server) waitForEnd() {
	signal.Notify(s.signalC, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-s.signalC
	s.close()
}

var gsvr = newServer()

func NewMutualTLSOption(localKeyFile, localCrtFile, remoteCAFile string) grpc.ServerOption {
	caCert, err := ioutil.ReadFile(remoteCAFile)
	if err != nil {
		log.WithError(err).Fatal("Fail to load client root ca certs")
	}
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		log.Fatal("Failed to append client certs")
	}

	certificate, err := tls.LoadX509KeyPair(localCrtFile, localKeyFile)

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    caCertPool,
	}

	return grpc.Creds(credentials.NewTLS(tlsConfig))
}
