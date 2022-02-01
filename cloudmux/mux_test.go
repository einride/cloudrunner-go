package cloudmux

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"go.einride.tech/cloudrunner/cloudzap"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"gotest.tools/v3/assert"
)

func TestServe_Canceled(t *testing.T) {
	t.Parallel()
	fx := newTestFixture(t)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fx.listen()
		wg.Done()
	}()

	// wait for server to be ready
	time.Sleep(time.Millisecond * 20)

	// stop listening
	fx.stop()

	wg.Wait()
	assert.NilError(t, fx.lisErr)
}

func TestServe_GracefulGRPC(t *testing.T) {
	t.Parallel()
	fx := newTestFixture(t)
	fx.grpc.latency = time.Second
	requestConn := make(chan struct{})
	fx.grpc.requestRecvChan = requestConn

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fx.listen()
		wg.Done()
	}()

	client := greeterClient(t, fx.lis.Addr())
	var callErr error
	wg.Add(1)
	go func() {
		_, callErr = client.SayHello(context.Background(), &helloworld.HelloRequest{
			Name: "world",
		})
		wg.Done()
	}()

	// wait for server to have received request
	<-requestConn

	// stop listening
	fx.stop()

	wg.Wait()
	assert.NilError(t, callErr)
	assert.NilError(t, fx.lisErr)
}

func TestServe_GracefulHTTP(t *testing.T) {
	t.Parallel()
	fx := newTestFixture(t)
	fx.http.latency = time.Second
	requestConn := make(chan struct{})
	fx.http.requestRecvChan = requestConn

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		fx.listen()
		wg.Done()
	}()

	// request needs to have a timeout in order to be blocking
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fx.url(), nil)
	assert.NilError(t, err)

	var callErr error
	wg.Add(1)
	go func() {
		res, err := http.DefaultClient.Do(req)
		callErr = err
		if err == nil {
			_ = res.Body.Close()
		}
		wg.Done()
	}()

	// wait for server to have received request
	<-requestConn

	// stop listening
	fx.stop()

	wg.Wait()
	assert.NilError(t, callErr)
	assert.NilError(t, fx.lisErr)
}

func newTestFixture(t *testing.T) *testFixture {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	ctx = cloudzap.WithLogger(ctx, zaptest.NewLogger(t))
	var lc net.ListenConfig
	lis, err := lc.Listen(ctx, "tcp", ":0")
	assert.NilError(t, err)

	grpcS := grpc.NewServer()
	grpcH := &grpcServer{}
	helloworld.RegisterGreeterServer(grpcS, grpcH)
	httpH := &httpServer{}
	httpS := &http.Server{Handler: httpH}

	return &testFixture{
		ctx:   ctx,
		stop:  cancel,
		lis:   lis,
		grpcS: grpcS,
		grpc:  grpcH,
		httpS: httpS,
		http:  httpH,
	}
}

type testFixture struct {
	ctx    context.Context
	stop   func()
	lis    net.Listener
	grpcS  *grpc.Server
	grpc   *grpcServer
	httpS  *http.Server
	http   *httpServer
	lisErr error
}

func (fx *testFixture) url() string {
	return fmt.Sprintf("http://localhost:%d", fx.lis.Addr().(*net.TCPAddr).Port)
}

func (fx *testFixture) listen() {
	if err := ServeGRPCHTTP(fx.ctx, fx.lis, fx.grpcS, fx.httpS); err != nil {
		fx.lisErr = err
	}
}

func greeterClient(t *testing.T, addr net.Addr) helloworld.GreeterClient {
	t.Helper()
	conn, err := grpc.Dial(
		addr.String(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	assert.NilError(t, err)
	return helloworld.NewGreeterClient(conn)
}

var _ helloworld.GreeterServer = &grpcServer{}

type grpcServer struct {
	latency time.Duration
	helloworld.UnimplementedGreeterServer
	requestRecvChan chan<- struct{}
}

func (g *grpcServer) SayHello(
	_ context.Context,
	request *helloworld.HelloRequest,
) (*helloworld.HelloReply, error) {
	if g.requestRecvChan != nil {
		g.requestRecvChan <- struct{}{}
	}
	if g.latency != 0 {
		time.Sleep(g.latency)
	}
	return &helloworld.HelloReply{Message: "Hello " + request.GetName()}, nil
}

type httpServer struct {
	requestRecvChan chan<- struct{}
	latency         time.Duration
}

func (h *httpServer) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if h.requestRecvChan != nil {
		h.requestRecvChan <- struct{}{}
	}
	if h.latency != 0 {
		time.Sleep(h.latency)
	}
	fmt.Fprintf(w, "Hello")
}
