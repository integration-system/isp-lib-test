package mock

import (
	"context"
	"encoding/json"
	"github.com/googollee/go-socket.io"
	"github.com/integration-system/isp-lib/structure"
	"github.com/integration-system/isp-lib/utils"
	"net"
	"net/http"
	"sync"
)

type ConfigServiceOption func(cs *mockConfigService)

type mockConfigService struct {
	server            *socketio.Server
	httpServer        *http.Server
	remoteConfigs     map[string][]byte
	discoveredModules map[string][]byte

	readyModules map[string]bool
	lock         sync.Mutex
}

func (cs *mockConfigService) Shutdown() error {
	if cs.httpServer != nil {
		return cs.httpServer.Shutdown(context.Background())
	}
	return nil
}

func (cs *mockConfigService) AsyncServe(address string) error {
	mux := http.NewServeMux()
	mux.Handle("/socket.io/", cs.server)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	cs.httpServer = &http.Server{Handler: mux, Addr: address}
	go func() {
		_ = cs.httpServer.Serve(listener)
	}()

	return nil
}

func (cs *mockConfigService) ModuleIsReady(module string) bool {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	return cs.readyModules[module]
}

func (cs *mockConfigService) onConnect() {

}

func (cs *mockConfigService) onDisconnect() {

}

func (cs *mockConfigService) onError() {

}

func (cs *mockConfigService) onReceivedModuleRequirements() {

}

func (cs *mockConfigService) onReceiveModuleReady() {

}

func (cs *mockConfigService) onReceiveRemoteConfigSchema() {

}

func (cs *mockConfigService) onReceiveRoutesUpdate() {

}

func NewMockConfigService(opts ...ConfigServiceOption) (*mockConfigService, error) {
	srv, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	cs := &mockConfigService{
		server: srv,
	}
	for _, v := range opts {
		v(cs)
	}

	_ = srv.On("connection", cs.onConnect)
	_ = srv.On("disconnection", cs.onDisconnect)
	_ = srv.On("error", cs.onError)
	_ = srv.On(utils.ModuleSendRequirements, cs.onReceivedModuleRequirements)
	_ = srv.On(utils.ModuleReady, cs.onReceiveModuleReady)
	_ = srv.On(utils.ModuleSendConfigSchema, cs.onReceiveRemoteConfigSchema)
	_ = srv.On(utils.ModuleUpdateRoutes, cs.onReceiveRoutesUpdate)

	return cs, nil
}

func WithDiscoveredModules(modules map[string][]structure.AddressConfiguration) ConfigServiceOption {
	marshaled := make(map[string][]byte, len(modules))
	for k, v := range modules {
		bytes, _ := json.Marshal(v)
		marshaled[k] = bytes
	}

	return func(cs *mockConfigService) {
		cs.discoveredModules = marshaled
	}
}

func WithRemoteConfiguration(configs map[string]interface{}) ConfigServiceOption {
	marshaled := make(map[string][]byte, len(configs))
	for k, v := range configs {
		bytes, _ := json.Marshal(v)
		marshaled[k] = bytes
	}

	return func(cs *mockConfigService) {
		cs.remoteConfigs = marshaled
	}
}
