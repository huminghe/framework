/*
Copyright 2016 Medcl (m AT medcl.net)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ui

import (
	log "github.com/cihub/seelog"
	uis "github.com/huminghe/framework/core/api"
	. "github.com/huminghe/framework/core/config"
	"github.com/huminghe/framework/core/logger"
	"github.com/huminghe/framework/core/ui"
	"github.com/huminghe/framework/core/ui/websocket"
	"github.com/huminghe/framework/core/vfs"
	"github.com/huminghe/framework/modules/ui/admin"
	"github.com/huminghe/framework/modules/ui/public"
	"github.com/huminghe/framework/static"
	_ "net/http/pprof"
)

type UIModule struct {
}

func LoggerReceiver(message string, level log.LogLevel, context log.LogContextInterface) {

	websocket.BroadcastMessage(message)
}

func (module UIModule) Name() string {
	return "Web"
}
func (module UIModule) Setup(cfg *Config) {

	uiConfig := ui.UIConfig{}
	cfg.Unpack(&uiConfig)

	if uiConfig.Enabled {

		uis.EnableAuth(uiConfig.AuthConfig.Enabled)

		//init admin ui
		admin.InitUI()

		//init public ui
		public.InitUI(uiConfig.AuthConfig)

		//register websocket logger
		logger.RegisterWebsocketHandler(LoggerReceiver)

		ui.StartUI(&uiConfig)

		vfs.RegisterFS(static.StaticFS{StaticFolder: "static", TrimLeftPath: "", CheckLocalFirst: true})
	}

}

func (module UIModule) Start() error {

	return nil
}
func (module UIModule) Stop() error {

	return nil
}
