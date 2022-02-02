package file

import (
	"fmt"
	"sync"

	"github.com/spf13/cast"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/plugin"
	"github.com/cosmos/cosmos-sdk/plugin/plugins/trace/service"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Plugin name and version
const (
	// PLUGIN_NAME is the name for this streaming service plugin
	PLUGIN_NAME = "trace"

	// PLUGIN_VERSION is the version for this streaming service plugin
	PLUGIN_VERSION = "0.0.1"
)

// TOML configuration parameter keys
const (
	// KEYS_PARAM is a list of the StoreKeys we want to expose for this streaming service
	KEYS_PARAM = "keys"

	// DELIVERED_BLOCK_TIMEOUT_SECONDS is the timeout setting used in the implementation of ABCIListener.ListenSuccess()
	DELIVERED_BLOCK_TIMEOUT_SECONDS = "deliveredBlockTimeoutSeconds"
)

// Plugins is the exported symbol for loading this plugin
var Plugins = []plugin.Plugin{
	&streamingServicePlugin{},
}

type streamingServicePlugin struct {
	kss  *service.TraceStreamingService
	opts serverTypes.AppOptions
}

var _ plugin.StateStreamingPlugin = (*streamingServicePlugin)(nil)

// Name satisfies the plugin.Plugin interface
func (ssp *streamingServicePlugin) Name() string {
	return PLUGIN_NAME
}

// Version satisfies the plugin.Plugin interface
func (ssp *streamingServicePlugin) Version() string {
	return PLUGIN_VERSION
}

// Init satisfies the plugin.Plugin interface
func (ssp *streamingServicePlugin) Init(env serverTypes.AppOptions) error {
	ssp.opts = env
	return nil
}

// Register satisfies the plugin.StateStreamingPlugin interface
func (ssp *streamingServicePlugin) Register(bApp *baseapp.BaseApp, marshaller codec.BinaryCodec, keys map[string]*sdk.KVStoreKey) error {
	// load all the params required for this plugin from the provided AppOptions
	deliveredBlockTimeoutSeconds := cast.ToDuration(ssp.opts.Get(fmt.Sprintf("%s.%s.%s.%s", plugin.PLUGIN_TOML_KEY, plugin.STREAMING_TOML_KEY, PLUGIN_NAME, DELIVERED_BLOCK_TIMEOUT_SECONDS)))
	// get the store keys allowed to be exposed for this streaming service
	exposeKeyStrings := cast.ToStringSlice(ssp.opts.Get(fmt.Sprintf("%s.%s.%s.%s", plugin.PLUGIN_TOML_KEY, plugin.STREAMING_TOML_KEY, PLUGIN_NAME, KEYS_PARAM)))
	var exposeStoreKeys []sdk.StoreKey

	if len(exposeKeyStrings) > 0 {
		exposeStoreKeys = make([]sdk.StoreKey, 0, len(exposeKeyStrings))
		for _, keyStr := range exposeKeyStrings {
			if storeKey, ok := keys[keyStr]; ok {
				exposeStoreKeys = append(exposeStoreKeys, storeKey)
			}
		}
	} else { // if none are specified, we expose all the keys
		exposeStoreKeys = make([]sdk.StoreKey, 0, len(keys))
		for _, storeKey := range keys {
			exposeStoreKeys = append(exposeStoreKeys, storeKey)
		}
	}

	var err error
	ssp.kss, err = service.NewTraceStreamingService(exposeStoreKeys, marshaller, deliveredBlockTimeoutSeconds)
	if err != nil {
		return err
	}
	// register the streaming service with the BaseApp
	bApp.SetStreamingService(ssp.kss)
	return nil
}

// Start satisfies the plugin.StateStreamingPlugin interface
func (ssp *streamingServicePlugin) Start(wg *sync.WaitGroup) {
	ssp.kss.Stream(wg)
}

// Close satisfies io.Closer
func (ssp *streamingServicePlugin) Close() error {
	return ssp.kss.Close()
}
